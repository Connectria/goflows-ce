from flask import jsonify, request, Blueprint, render_template

web = Blueprint('web', __name__)

############################################################################

@web.route('/')
def index():

  # Base URL for Web UI Content (index.html) 
 
  return render_template('index.html')

