import sys,os 
import logging
import subprocess
import threading

from flask import Flask, jsonify, Blueprint
from flask_cors import CORS
from flask_swagger_ui import get_swaggerui_blueprint
from flask_mongoengine import MongoEngine

file_dir = os.path.dirname(__file__)
sys.path.append(file_dir)

app = Flask(__name__)
CORS(app)

# Logging configuration
#logging.basicConfig(filename='/var/log/nginx/goflows-frontend/app.log', 
logging.basicConfig(filename=app.root_path + '/logs/app.log', 
                    format='%(asctime)s - %(levelname)s : %(message)s',
                    level=logging.INFO, datefmt='%b %d %H:%M:%S')

# Change default pika logging to Warning instead of INFO.
logging.getLogger("pika").setLevel(logging.WARNING)


# Load config from config.py
app.config.from_object('config.BaseConfig')

db = MongoEngine(app)

#### BLUEPRINTS

from api_blueprint import api as api_blueprint
app.register_blueprint(api_blueprint)

#from web_blueprint import web as web_blueprint
#app.register_blueprint(web_blueprint)

from flows_blueprint import flows as flows_blueprint
app.register_blueprint(flows_blueprint)

from jobs_blueprint import jobs as jobs_blueprint
app.register_blueprint(jobs_blueprint)

from triggers_blueprint import triggers as triggers_blueprint
app.register_blueprint(triggers_blueprint)

from history_blueprint import history as history_blueprint
app.register_blueprint(history_blueprint)


####################################################
SWAGGER_URL = '/swagger'
API_URL = '/static/swagger.yaml'

SWAGGERUI_BLUEPRINT = get_swaggerui_blueprint(
    SWAGGER_URL,
    API_URL,
    config={
        'app_name': "GoFlows Frontend API",
    }
)
app.register_blueprint(SWAGGERUI_BLUEPRINT, url_prefix=SWAGGER_URL)
### end swagger specific ###


