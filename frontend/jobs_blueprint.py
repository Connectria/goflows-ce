import uuid
import time
import logging

from flask import jsonify, request, render_template, Blueprint
from models import *
from rabbit import MqMessage
from decorators import authenticated, is_admin

jobs = Blueprint('jobs', __name__)

#####################################################################

# GET ALL 
@jobs.route('/api/jobs', methods = ['GET'])
@authenticated
def getJobs():
    req_data = request.get_json()
    #jobs = GoFlowJobs.objects()#.order_by('jobCreateTime')
    #jobs = GoFlowJobs.objects.order_by('-id').first()#.order_by('jobCreateTime')
    jobs = GoFlowJobs.objects()

    jobOutput = []

    for job in jobs:
      jobOutput.append(job.getInfo())

    return jsonify( { "jobs": jobOutput } )


# GET ONE 
@jobs.route('/api/jobs/<jobUUID>', methods = ['GET'])
@authenticated
def jobInfo(jobUUID):

  try:
    myJob = GoFlowJobs.objects(jobId = jobUUID)[0]

  except ValueError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Invalid Job ID Value"}), 400

  except IndexError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Job ID {} Not Found".format(jobUUID)}), 404

  return myJob.getInfo()

####################################################################

# CREATE JOB
@jobs.route('/api/jobs', methods = ['POST'])
@authenticated
def createJob():

  req_data = request.get_json()


  logging.info("New Request: {}".format(req_data))

  if req_data["flowInputs"]:
    print(req_data)
    runInputs = GoFlowHelper.jsonArrayToRunTimeInputs(req_data["flowInputs"])
  else:
    runInputs = None

  newJob = GoFlowJobs(
    jobId = req_data["jobId"],
    jobName = req_data["jobName"],
    srcFlowName = req_data["srcFlowName"],
    srcFlow = req_data["srcFlow"],
    jobControlStatus = req_data["jobControlStatus"],
    jobStepStatus = req_data["jobStepStatus"],
    jobCreateTime = req_data["jobCreateTime"],
    jobStartTime = req_data["jobStartTime"],
    jobStopTime = req_data["jobStopTime"],
    jobDuration = req_data["jobDuration"],
    jobInfo = req_data["jobInfo"],
    jobStepID = req_data["jobStepId"],
    triggerId = req_data["triggerId"],
    jobControlExit = req_data["jobControlExit"],
    jobFlowInputs = runInputs
  ).save()

  msg = MqMessage()

  message = jsonify({"type": "createJob",
                     "class": "GoFlowJob",
                     "timestamp": int(time.time()),
                     "payload": {"CreatedJob": newJob.getInfo()}
                    }).get_data(as_text=True)

  if not msg.mqsend(message):
    print("!!! ERROR: Sending Message: DELETE Job", jobUUID)
  
  return newJob.getInfo()


# UPDATE JOB
@jobs.route('/api/jobs/<jobUUID>', methods = ['PUT', 'PATCH'])
@authenticated
def updateJob(jobUUID):

  req_data = request.get_json()

  try:
    myJob = GoFlowJobs.objects(jobId = jobUUID)[0]

  except ValueError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Invalid Job ID Value"}), 400

  except IndexError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Job ID {} Not Found".format(jobUUID)}), 404

  myPendingUpdate = {}

  for field in GoFlowJobs.getUpdateFields(): 

    if field in req_data:
      if field == "flowInputs":
        # Needed to convert the JSON to a input object
        runInputs = GoFlowHelper.jsonArrayToRunTimeInputs(req_data["flowInputs"])
        myPendingUpdate[GoFlowJobs.getUpdateFields()[field]] = runInputs
      else:
        print("FLOW UPDATE Need to add:", GoFlowJobs.getUpdateFields()[field], "=", req_data[field])
        myPendingUpdate[GoFlowJobs.getUpdateFields()[field]] = req_data[field]

  print(myPendingUpdate)
  for update in myPendingUpdate:
    print("Trying to update: ", update, " to ", myPendingUpdate[update])
    myJob[update] = myPendingUpdate[update]

  myJob.save()

  msg = MqMessage()

  message = jsonify({"type": "updateJob",
                     "class": "GoFlowJob",
                     "timestamp": int(time.time()),
                     "payload": {"UpdatedJob": myJob.getInfo()}
                    }).get_data(as_text=True)

  if not msg.mqsend(message):
    print("!!! ERROR: Sending Message: DELETE Job", jobUUID)


  return myJob.getInfo()



# DELETE #
@jobs.route('/api/jobs/<jobUUID>', methods= ['DELETE'])
@authenticated
def deleteJob(jobUUID):

  try:
    myJob = GoFlowJobs.objects(jobId = jobUUID)[0]

  except ValueError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Invalid Job ID Value"}), 400

  except IndexError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "Job ID {} Not Found".format(jobUUID)}), 404

  else:
    myJob.delete()

  msg = MqMessage()

  message = jsonify({"type": "deleteJob",
                     "class": "GoFlowJob",
                     "timestamp": int(time.time()),
                     "payload": {"DELETED": jobUUID}
                    }).get_data(as_text=True)

  if not msg.mqsend(message):
    print("!!! ERROR: Sending Message: DELETE Job", jobUUID)

  return jsonify({"DELETED": jobUUID})



# DELETE ALL #
# NOTE: This feautre is disabled
"""
@jobs.route('/api/jobs', methods= ['DELETE'])
@authenticated
def deleteAllJobs():
  myJobs = GoFlowJobs.objects()
  myJobs.delete()

  msg = MqMessage()

  message = jsonify({"type": "deleteJobs",
                     "class": "GoFlowJob",
                     "timestamp": int(time.time()),
                     "payload": {"DELETED": "All Jobs!"}
                    }).get_data(as_text=True)

  if not msg.mqsend(message):
    print("!!! ERROR: Sending Message: DELETE ALL JOBS")

  return jsonify({"DELETED": "All Jobs!"})
"""
