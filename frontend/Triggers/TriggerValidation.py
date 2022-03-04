import logging
from models import *

class GenericTriggerValidation():

    def __init__(self, myTrigger = None):
        self.myTrigger = myTrigger

    # We want to be able to call this without creating the object?
    def checkForRequiredTopLevelFields(req_data):
        check_flag = False
        check_message = []

        required_fields_to_create_trigger = {"name", "active", "flowIds", "inputs",
                            "triggerLogic"}

        logging.info("Generic Trigger Validation checkForRequiredTopLevelFeilds: {}".format(req_data))
        if req_data == None:
            check_flag = False
            check_message.append( { 
                "Error": "Empty POST Payload",
                "Required Fields": str(required_fields_to_create_trigger) 
                })

        elif not req_data.keys() >= required_fields_to_create_trigger:
            check_flag = False
            check_message .append( { 
                "Error": "Missing One or More required fields",
                "Required Fields": str(required_fields_to_create_trigger) 
                })
        else:
            check_flag = True

        return {"check_flag": check_flag, "check_message": check_message}

    def checkValidFlowId(self,myFlowId):
        check_flag = False
        check_message = "Generic Check for FlowId"
        # Flow ID Validation
        try:
            myFlow = GoFlows.objects(flowId = myFlowId).first()            
            check_flag = True
        except ValueError as e:
            #logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
            check_flag = False#jsonify({"Error": "Invalid Flow ID Value"}), 400
            check_message = "Invalid FlowId Value"
        else:
            if not myFlow:
                logging.error("FlowId Not Found: {}".format(myFlowId))
                check_message = "FlowId Not Found: {}".format(myFlowId)                
                check_flag = False

        return {"check_flag": check_flag, "check_message": check_message}

class TriggerLogicChecker():

    def __init__(self,triggerLogic):
        self.myTriggerLogic = triggerLogic
        self.triggerType = None

    #Trigger Logic Sanity Check
    def sanityCheck(self):
        must_have_one_fields = {"temporal-rule","event-rules","text"}
        sanity_flag = False
        sanity_count = 0
        sanity_message = "TriggerLogic Failed General Sanity Test"
        triggerLogic = self.myTriggerLogic

        for logic_key in triggerLogic.keys():
            if logic_key in must_have_one_fields:
                typeHint = logic_key
                sanity_count += 1        

        if sanity_count == 1:
            sanity_flag = True            
            sanity_message = "PASSED TriggerLogic SANITY CHECK, good job :) "
            
            if typeHint == 'temporal-rule':
                self.triggerType = 'Task'

            elif typeHint == 'event-rules':
                self.triggerType = 'Event'
            
            else:
                self.triggerType = 'Unkown'

        elif sanity_count == 0:
            sanity_flag = False
            sanity_message = "FAILED SANITY CHECK: must have at least one " + str(must_have_one_fields)

        elif sanity_count > 1:
            sanity_flag = False
            sanity_message = "FAIILED SANITY CHECK: you have more fields then are supported, only use one of the following" + str(must_have_one_fields)
        
        # One more case "triggerLogic": {"Proto":[], "temporal-rule":[]} # <<<< One is ok the other is not
        # But also need to support "triggerLogic": {"triggerType":"Event", "triggerSubType":"Opsgenie","event-rules":[]} 
             

        logging.info("TriggerLogicSanityChcker reports " + str(sanity_count))
        logging.info("Detected Trigger Type of " + str(self.triggerType))

        self.sanity_flag = sanity_flag
        self.sanity_message = sanity_message

        return sanity_flag

