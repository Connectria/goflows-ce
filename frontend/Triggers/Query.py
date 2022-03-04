from models import *
import logging

def listOfTriggers():
  triggers = GoFlowTriggers.objects()

  outTriggers = {"triggers": []}
  for trigger in triggers:
    outTriggers["triggers"].append(
      {
        "triggerId": trigger.triggerId,
        "name": trigger.triggerName,
        "active": trigger.triggerActive,
        "createdOn": trigger.createdOn,
        "updatedOn": trigger.updatedOn
      }
    )

  return jsonify(outTriggers)

def queryTriggers(args_hash):
  t_return = []
  for arg in args_hash:

    if "." in arg:
      p = arg.split('.')
      logging.info("Split Arg: {}".format(p))
      p[0] = GoFlowTriggers.getUpdateFields()[p[0]]
      key_feild = '.'.join(p)
    else:
      key_feild = GoFlowTriggers.getUpdateFields()[arg]

    if args_hash[arg].lower() == 'true':
      right_of_query = True
    elif args_hash[arg].lower() == 'false':
      right_of_query = False

    if args_hash[arg].isnumeric():
      right_of_query = int(args_hash[arg])
    else:
      right_of_query = args_hash[arg]

    if(right_of_query == "$exists"):
      triggers = GoFlowTriggers.objects(
        __raw__= { key_feild : { right_of_query: True } }
      )
    else:
      triggers = GoFlowTriggers.objects(
        __raw__={ key_feild : right_of_query }
      )

    if key_feild == 'triggerActive':


      if right_of_query.lower() == 'true':
        triggers = GoFlowTriggers.objects(triggerActive = True)
      elif right_of_query.lower() == 'false':
        triggers = GoFlowTriggers.objects(triggerActive = False)

    for t in triggers:
      t_return.append(t.getInfo())

  return t_return
