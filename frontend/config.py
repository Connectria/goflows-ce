import configparser
from goflows_frontend import app


# Parse username/password from mongo.ini
mongoConfig = configparser.ConfigParser()
mongoConfig.read(app.root_path + '/params.ini')

username = mongoConfig['Mongo']['username']
password = mongoConfig['Mongo']['password']
host = mongoConfig['Mongo']['host']

###

class BaseConfig():
  """Base configuration."""
  MONGODB_ALIAS = 'default'
  MONGODB_DB = 'goflow-mock'
  MONGODB_SETTINGS = {
                        'db':'goflow-mock',
                        'alias':'default',
                        'host': host,
                        'username': username,
                        'password': password 
                     }

  #SEND_FILE_MAX_AGE_DEFAULT = 0

