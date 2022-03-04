import pika
import configparser
import sys
import logging
from goflows_frontend import app

class MqMessage():

  def __init__(self):
    mqConfig = configparser.ConfigParser()
    mqConfig.read(app.root_path + '/params.ini')

    credentials = pika.PlainCredentials(
      mqConfig['MQProducer']['user'],
      mqConfig['MQProducer']['password']
    )

    try:
      self.connection = pika.BlockingConnection(
        pika.ConnectionParameters(
          host= mqConfig['MQProducer']['host'],
          virtual_host=mqConfig['MQProducer']['virtualhost'],
          credentials = credentials
        )
      )
    except:
      print("!!! Error connecting to RabbitMQ", sys.exc_info()[0])
      self.connection = False

    self.exchange = mqConfig['MQProducer']['exchange']
    self.aoee_exchange = mqConfig['MQProducer']['aoee_exchange']


  def mqsend(self, message):
    pass
    if self.connection:
      channel = self.connection.channel()

      # send to event bridge 
      channel.exchange_declare(exchange=self.exchange, exchange_type='fanout')
      try:
        channel.basic_publish(exchange = self.exchange, routing_key='', body=message)
        print(" [x] Sent to event bridge:  %r" % message)
      except:
        self.connection.close()
        print("ERROR: Sending message to event bride:", message)
        return False

      # send to auto-ops event enrichment service
      channel.exchange_declare(exchange=self.aoee_exchange, exchange_type='fanout')
      try:
        channel.basic_publish(exchange = self.aoee_exchange, routing_key='', body=message)
        print(" [x] Sent to auto-ops enrichment service:  %r" % message)
      except:
        self.connection.close()
        print("ERROR: Sending message to auto-ops enrichment service:", message)
        return False

      self.connection.close()
      return True

    else:
      return False
