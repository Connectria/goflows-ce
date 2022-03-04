import logging

from models import *
from croniter import croniter
from datetime import datetime
from rabbit import MqMessage

from Triggers.Task.BackEndClient.scheduler import Scheduler

class TaskBackEndClient():
    def __init__(self, myTrigger):
        self.myTrigger = myTrigger
    
    def submitNewTaskTrigger(self):
        myNewTrigger = self.myTrigger
        
        # Assemble Payload for GoFlow-Scheduler
        payload = {}
        payload["triggerId"] = myNewTrigger.triggerId
        payload["triggerName"] = myNewTrigger.triggerName
        payload["flowIDs"] = []

        # The following generates the flowIDs dict for payload
        #
        for myFlowId in myNewTrigger.flowIds:      
            myFlow = GoFlows.objects(flowId = myFlowId).first()      
            this_flow = {}
            this_flow["FlowID"] = myFlowId
            this_flow["funcName"] = myFlow.flowDocument["funcName"]
            this_flow["flowDescription"] = myFlow.flowDescription
            this_flow["flowName"] = myFlow.flowName
            payload["flowIDs"].append(this_flow)


        # The following is used to determine if the trigger is at vs. cron job
        # and sets the necessary parameters for the API call
        temporal_rule = myNewTrigger.triggerLogic["temporal-rule"]

        job_type = [key for key in temporal_rule.keys()]

        if job_type[0] == "cron-expression":
            for k,v in temporal_rule["cron-expression"].items():
                if k == "expression":
                    if True: #croniter.is_valid(v): #<<< Disable this check akb, not working with TZ
                        payload["dutyCycle"] = v
                    else:
                        logging.error("ERROR: Invalid Cron expression! {}".format(v))
                        return jsonify({"Error": "Invalid Cron Expression"}), 400

                elif k == "repeat-until":

                    if len(str(v)) == 10 and str(v).isnumeric():
                        payload["repeat-until"] = v 
                    else:
                        logging.error("ERROR: Invalid repeat-until for Cron ! {}".format(v))
                        return jsonify({"Error": "Invalid repeat-until value (cron)"}), 400

        elif job_type[0] == "at-list":

            at_list = temporal_rule["at-list"]

            # AT Validation
            new_at_list = [ y for y in at_list if len(str(y)) == 10 and str(y).isnumeric()]

            if len(new_at_list) == len(temporal_rule["at-list"]):
                payload["at-list"] = temporal_rule["at-list"]
            else:
                logging.error("ERROR: Invalid At-list! {}".format(at_list))
                return jsonify({"Error": "Invalid Parameter in at-list"}), 400
        

        # Inputs are stored in Mongo as a object.
        # The following for is to convert back to str
        inputs = []

        for input in myNewTrigger.flowInputs:
            inputs.append({"inputName": input.inputName,
                        "inputValue": input.inputValue})

        payload["inputs"] = inputs

        ### Scheduler Object - Handles Backend calls to scheduler
        #
        schedule = Scheduler()
        #
        ###
        if myNewTrigger.triggerActive:

            # Create the job on the backend
            if not schedule.addJob(payload):
                return jsonify({"Error": "Failed to Schedule Job"})

        else:
            logging.info("New Trigger created with active=false, skipping scheduler..")
            logging.info(payload)

        return "I Tried but don't really know what happend (submit to backend)"
    
    def updateTaskTrigger(self):
        myTrigger = self.myTrigger
        triggerUUID = self.myTrigger.triggerId
        # Assemble Payload for GoFlow-Scheduler
        payload = {}
        payload["triggerId"] = str(myTrigger.triggerId)
        payload["triggerName"] = myTrigger.triggerName
        payload["flowIDs"] = []

        # Need to keep adding around line 274 from triggers_blueprint.py
        # The following generates the flowIDs dict for payload
        #
        for flow in myTrigger.flowIds:

            # Flow ID Validation
            try:
                myFlow = GoFlows.objects(flowId = flow).first()
            except ValueError as e:
                logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
                return jsonify({"Error": "Invalid Flow ID Value"}), 400
            else:
                if not myFlow:
                    logging.error("Flow ID Not Found: {}".format(payload))
                    return jsonify({"Error": "Flow ID not found"}), 404

        this_flow = {}
        this_flow["FlowID"] = str(flow)

        this_flow["funcName"] = myFlow.flowDocument["funcName"]
        this_flow["flowDescription"] = myFlow.flowDescription
        this_flow["flowName"] = myFlow.flowName

        payload["flowIDs"].append(this_flow)

        # The following is used to determine if the trigger is at vs. cron job
        # and sets the necessary parameters for the API call
        if 'temporal-rule' in myTrigger.triggerLogic:
            temporal_rule = myTrigger.triggerLogic["temporal-rule"]
            job_type = [key for key in temporal_rule.keys()]
            if job_type[0] == "cron-expression":
                for k,v in temporal_rule["cron-expression"].items():
                    if k == "expression":
                        if True:#croniter.is_valid(v): #<<< Disabled, not working with TZ AKB
                            payload["dutyCycle"] = v
                        else:
                            logging.error("ERROR: Invalid Cron expression! {}".format(v))
                            return jsonify({"Error": "Invalid Cron Expression"}), 400
                    elif k == "repeat-until":
                        if len(str(v)) == 10 and str(v).isnumeric():
                            payload["repeat-until"] = v
                        else:
                            logging.error("ERROR: Invalid repeat-until for Cron ! {}".format(v))
                            return jsonify({"Error": "Invalid repeat-until value (cron)"}), 400
            elif job_type[0] == "at-list":
                at_list = temporal_rule["at-list"]
                # AT Validation
                new_at_list = [ y for y in at_list if len(str(y)) == 10 and str(y).isnumeric()]
                if len(new_at_list) == len(temporal_rule["at-list"]):
                    payload["at-list"] = temporal_rule["at-list"]
                else:
                    logging.error("ERROR: Invalid At-list! {}".format(at_list))
                    return jsonify({"Error": "Invalid Parameter in at-list"}), 400

        # Inputs are stored in Mongo as a object.
        # The following for is to convert back to str
        inputs = []

        for input in myTrigger.flowInputs:
          inputs.append({"inputName": input.inputName,
                    "inputValue": input.inputValue})

        payload["inputs"] = inputs

        #### Scheduler Object - Handles Create, Delete/Create of Job on backend    
        schedule = Scheduler()
        job = schedule.getJob(triggerUUID)
        if myTrigger.triggerActive:
        # Check for the job on the backend
            if not job:
                # Re-create the Trigger on the backend
                print("--- Recreating Job on Backend --- ")
                schedule.addJob(payload)
                #return jsonify({"Error": "Failed to find Job with Trigger ID"}) 
            else:
                # Delete the Trigger from the backend
                print("--- Deleting Job from Backend --- ")
                schedule.deleteJob(triggerUUID)
                # Re-create the Trigger on the backend
                print("--- Recreating Job on Backend --- ")
                schedule.addJob(payload)
        else:
            if not job:
                return jsonify({"Status": "Job not in scheduler already."})
            else:
                # Delete the Trigger from the backend
                print("--- Deleting Job from Backend --- ")
                return jsonify(schedule.deleteJob(triggerUUID))

        pass
        
