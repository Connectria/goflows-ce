from flask import jsonify, request, render_template, Blueprint
import uuid
import base64
import datetime
import sys
import time
import json 
import os
import secrets
import logging

from models import *

from rabbit import MqMessage

from sessions import Context

from decorators import authenticated, is_admin

api = Blueprint('api', __name__)


@api.route('/api')
def api_base():

  # Base URL of the API.  Used for documenting API Endpoints/Usage
  return render_template('api-help.html')  

@api.route('/ping')
def apiPing():
    return jsonify({ "ping": "pong"  })

@api.route('/api/test')
def apiTest():        

    c = Context(request)

    return jsonify(
      {
        "message": "Say Hello to GoFlows", 
        "CustomerId": c.tenantId,
      }
    )
  

#----------------------------------------------------
# Utility Endpoints Allow for CRUD Operations for Mock Operations
#----------------------------------------------------
@api.route('/api/mq/test')
def mqtest():

  message = jsonify({"type": "testMQ",
                     "class": "test",
                     "timestamp": int(time.time()),
                     "payload": {"message": "test"}
                   }).get_data(as_text=True)

  msg = MqMessage()

  if msg.mqsend(message):
    return jsonify({"return": True})
  else:
    return jsonify({"return": False})


####################
# Get All Users

@api.route('/api/users', methods=['GET'])
@is_admin
def getUsers():

  users = GoFlowUsers.objects()

  allUsers = {"users": []}

  for user in users:
    allUsers["users"].append(
          { "userName": user.userName, "email": user.email,
            "userId": user.userId, "admin": user.admin}
    )

  return jsonify(allUsers)



####################
# Delete a User

@api.route('/api/users/<userId>', methods=['DELETE'])
@is_admin
def deleteUser(userId):

  # Check for Invalid/Non-existand Trigger IDs

  try:
    thisUser = GoFlowUsers.objects(userId = userId)[0]
    thisUser.delete()
    return jsonify({"Status": "User Deleted"})

  except IndexError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "User ID {} Not Found".format(userId)}), 404


###################
# Update User

@api.route('/api/users/<userId>', methods=['PUT','PATCH'])
@is_admin
def updateUser(userId):

  req_data = request.get_json()

  try:
    thisUser = GoFlowUsers.objects(userId = userId)[0]
  except IndexError as e:
    logging.error("ERROR: {} | {} - {}".format(e,request.url, request.method))
    return jsonify({"Error": "User ID {} Not Found".format(userId)}), 404

  # Now go though the inputs and figure out what was passed
  myPendingUpdate = {}
  for field in GoFlowUsers.getUpdateFields():
    if field in req_data:

      if field == "key":
        myPendingUpdate[GoFlowUsers.getUpdateFields()[field]] = req_data[field].encode()
      else:
        myPendingUpdate[GoFlowUsers.getUpdateFields()[field]] = req_data[field]
  
      print("Need to add:",
            GoFlowUsers.getUpdateFields()[field], "=", req_data[field])


  # itt though the updates and then save
  for update in myPendingUpdate:
    print("Trying to update: ", update, " to ", myPendingUpdate[update])
    thisUser[update] = myPendingUpdate[update]

  thisUser.save()

  return jsonify({"Updated": thisUser.getInfo()})



####################
# Create User

@api.route('/api/users', methods=['POST'])
@is_admin
def createUser():

  req_data = request.get_json()

  # Check if all required values are present in request
  if not req_data.keys() >= {"userName", "email"}:

    return jsonify({"Error": "Missing One or More required fields"}), 400

  else:

    thisUser = GoFlowUsers.objects(userName = req_data["userName"])

    if not thisUser:

      myNewUser = GoFlowUsers(
      userName = req_data["userName"],
      email = req_data["email"],
      key = secrets.token_urlsafe(48).encode(),
      userId = uuid.uuid4(),
      admin = req_data["admin"]
      )

      myNewUser.save()

      return jsonify({"Created": myNewUser.getInfo()})

    else:
      return jsonify({"Error": "User {} already exists!".format(req_data["userName"])})

 


@api.route('/api/usertest')
@authenticated
def test():
  return jsonify({"Status": True})

