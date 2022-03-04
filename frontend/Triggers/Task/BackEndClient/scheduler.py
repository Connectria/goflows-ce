import requests
import logging
import json
import configparser
from goflows_frontend import app

class Scheduler():

  def __init__(self):

    # Parse username/password from params.ini
    schedulerConfig = configparser.ConfigParser()
    schedulerConfig.read(app.root_path + '/params.ini')

    self.url = schedulerConfig['Backend']['url']
    print(self.url)

  def getAllJobs(self):

    ## Get Schedule List from goflow-scheduler

    response = requests.get(self.url + "/list")

    if response.status_code == 200:

      logging.info("Response CODE: {}".format(response.status_code))

      result = json.loads(response.content.decode('utf-8'))

      return result
 
    else:

      logging.error("Unable to get All jobs from scheduler")

      return False


  def addJob(self, payload):

    ## Make the Call to GoFlow-Scheduler to create 

    response = requests.post(self.url + "/add",
                            data=json.dumps(payload))

    print(response.content)

    if response.status_code == 200:
 
      logging.info("Response CODE: {}".format(response.status_code))

      result = json.loads(response.content.decode('utf-8'))

      return result

    else:

      logging.error("Unable to add Job to scheduler - {}".format(payload))

      return False 


  def getJob(self, triggerId):

    ## Get Single Job

    response = requests.get(self.url + "/list?triggerId={}".format(triggerId))

    if response.status_code == 200:

      logging.info("Response CODE: {}".format(response.status_code))

      result = json.loads(response.content.decode('utf-8'))

      # If scheduler-list is empty, the trigger ID isn't scheduled
      if result["scheduler-list"]:
        return result
      else:
        return False

    else:

      logging.error("Unable to locate {} in scheduler".format(triggerId))

      return False


  def deleteJob(self, triggerId):

    ## Delete Single Job

    response = requests.delete(self.url + "/remove/triggerId/{}".format(triggerId))

    if response.status_code == 200:

      logging.info("Removed {} from scheduler".format(triggerId))

      result = json.loads(response.content.decode('utf-8'))

      return result

    else:

      logging.error("DELETE from Scheduler Failed - {}".format(triggerId))

      return False


  def runNow(self, payload):

    ## Run Job immediately 

    response = requests.post(self.url + "/runNow",
                             data=json.dumps(payload))

    print(response.content)

    if response.status_code == 200:

      logging.info("Succesfully Scheduled run now - {}".format(payload))

      result = json.loads(response.content.decode('utf-8'))

      return result

    else:

      logging.error("Error Scheduling Job (runNow) - {}".format(payload))

      return False







