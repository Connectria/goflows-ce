import unittest
import configparser
from pymongo import MongoClient


### Parse Mongo Config from params.ini
mongoConfig = configparser.ConfigParser()
mongoConfig.read('../params.ini')

username = mongoConfig['Mongo']['username']
password = mongoConfig['Mongo']['password']
host = mongoConfig['Mongo']['host']
new_host = host.split("mongodb://")
###

#password = "adsl;dfjk"


class TestMongo(unittest.TestCase):

  def connect(self):
    try:
      myclient = MongoClient("mongodb://{}:{}@{}".format(username,
                                                              password,
                                                              new_host[1]))
      return myclient

    except:
 
      return False


  def test_connect(self):

    """ Simple Connection to Mongo
    """
    self.assertIsInstance(self.connect(), MongoClient) 



  def test_databases(self):

    """ Check for goflow-mock Database
    """

    x = self.connect()

    dbs = x.list_database_names()

    self.assertIn("goflow-mock", dbs)


  def test_collections(self):

    """ Check for existence of collections requred by GoFlows
    """

    x = self.connect()

    db = x["goflow-mock"]

    collections = db.list_collection_names()

    self.assertIn("GoFlows", collections)
    self.assertIn("GoFlowTriggers", collections)
    self.assertIn("GoFlowUsers", collections)
    self.assertIn("GoFlowJobs", collections)

  def test_insert_delete(self):

    """ Try inserting a test document to GoFlows collection
    """

    x = self.connect()

    db = x["goflow-mock"]

    collection = db["GoFlows"]

    #print(collection.find_one({"Test": "true"}))


    print("Inserting test..")
    new_doc = collection.insert_one({"Test": "true"}).inserted_id
    print(new_doc)

    # Check if the new Document was inserted:
    self.assertIsNotNone(collection.find_one({"_id": new_doc}))

    # Delete the new Document
    print("Deleting test...")
    del_doc = collection.delete_one({"_id": new_doc})
    print(del_doc)

    # Check that the doc was deleted.
    self.assertIsNone(collection.find_one({"_id": new_doc}))


# Run the unit tests.
if __name__ == '__main__':
   unittest.main() 
