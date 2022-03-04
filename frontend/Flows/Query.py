from models import *
import logging

def listOfFlows():

    flows = GoFlows.objects()
      # Let's do some light-weight listification
    outFlows = {"flows": []}
    for flow in flows:
      outFlows["flows"].append(
            { "flowId": flow.flowId,
              "flowName": flow.flowName,
              "createdOn": flow.createdOn,
              "updatedOn": flow.updatedOn }
      )

    return jsonify(outFlows)

def queryFlows(args_hash):

    t_return = []
    for arg in args_hash:
        if "." in arg:
          p = arg.split('.')
          logging.info("Split Arg: {}".format(p))
          p[0] = GoFlows.getUpdateFields()[p[0]]
          key_feild = '.'.join(p)
        else:
          key_feild = GoFlows.getUpdateFields()[arg]

        if args_hash[arg].lower() == 'true':
          right_of_query = True
        elif args_hash[arg].lower() == 'false':
          right_of_query = False

        if args_hash[arg].isnumeric():
          right_of_query = int(args_hash[arg])
        else:
          right_of_query = args_hash[arg]

        if(right_of_query == "$exists"):
          flows = GoFlows.objects(
            __raw__= { key_feild : { right_of_query: True } }
          )
        else:
          flows = GoFlows.objects(
            __raw__={ key_feild : right_of_query }
          )

        for t in flows:
          t_return.append(t.getInfo())

    return jsonify( {"query": args_hash, "flows": t_return } )

