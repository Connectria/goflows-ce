from goflows_frontend import db
from mongoengine import *
from flask import jsonify
import datetime
import json

# Valid Types - string, number, object, array, boolean, null
class GoFlowInput(db.EmbeddedDocument):
    inputName = StringField()
    inputType = StringField()
    inputDescription = StringField() 
    pass

class GoFlowOutput(db.EmbeddedDocument):
    outputName = StringField()
    outputType = StringField()
    outputDescription = StringField() 
    pass

class GoFlows(db.Document):
    flowId = UUIDField(required=True)
    flowName = StringField(required=True)
    flowShortDescription = StringField()
    flowDescription = StringField()

    flowInputs = EmbeddedDocumentListField(GoFlowInput)
    flowOutputs = EmbeddedDocumentListField(GoFlowOutput)
    flowDocument = DictField()

    createdOn = DateTimeField()
    updatedOn = DateTimeField()

    def getUpdateFields():
        return {
            "flowName": "flowName",
            "flowShortDescription": "flowShortDescription",
            "flowDescription": "flowDescription",
            "flowInputs": "flowInputs",
            "flowOutputs": "flowOutputs",
            "flowDocument": "flowDocument"
        }

    def getInfo(self):
        return {
            "flowId": self.flowId,
            "flowName": self.flowName,
            "flowShortDescription": self.flowDescription,
            "flowDescription": self.flowDescription,
            "flowInputs": self.flowInputs,
            "flowOutputs": self.flowOutputs,
            "createdOn": self.createdOn,
            "updatedOn": self.createdOn
        }

    def getDetails(self):
        return {
            "flowId": self.flowId,
            "flowName": self.flowName,
            "flowShortDescription": self.flowDescription,
            "flowDescription": self.flowDescription,
            "flowInputs": self.flowInputs,
            "flowOutputs": self.flowOutputs,
            "flowDocument": self.flowDocument,
            "createdOn": self.createdOn,
            "updatedOn": self.updatedOn
        }


    meta = {'collection': 'GoFlows'}
    pass

class RunTimeInput(db.EmbeddedDocument):
    inputName = StringField()
    inputValue = StringField()
    pass


class GoFlowJobs(db.Document):
    #Summary Fields
    jobId = UUIDField(required=True)
    jobName = StringField()
    srcFlowName = StringField(required=True)
    srcFlow = UUIDField(required=True)
    jobXref = StringField()

    #Input Information
    jobFlowInputs = EmbeddedDocumentListField(RunTimeInput)
    jobExtendedAttr = DictField()

    #Status Information
    jobControlStatus = StringField()
    jobStepStatus = StringField()

    #Date Times
    jobCreateTime = LongField()
    jobStartTime = LongField()
    jobStopTime = LongField()

    # SSP-60
    jobDuration = FloatField()
    jobInfo = StringField()
    jobStepID = IntField()
    triggerId = UUIDField()

    #Not sure if I am going to use these this way ...
    jobCustomer = StringField()
    jobHistoryCallback = StringField()
    jobControlExit = IntField()

    def getInfo(myObject):

        return {
            "jobId": myObject.jobId,
            "jobName": myObject.jobName,
            "jobSrcFlowName": myObject.srcFlowName,
            "jobSrcFlowId": myObject.srcFlow, 
            "jobFlowInputs": myObject.jobFlowInputs,
            "jobCreateTime": myObject.jobCreateTime,
            "jobControlStatus": myObject.jobControlStatus,
            "jobStepStatus": myObject.jobStepStatus,
            "jobStartTime": myObject.jobStartTime,
            "jobStopTime": myObject.jobStopTime,
            "jobDuration": myObject.jobDuration,
            "jobInfo": myObject.jobInfo,
            "triggerId": myObject.triggerId,
            "jobControlExit": myObject.jobControlExit
        }

    def getUpdateFields():
        return {
            "jobId":             "jobId",
            "jobName":           "jobName",
            "srcFlowName":       "srcFlowName",
            "srcFlow":           "srcFlow",
            "flowInputs":        "jobFlowInputs",
            "jobCreateTime":     "jobCreateTime",
            "jobControlStatus":  "jobControlStatus",
            "jobStepStatus":     "jobStepStatus",
            "jobStartTime":      "jobStartTime",
            "jobStopTime":       "jobStopTime",
            "jobDuration":       "jobDuration",
            "jobInfo":           "jobInfo",
            "triggerId":         "triggerId",
            "jobControlExit":    "jobControlExit"
        }

    meta = {'collection': 'GoFlowJobs'}

#Need to send notification when these are expiring or running out
class GoFlowTriggers(db.Document):
    triggerId = UUIDField(required=True)
    triggerName = StringField()
    triggerActive = BooleanField()    
    flowIds = ListField(UUIDField())                        # Flow(s) to run when this is triggered
    flowInputs = EmbeddedDocumentListField(RunTimeInput)
    triggerIds = ListField(UUIDField())                     # Other Triggers to Fire off, allows for a "Trigger Chain"
    triggerLogic = DictField()
    schedulerJobID = StringField()                          # Used for external scheduler    
    createdOn = DateTimeField()
    updatedOn = DateTimeField()
 
    def getInfo(self):
        return {
            "triggerId": self.triggerId,
            "name": self.triggerName,
            "active": self.triggerActive,
            "flowIds": self.flowIds,            
            "inputs": self.flowInputs,
            "triggerIds": self.triggerIds,
            "triggerLogic": self.triggerLogic,
            "createdOn": self.createdOn,
            "updatedOn": self.updatedOn
        }
    
    def getUpdateFields():        
        return {
            "triggerId":    "triggerId",
            "name":         "triggerName" ,
            "inputs":       "flowInputs",
            "active":       "triggerActive" ,
            "flowIds":      "flowIds",
            "triggerIds":   "triggerIds",
            "triggerLogic": "triggerLogic"
        }

    meta = {'collection': 'GoFlowTriggers'}

class GoFlowHistory(db.Document):
    historyId = UUIDField(required=True)
    jobId = UUIDField()
    stepId = UUIDField()
    timeStamp = DateTimeField()
    meta = {'collection': 'GoFlowHistory'}

class GoFlowHelper():
 
    """Stringify Json when Appropriate otherwise leave it alone"""
    @staticmethod
    def stringifyWhenNotString(payload):
        """
        Test with something like this print("PAYLOAD: ", type(payload))
        """
        if isinstance(payload, str):
            o = payload
        else:
            o = json.dumps(payload)

        return o

    @staticmethod
    def jsonArrayToRunTimeInputs(myInputs):

        runInputs = []

        for myInput in myInputs:
          myInputName = myInput["inputName"]

          # We need to stringify JSON payloads
          #  * At this point its a primitive i.e. str, int, list, tuple
          myInputValue = GoFlowHelper.stringifyWhenNotString(myInput["inputValue"])

          runInputs.append(
              RunTimeInput(
                  inputName = myInputName,
                  inputValue = myInputValue
              )
          )

        return runInputs

class GoFlowUsers(db.Document):

    userId = UUIDField()
    userName = StringField(required=True)
    email = StringField(required=True)
    key = BinaryField(required=True)
    admin = BooleanField(default=False)

    def getInfo(myObject):
        return {
            "userId": myObject.userId,
            "userName": myObject.userName,
            "email": myObject.email,
            "key": myObject.key.decode(),
            "admin": myObject.admin
        }

    def getUpdateFields():
        return {
            "key":         "key",
            "userName":    "userName",
            "email":       "email",
            "admin":       "admin"
        }

    meta = {'collection': 'GoFlowUsers'}


