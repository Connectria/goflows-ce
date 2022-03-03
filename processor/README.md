# README.md

Engine for processing GoFlows received from **goflows-scheduler** or **opsgenie-reader** Logs are obtained using the **goflows-api** program

Please note that this is a work in progress and subject to **frequent changes and updates**.

## Configuration

Review the `sample.env` file and follow steps listed there.

## Usage

```bash
$ ./goflows-processor 
NAME:
   goflows-processor - CLI to manage the GoFlows processor daemon

USAGE:
   goflows-processor [global options] command [command options] [arguments...]

COMMANDS:
   config, c         List configuration
   list, l           List compiled goflows
   process, p, proc  Process OpsGenie evenit with messageID with compiled goflows (live)
   status            Check the status of the processor-daemon
   start             Start the processor-daemon
   stop              Stop the processor-daemon
   help, h           Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

### Daemon

```bash
$ ./goflows-processor status 
daemonCmd(status) - daemon is not running

$ ./goflows-processor start
daemonCmd(start) - Daemon process ID is '23286'
savePID(23286) - saved process ID to '/var/tmp/goflows-processor.pid'

$ ./goflows-processor status
daemonCmd(status) - Daemon process ID is '23286'
vince@PC0XW4CM:~/projects/src/goflows-processor$ date
Tue Aug 11 11:43:08 CDT 2020

$ ./goflows-processor stop 
daemonCmd(stop) - Killing process ID '23286'
daemonCmd(stop) - Killed process ID '23286'
```

## API

*Swagger version at http://10.0.1.169:8082/swagger/index.html*

### verify API is running

```bash
$ http localhost:8082/ping 

HTTP/1.1 200 OK
Content-Length: 15
Content-Type: application/json; charset=utf-8
Date: Thu, 11 Feb 2021 22:22:47 GMT

{
    "ping": "pong"
}
```

### Pull flow log for specific triggerId (i.e. GoFlow triggerId alert)

```bash
http GET localhost:8082/api/history?triggerId=2fbc94f9-5b04-425d-9d78-d6faca2201d9
```

### Pull flow log for specific jobID (i.e., GoFlow jobID)

```bash
$ http GET localhost:8082/api/history?jobID=9b3b673f-ec87-ec71-036d-5b3315b6e09d

```

### Pull flow log for specific eventAlertID (i.e. OpsGenie alert)

```bash
http GET localhost:8082/api/history?eventAlertID=f83352a8-3e83-49fa-8413-584f8c54c361-1607922831803 
```

### Pull flow log for specific messageID (i.e. SQS MessageID)

```bash
http GET localhost:8082/api/history?messageID=d420f422-35b7-459d-abb2-14497e0b94c7
```

### Retrieve flow log history between specific time

* To get logs "up to a specific time", omit the "startTime" parameter"
* To get logs "from a specific time", omit the "endTime" parameter"
* Time is specified in "UNIX time"

```bash
http GET localhost:8082/api/history?startTime=1607923260\&endTime=1607923300
```

### List compiled GoFlows

```bash
 http get localhost:8082/api/listFuncs

HTTP/1.1 200 OK
Content-Length: 383
Content-Type: application/json; charset=utf-8
Date: Thu, 11 Feb 2021 22:26:16 GMT

{
    "funcs": {
        "eventFuncs": [
            {
                "funcName": "MessageMonitorParsedMessages"
            },
            {
                "funcName": "MimixAssurePortalIssue"
            },
            {
                "funcName": "MimixCatchAll"
            },
            {
                "funcName": "MimixExpiringAllCustomers"
            },
            {
                "funcName": "MimixLicenseExpiredAllCustomers"
            },
            {
                "funcName": "WindowsDiskAlerts"
            }
        ],
        "taskFuncs": [
            {
                "funcName": "dailychuck"
            },
            {
                "funcName": "TestAzureWebhook"
            },
            {
                "funcName": "TestCallBack"
            },
            {
                "funcName": "TestEmail"
            }
        ]
    }
}
```

### List call backs

```bash
$ http GET localhost:8082/api/listCallBacks

HTTP/1.1 200 OK
Content-Length: 252
Content-Type: application/json; charset=utf-8
Date: Wed, 27 Jan 2021 01:44:45 GMT

{
    "callBacks": {
        "342a6361-6aaf3c3a": {
            "callBackData": {
                "inputs": [
                    {
                        "inputName": "Item1",
                        "inputValue": "The cow says moo!"
                    },
                    {
                        "inputName": "Item2",
                        "inputValue": "The pig says oink!"
                    }
                ]
            },
            "callBackRef": "testCallBack",
            "jobID": "52440816-e82a-376e-43d0-e65755369dcd"
        }
    }
}

$ http GET localhost:8082/api/listCallBacks

HTTP/1.1 200 OK
Content-Length: 16
Content-Type: application/json; charset=utf-8
Date: Wed, 27 Jan 2021 01:50:10 GMT

{
    "callBacks": {}
}
```

### List trigger event rules (in cache)
```bash
 http get :8082/api/listTriggerEventRules 
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Fri, 21 May 2021 19:45:54 GMT
Transfer-Encoding: chunked

{
    "TTL": 300,
    "lastUpdate": 1621626343,
    "triggers": [
        {
            "active": true,
            "createdOn": "Thu, 20 May 2021 20:23:38 GMT",
            "flowIds": [
                "86e0bfff-c67e-413b-b0ac-ad23049444dd",
                "6ca72d74-13a6-4216-bc68-5672be83a922"
            ],
            "inputs": [],
            "name": "Call log trigger example for Vince test alert in OpsGenie",
            "triggerId": "a3deabfb-9868-4fdd-a6aa-8402cf8da217",
            "triggerIds": [],
            "triggerLogic": {
                "event-rules": [
                    {
                        "exact-match": {
                            "EventData/alert/extraProperties/CustomerID": "XXXXX"
                        }
                    },
                    {
                        "exact-match": {
                            "EventData/alert/extraProperties/DeviceID": "42995"
                        }
                    },
                    {
                        "set-var-from-source-value": {
                            "customerID": "EventData/alert/extraProperties/CustomerID"
                        }
                    },
                    {
                        "set-var-from-source-value": {
                            "deviceID": "EventData/alert/extraProperties/DeviceID"
                        }
                    },
                    {
                        "set-var-from-source-value": {
                            "eventTime": "EventData/alert/createdAt"
                        }
                    },
                    {
                        "set-vars-from-source-regex": {
                            "group": "2",
                            "source": "EventData/alert/alias",
                            "vars": [
                                {
                                    "num": "(.*?/){3,3}([^/]*)"
                                },
                                {
                                    "jobID": "(.*?/){4,4}([^/]*)"
                                },
                                {
                                    "msgID": "(.*?/){5,5}([^/]*)"
                                },
                                {
                                    "msg": "(.*?/){6}(.*[^/])"
                                }
                            ]
                        }
                    }
                ],
                "temporal-rule": {
                    "cron-expression": {}
                },
                "triggerSubType": "Opsgenie",
                "triggerType": "Event"
            },
            "updatedOn": "Fri, 21 May 2021 13:15:40 GMT"
        }
    ]
}
```

### change the event trigger cache TTL without restarting the daemon 

*When updating the TTL, the events are immediately updated and the new TTL clock starts. Whatever is set in the `.env` file will be used when the daemon is restarted.*

```bash 
$ http put :8082/api/setTriggerEventTTL?TTL=120
HTTP/1.1 200 OK
Content-Length: 46
Content-Type: application/json; charset=utf-8
Date: Fri, 21 May 2021 20:08:55 GMT

{
    "TTL": 120,
    "message": "event triggers updated"
}

$ http put :8082/api/setTriggerEventTTL?TTL=12
HTTP/1.1 400 Bad Request
Content-Length: 39
Content-Type: application/json; charset=utf-8
Date: Fri, 21 May 2021 19:59:31 GMT

{
    "error": "TTL must be greater than 30"
}
```

### evaluate an event against rule 

*To see how an existing event (i.e., OpsGenie alert) would be evaluated against the current event rules/triggers. messageID is the SQSMessageID of the event* 

```bash 
$ http GET localhost:8082/api/validateRules?messageID=cd6d34c7-53ba-44b2-bd3d-6bfb7f7d9cec
HTTP/1.1 400 Bad Request
Content-Length: 125
Content-Type: application/json; charset=utf-8
Date: Thu, 12 Aug 2021 22:41:08 GMT

{
    "error": "finished event rules analysis; no match",
    "messageID": "cd6d34c7-53ba-44b2-bd3d-6bfb7f7d9cec",
    "status": "not-matched"
}

$ http GET localhost:8082/api/validateRules?messageID=753de5dd-cce0-448f-a111-d1a0fa5d7c1b
HTTP/1.1 200 OK
Content-Length: 322
Content-Type: application/json; charset=utf-8
Date: Thu, 12 Aug 2021 22:47:16 GMT

{
    "eventAlertID": "9bb98b51-6f59-487d-a158-dfc81c864584-1628725820052",
    "messageID": "753de5dd-cce0-448f-a111-d1a0fa5d7c1b",
    "setVars": {
        "customerID": "XXXXX",
        "deviceID": "6567",
        "eventTime": "{\"$numberLong\":\"1628725820052\"}"
    },
    "status": "matched",
    "triggerID": "cd472aa3-6133-43ba-9c9a-6fbd8c08643e",
    "triggerName": "Vince DEV test"
}
```

### Retrieve auth history between specific time

* To get authorizations "up to a specific time", omit the "startTime" parameter"
* To get authorizations "from a specific time", omit the "endTime" parameter"
* Time is specified in "UNIX time"

```bash
http GET localhost:8082/api/authHistory?startTime=1607923260\&endTime=1607923300
```
