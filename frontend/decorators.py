import jsonify
import requests
import logging

from functools import wraps
from flask import request
from models import GoFlowUsers

#######################################################################
#  This file contains decorators to be used in multiple blueprints

def authenticated(f):
  @wraps(f)
  def wrap(*args, **kwargs):

    header_key = request.headers.get('Key')

    if header_key:

      user_token = header_key.split(':')

      FlowUser = GoFlowUsers.objects(userName = user_token[0]).first()

      if FlowUser and user_token[1] == FlowUser.key.decode():

        logging.info("{} {} - User:{}".format(request.method, request.endpoint,
                                              FlowUser.userName))
        
        return f(*args, **kwargs)
      else:
        return({"Error": "Invalid User/Key"}, 401)

    else:
      return({"Error": "Authentication Required"}, 401)

  return wrap


def is_admin(f):
  @wraps(f)
  def wrap(*args, **kwargs):

    header_key = request.headers.get('Key')

    if header_key:

      user_token = header_key.split(':')

      FlowUser = GoFlowUsers.objects(userName = user_token[0]).first()

      # Check if the  user was found and if the key is correct  
      if FlowUser and user_token[1] == FlowUser.key.decode():

        # Check if the user has admin=True
        if FlowUser.admin:

          logging.info("{} {} - User:{}".format(request.method, 
                                                request.endpoint,
                                                FlowUser.userName))

          return f(*args, **kwargs)

        else:

          return({"Error": "Insufficient Permissions"}, 401)

      else:

        return({"Error": "Invalid User/Key"}, 401)

    else:
      return({"Error": "Authentication Required"}, 401)

  return wrap








