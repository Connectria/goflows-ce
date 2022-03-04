DEV_DUMMY_FLOW="db1ec516-99be-4bda-bb7b-190ba956d1b9" # Dummy Flow
STAGE_DUMMY_FLOW="f008c5ee-20a5-40c7-98d1-75b1a40476ce" # Dummy Flow

cat <<-EOF | http --verify=no POST ${APIURL}/api/triggers ${APIKEY}
{
    "active": true,
    "flowIds": [
        "${STAGE_DUMMY_FLOW}"
    ],
    "inputs": [
        {
            "inputName": "label",
            "inputValue": "This is just a test"
        }

    ],
    "name": "",
    "triggerIds": [],
    "triggerLogic": {
        "temporal-rule": {
            "cron-expression": {
                "expression": "CRON_TZ=America/Chicago 0 19 * * *",
                "repeat-until": 1643608800
            }  
        }
    }
}
EOF


