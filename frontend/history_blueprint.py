from flask import jsonify, request, render_template, Blueprint

history = Blueprint('history', __name__)

#----------------------------------------------------
'''
*** History ***
GET - /api/history/flow/<FLOW-UUID>
GET - /api/history/job/<JOB-UUID>
POST - /api/history/search

'''
#----------------------------------------------------

