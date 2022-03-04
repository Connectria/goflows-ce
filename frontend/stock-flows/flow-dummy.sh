# - Create a Flow
#   This has the correct regexs to assign vars
#!/bin/bash

cat <<-EOF | http --verify=no POST ${APIURL}/api/flows ${APIKEY}
{
    "flowDescription": "Does nothing, but is a good experment for rules and other tests",
    "flowDocument": {
        "funcName": "Dummy"
    },
    "flowInputs": [],
    "flowName": "Dummy Flow",
    "flowOutputs": [],
    "flowShortDescription": "Delete Me When Your done testing"
}
EOF
