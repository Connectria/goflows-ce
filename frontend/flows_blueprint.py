import uuid
import time
import logging
from flask import jsonify, request, render_template, Blueprint
from models import *
from rabbit import MqMessage
from functools import wraps
from decorators import authenticated
from Triggers.Task.BackEndClient.scheduler import Scheduler
from datetime import datetime
from decorators import authenticated, is_admin

from Flows.Query import listOfFlows
from Flows.Query import queryFlows

flows = Blueprint('flows', __name__)

#########################################################
###  /api/flows

# GET ALL
@flows.route('/api/flows', methods = ['GET'])
@authenticated
def getFlows():
  args_hash = request.args
  query_flag = False

  for arg in args_hash:
    query_flag = True

  if query_flag is False:
    logging.info("GET Flows")
    return listOfFlows()

  else:
    return queryFlows(args_hash)


# GET ONE 
@flows.route('/api/flows/<flowUUID>', methods= ['GET'])
@authenticated
def getFlow(flowUUID):

  try:
    myFlow = GoFlows.objects(flowId = flowUUID)[0]

  except ValueError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Invalid Flow ID Value"}), 400

  except IndexError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Flow ID {} Not Found".format(flowUUID)}), 404

  return jsonify(myFlow.getDetails())


# CREATE
@flows.route('/api/flows', methods=['POST'])
@authenticated
def createFlow():

  d = request.get_json()

  print(d)

  myPendingCreate = {}
  for field in GoFlows.getUpdateFields():

    if field in d.keys():
      myPendingCreate[field] = d[field]
    else:
      myPendingCreate[field] = None

  myFlowInputs = []
  if "flowInputs" in d.keys():
    pInputs = d["flowInputs"]
    for input in pInputs:
      if "inputDescription" in input:
        myFlowInputs.append(
          GoFlowInput(
            inputName = input["inputName"],
            inputType = input["inputType"],
            inputDescription = input["inputDescription"]
          )
        )
      else:
        myFlowInputs.append(
          GoFlowInput(
            inputName = input["inputName"],
            inputType = input["inputType"]
          )
        )
  else:
    myFlowInputs = []

  myPendingCreate["flowInputs"] = myFlowInputs

  ret = GoFlows(
    flowId = uuid.uuid4(),
    flowName = myPendingCreate["flowName"],
    flowShortDescription = myPendingCreate["flowShortDescription"],
    flowDescription = myPendingCreate["flowDescription"],
    flowInputs = myPendingCreate["flowInputs"],
    flowOutputs = myPendingCreate["flowOutputs"],
    flowDocument = myPendingCreate["flowDocument"],
    createdOn = datetime.now(),
    updatedOn = datetime.now()
  ).save()

  msg = MqMessage()

  print(msg)

  message = jsonify({"type": "createFlow",
                     "class": "GoFlow",
                     "timestamp": int(time.time()),
                     "payload": ret.getDetails()
                    }).get_data(as_text=True)

  if not msg.mqsend(message):
    print("!!! ERROR: Sending Message:", ret.getDetails())

  return jsonify(ret.getDetails())



# UPDATE/PUT 
@flows.route('/api/flows/<flowUUID>', methods= ['PUT', 'PATCH'])
@authenticated
def updateFlow(flowUUID):

  req_data = request.get_json()

  try:
    myGoFlow = GoFlows.objects(flowId = flowUUID)[0]

  except ValueError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Invalid Flow ID Value"}), 400

  except IndexError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Flow ID {} Not Found".format(flowUUID)}), 404

  print(req_data)

  myPendingUpdate = {}
  for field in GoFlows.getUpdateFields():
    if field in req_data:
      print("FLOW UPDATE Need to add:", GoFlows.getUpdateFields()[field], 
            "=", req_data[field])

      if GoFlows.getUpdateFields()[field] == "flowInputs":
        
        myFlowInputs = []

        pInputs = req_data["flowInputs"]

        for input in pInputs:

          if "inputDescription" in input:
            myFlowInputs.append(
            GoFlowInput(
              inputName = input["inputName"],
              inputType = input["inputType"],
              inputDescription = input["inputDescription"]
              )
            )
          else:
            myFlowInputs.append(
            GoFlowInput(
              inputName = input["inputName"],
              inputType = input["inputType"]
              )
            )

        myPendingUpdate[GoFlows.getUpdateFields()[field]] = myFlowInputs

      else:
        myPendingUpdate[GoFlows.getUpdateFields()[field]] = req_data[field]

  for update in myPendingUpdate:
    print("Trying to update: ", update, " to ", myPendingUpdate[update])
    myGoFlow[update] = myPendingUpdate[update]

  # Set updated field
  myGoFlow["updatedOn"] = datetime.now()

  myGoFlow.save()

  msg = MqMessage()

  message = jsonify({"type": "updateFlow",
                     "class": "GoFlow",
                     "timestamp": int(time.time()),
                     "payload": myGoFlow.getDetails()
                    }).get_data(as_text=True)

  if not msg.mqsend(message):
    print("!!! ERROR: Sending Message:", myGoFlow.getDetails())

  return jsonify(myGoFlow.getDetails())


# DELETE ONE
@flows.route('/api/flows/<flowUUID>', methods= ['DELETE'])
@authenticated
def deleteFlow(flowUUID):

  try:
    myFlow = GoFlows.objects(flowId = flowUUID)[0]

  except ValueError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Invalid Flow ID Value"}), 400

  except IndexError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Flow ID {} Not Found".format(flowUUID)}), 404

  else:
    myFlow.delete()

  msg = MqMessage()

  message = jsonify({"type": "deleteFlow",
                     "class": "GoFlow",
                     "timestamp": int(time.time()),
                     "payload": {"DELETED": flowUUID}
                    }).get_data(as_text=True)

  if not msg.mqsend(message):
    print("!!! ERROR: Sending Message: DELETE flow", flowUUID)

  return jsonify({"DELETED": flowUUID})


# DELETE ALL 
@flows.route('/api/flows', methods= ['DELETE'])
@authenticated
def deleteAllFlows():
  # Disabling
  return jsonify({"Error": "Not Authorized"}, 401)


  myFlows = GoFlows.objects()
  myFlows.delete()

  msg = MqMessage()

  message = jsonify({"type": "deleteFlow",
                     "class": "GoFlow",
                     "timestamp": int(time.time()),
                     "payload": {"DELETED": "All Flows!"}
                    }).get_data(as_text=True)

  if not msg.mqsend(message):
    print("!!! ERROR: Sending Message: DELETE ALL FLOWS")

  return jsonify({"DELETED": "All Flows!"})


# GET TRIGGERS FOR FLOW
@flows.route('/api/flows/<flowUUID>/triggers', methods= ['GET'])
@authenticated
def getTriggersForAFlow(flowUUID):

  try:
    triggers = GoFlowTriggers.objects(flowIds = flowUUID)

  except ValueError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Invalid Flow ID Value"}), 400

  except IndexError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Flow ID {} Not Found".format(flowUUID)}), 404

  r_triggers = []
  for trigger in triggers:
    r_triggers.append(trigger.getInfo())

  return jsonify({"triggers": r_triggers })


# RUN FLOW ( CREATES A JOB ) 
@flows.route('/api/flows/<flowUUID>/run', methods = ['POST'])
@authenticated
def runFlow(flowUUID):
  req_data = request.get_json()

  try:
    goFlow = GoFlows.objects(flowId = flowUUID)[0] #<- Get our source flow

  except ValueError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Invalid Flow ID Value"}), 400

  except IndexError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Flow ID {} Not Found".format(flowUUID)}), 404

  if req_data is None:
    runInputs = None
    logging.info("req_data is None")
  else:
    logging.info("req_data (inputs)" + str(req_data["flowInputs"]))
    runInputs = GoFlowHelper.jsonArrayToRunTimeInputs(req_data["flowInputs"])

  myTriggerName = "RunNow-{}-{}".format(flowUUID,int(time.time())) 


  # The bug I think has todo with string vs UUID
  myNewTrigger = GoFlowTriggers(
    triggerId = str(uuid.uuid4()),
    triggerName = "RunNow-{}-{}".format(flowUUID,int(time.time())),
    triggerActive = True,
    flowIds = [flowUUID],
    flowInputs = runInputs,
    createdOn = datetime.now(),
    updatedOn = datetime.now()
  )

  myNewTrigger.save()
  logging.info("New Trigger: ", myNewTrigger.getInfo())


  ### Assemble payload for call to scheduler

  payload = {}
  payload["triggerId"] = myNewTrigger.triggerId
  payload["triggerName"] = myNewTrigger.triggerName

  flow = {}
  #
  # Edited by AKB adding flowIds[0] since it was being passed as an array to the scheduler
  flow["FlowID"] = str(myNewTrigger.flowIds[0])

  #flow["funcName"] = "ScheduleFlow"
  flow["funcName"] = goFlow.flowDocument["funcName"]

  flow["flowDescription"] = goFlow.flowDescription
  flow["flowName"] = goFlow.flowName

  payload["flowIDs"] = [flow]

  # Inputs are stored in Mongo as a object.
  # The following for is to convert back to str
  inputs = []

  for input in myNewTrigger.flowInputs:

    inputs.append({"inputName": input.inputName,
                   "inputValue": input.inputValue})

  payload["inputs"] = inputs

  print(payload)

  # Instantiate a Scheduler Object
  schedule = Scheduler()

  # Schedule the job to run Immediately on backend
  if not schedule.runNow(payload):
    return jsonify({"Error": "Failed to Run Job on Scheduler"})


  ### Post to EventBridge
  msg = MqMessage()

  message = jsonify({"type": "runFlow",
                     "class": "GoFlowJob",
                     "timestamp": int(time.time()),
                     "payload": myNewTrigger.getInfo()
                    }).get_data(as_text=True)

  if not msg.mqsend(message):
    print("!!! ERROR: Sending Message:", myNewTrigger.getInfo())

  return jsonify(myNewTrigger.getInfo())


