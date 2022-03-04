import uuid
import time
import requests
import json
import logging

from flask import jsonify, request, render_template, Blueprint
from models import *
from rabbit import MqMessage
from datetime import datetime
from decorators import authenticated, is_admin
from croniter import croniter

from Triggers.Task.BackEndClient.scheduler import Scheduler
from Triggers.TriggerValidation import GenericTriggerValidation
from Triggers.TriggerValidation import TriggerLogicChecker

from Triggers.Query import listOfTriggers
from Triggers.Query import queryTriggers

from Triggers.Task.BackEndClient.TaskBackEndClient import TaskBackEndClient


triggers = Blueprint('triggers', __name__)

########## Get All Triggers ##################################################

@triggers.route('/api/triggers', methods= ['GET'])
@authenticated
def getTriggers():
  args_hash = request.args
  query_flag = False

  for arg in args_hash:
    query_flag = True

  if query_flag is False:
    return listOfTriggers()
  else:
    return jsonify({ "triggers": queryTriggers(args_hash) })


########## Get One Trigger ###################################################

@triggers.route('/api/triggers/<triggerUUID>', methods= ['GET'])
@authenticated
def getTrigger(triggerUUID):

  # Check for Invalid/Non-existand Trigger IDs
  try:
    myTrigger = GoFlowTriggers.objects(triggerId = triggerUUID)[0]

  except ValueError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Invalid Trigger ID Value"}), 400

  except IndexError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Trigger {} Not Found".format(triggerUUID)}), 404

  else:
    return jsonify(myTrigger.getInfo())

##############################################################################
##############################################################################
########## New Trigger #######################################################

@triggers.route('/api/triggers', methods= ['POST'])
@authenticated
def createTrigger():
  req_data = request.get_json()

  logging.info("Create Trigger Called - {}".format(req_data))

  if GenericTriggerValidation.checkForRequiredTopLevelFields(req_data)['check_flag']:
    logging.info("Passed Min Required Trigger Fields")
    pass
  else:
    logging.info("Falied Min Required Trigger Fields")
    return jsonify(
        GenericTriggerValidation.checkForRequiredTopLevelFields(req_data)['check_message']
      ), 400

  # At this time Trigger IDs are technically optional ?
  if "triggerIds" in req_data.keys():
    myTriggerIds = req_data["triggerIds"]
  else:
    myTriggerIds = None

  # This is our Trigger Object, still have to run some addtional validation prior to writing it to db
  myNewTrigger = GoFlowTriggers(
    triggerId = str(uuid.uuid4()),
    triggerName = req_data["name"],
    triggerActive = req_data["active"],
    flowIds = req_data["flowIds"],
    flowInputs =  GoFlowHelper.jsonArrayToRunTimeInputs(req_data["inputs"]),
    triggerIds = myTriggerIds,
    triggerLogic = req_data["triggerLogic"],
    createdOn = datetime.now(),
    updatedOn = datetime.now()
  )

  # Still Doing Validation But on the Object its self
  v = GenericTriggerValidation(myNewTrigger)
  validation_flag = False # Always assume the worst
  
  #Checking for Valid FlowIDs included
  if len(myNewTrigger.flowIds) == 0:
    return jsonify(
      {"Error": "Validation Failure", "Message": "Empty flowId array, must sepecify at least one valid flow referance"}
    ), 400

  for flowId in myNewTrigger.flowIds:      
    r = v.checkValidFlowId(flowId)
    if r["check_flag"]:
      validation_flag = True
    else:
      return jsonify({
          "Error" : "Validation Failed",
          "Failure": r['check_message']
        }), 404
  
  # Checking TriggerLogic
  tCheck = TriggerLogicChecker(myNewTrigger.triggerLogic)

  if tCheck.sanityCheck():
    logging.info(tCheck.sanity_message)
  else:
    return jsonify(
      {
        "Error": "You did not pass the trigger logic sanity test",
        "Message": tCheck.sanity_message
      }
    ), 400

  ###
  ## This is where we decided how we are going to go, Task or Event
  if tCheck.triggerType == 'Task':
    logging.info("Workingo on Task submission")
    backend_client =  TaskBackEndClient(myNewTrigger)
    
    if validation_flag:
      backend_client.submitNewTaskTrigger()
      myNewTrigger.save()
  else:
    logging.info("Working on Event submission")
    myNewTrigger.save()

  logging.info("New Trigger: ", myNewTrigger.getInfo())

  ### MqMessage Object - Handles Sending event to EventBridge
  #
  msg = MqMessage()

  message = jsonify({"type": "createTrigger",
                  "class": "GoFlowTrigger",
                  "timestamp": int(time.time()),
                  "payload": {"trigger": myNewTrigger.getInfo()}
                  }).get_data(as_text=True)

  if not msg.mqsend(message):
      logging.info("!!! ERROR: Sending Message:" + str(myNewTrigger.getInfo()) )

  return jsonify({"trigger": myNewTrigger.getInfo()})

##############################################################################
##############################################################################
########## Modify a Trigger ##################################################

@triggers.route('/api/triggers/<triggerUUID>', methods= ['PUT','PATCH'])
@authenticated
def updateTrigger(triggerUUID):

  req_data = request.get_json()

  logging.info("Update Trigger - {}".format(req_data))

  # Check for Invalid/Non-existand Trigger IDs
  try:
    myTrigger = GoFlowTriggers.objects(triggerId = triggerUUID)[0]

  except ValueError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Invalid Trigger ID Value"}), 400

  except IndexError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Trigger {} Not Found".format(triggerUUID)}), 404


  # Now go though the inputs and figure out what was passed
  myPendingUpdate = {}
  for field in GoFlowTriggers.getUpdateFields():
    if field in req_data:
      print("Need to add:", 
            GoFlowTriggers.getUpdateFields()[field], "=", req_data[field])

      # Conversion of JSON Bool to Mongo Bool for triggerActive
      if GoFlowTriggers.getUpdateFields()[field] == "triggerActive":

        # If String is passed instead of Bool
        if type(req_data[field]) == str:
          if req_data[field].lower() == "true":
            req_data[field] = True
          else:
            req_data[field] = False

      elif GoFlowTriggers.getUpdateFields()[field] == "flowInputs":

        req_data[field] = GoFlowHelper.jsonArrayToRunTimeInputs(req_data["inputs"])

      myPendingUpdate[GoFlowTriggers.getUpdateFields()[field]] = req_data[field]

  # itt though the updates and then save
  for update in myPendingUpdate:
    print("Trying to update: ", update, " to ", myPendingUpdate[update])
    myTrigger[update] = myPendingUpdate[update]

  myTrigger["updatedOn"] = datetime.now()

  # TRIGGER SAVE!

    # Still Doing Validation But on the Object its self
  v = GenericTriggerValidation(myTrigger)
  validation_flag = False # Always assume the worst
  
  #Checking for And Empty Flow Array
  if len(myTrigger.flowIds) == 0:
    return jsonify(
      {"Error": "Validation Failure", "Message": "Empty flowId array, must sepecify at least one valid flow referance"}
    ), 400

  # Checking for Valid FlowIds
  for flowId in myTrigger.flowIds:      
    r = v.checkValidFlowId(flowId)
    if r["check_flag"]:
      validation_flag = True
    else:
      return jsonify({
          "Error" : "Validation Failed",
          "Failure": r['check_message']
        }), 404

  # Checking TriggerLogic
  tCheck = TriggerLogicChecker(myTrigger.triggerLogic)

  if tCheck.sanityCheck():
    logging.info(tCheck.sanity_message)
  else:
    return jsonify(
      {
        "Error": "You did not pass the trigger logic sanity test",
        "Message": tCheck.sanity_message
      }
    ), 400


  if tCheck.triggerType == 'Task':
    logging.info("Workingo on Task submission")
      
    if validation_flag:
      backend_client =  TaskBackEndClient(myTrigger)
      backend_client.updateTaskTrigger()
      myTrigger.save()

  else:
    logging.info("Working on Event submission")  
    myTrigger.save()

  ### MqMessage Object - Handles sending delete event to EventBridge
  #
  msg = MqMessage()


  message = jsonify({"type": "updateTrigger",
                    "class": "GoFlowTrigger",
                    "timestamp": int(time.time()),
                    "payload": {"trigger": myTrigger.getInfo()}
                    }).get_data(as_text=True)

  if not msg.mqsend(message):
    print("!!! ERROR: Sending Message:", myTrigger.getInfo())
  #
  ###

  return jsonify(myTrigger.getInfo())


########## Delete Trigger ####################################################

@triggers.route('/api/triggers/<triggerUUID>', methods= ['DELETE'])
@authenticated
def deleteTrigger(triggerUUID):

  req_data = request.get_json()

  # Check for Invalid/Non-existand Trigger IDs
  try:
    myTrigger = GoFlowTriggers.objects(triggerId = triggerUUID)[0]

  except ValueError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Invalid Trigger ID Value"}), 400

  except IndexError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Trigger {} Not Found".format(triggerUUID)}), 404

  else:
    myTrigger.delete()

### Secheduler Object - Make the backend call to the scheduler to remove job 
#
  schedule = Scheduler()

  if not schedule.deleteJob(triggerUUID):
    return jsonify({"Error": "Failed to Delete Job from Scheduler"})
#
###

### MqMessage Object - Handles sending delete event to EventBridge
#
  msg = MqMessage()

  message = jsonify({"type": "deleteTrigger",
                     "class": "GoFlowTrigger",
                     "timestamp": int(time.time()),
                     "payload": {"DELETE": triggerUUID}
                    }).get_data(as_text=True)

  if not msg.mqsend(message):
    print("!!! ERROR: Sending Message:", {"DELETED": triggerUUID })
#
###

  return jsonify({"DELETED": triggerUUID })


########## Delete All Trigger ################################################

@triggers.route('/api/triggers', methods= ['DELETE'])
@authenticated
def deleteAllTriggers():

  # Disabling
  return jsonify({"Error": "Not Authorized"}, 401)

  myTriggers = GoFlowTriggers.objects()

  myTriggers.delete()


  msg = MqMessage()

  message = jsonify({"type": "deleteTriggers",
                     "class": "GoFlowTrigger",
                     "timestamp": int(time.time()),
                     "payload": {"DELETE": "All Triggers!"}
                    }).get_data(as_text=True)

  if not msg.mqsend(message):
    print("!!! ERROR: Sending Message: DELETE ALL Triggers")

  return jsonify({"DELETED": "All Triggers!"})
