# goflows-scheduler

Cron-like scheduler for scheduling goflow execution by goflows-processor.

## Usage

```bash
$ ./goflows-scheduler
NAME:
   goflows-scheduler - CLI to manage the GoFlows scheduler daemon

USAGE:
   goflows-scheduler [global options] command [command options] [arguments...]

COMMANDS:
   config, c  List configuration
   status     Check the status of the scheduler-daemon
   start      Start the scheduler-daemon
   stop       Stop the scheduler-daemon
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

## API

*Swagger version at http://10.0.1.169:8083/swagger/index.html* 

### Verify API is running

```bash
$ http localhost:8083/ping 
HTTP/1.1 200 OK
Content-Length: 15
Content-Type: application/json; charset=utf-8
Date: Wed, 10 Feb 2021 17:36:30 GMT

{
    "ping": "pong"
}
```

### Schedule a GoFlow (cron style)

```bash
$ cat test-sched-email.sh 
TIME=$(expr $(date +%s) + 30)


cat <<EOF | http POST http://localhost:8083/api/add 
{ 
    "triggerId": "79304e95-c077-4ffd-92a3-0d133a08bfd9", 
    "triggerName": "test call back trigger", 
    "flowIDs": [ 
        { 
            "FlowID": "79304e95-c077-4ffd-92a3-0d133a08bfd9", 
            "funcName": "TestEmail", 
            "flowDescription": "Send test email", 
            "flowName": "Test Email " 
        }
    ],
    "at-list": [ 
        ${TIME}
    ], 
    "inputs": [
        { 
            "inputName": "payloadTo", 
            "inputValue": "example@connectria.com"
        }, 
        { 
            "inputName": "payloadCc", 
            "inputValue": ""
        },
        { 
            "inputName": "payloadStuff", 
            "inputValue": "Twenty percent of all input forms filled out by people contain bad data. ~Dennis Ritchie"
        }
    ]
}
EOF

 ./test-cron-email.sh 
HTTP/1.1 200 OK
Content-Length: 300
Content-Type: application/json; charset=utf-8
Date: Wed, 10 Feb 2021 17:58:26 GMT

{
    "dutyCycle": "59 21 * * *",
    "flowIDs": [
        {
            "flowDescription": "Send test email",
            "flowID": "79304e95-c077-4ffd-92a3-0d133a08bfd9",
            "flowName": "Test Email ",
            "funcName": "TestEmail"
        }
    ],
    "repeat-until": 1613009905,
    "schedulerJobID": 50,
    "status": "added to scheduler",
    "triggerId": "79304e95-c077-4ffd-92a3-0d133a08bfd9"
}
```

### Schedule a GoFlow (at job style)

```bash
$ cat test-sched-email.sh 

TIME=$(expr $(date +%s) + 30)


cat <<EOF | http POST http://localhost:8083/api/add 
{ 
    "triggerId": "79304e95-c077-4ffd-92a3-0d133a08bfd9", 
    "triggerName": "test call back trigger", 
    "flowIDs": [ 
        { 
            "FlowID": "79304e95-c077-4ffd-92a3-0d133a08bfd9", 
            "funcName": "TestEmail", 
            "flowDescription": "Send test email", 
            "flowName": "Test Email " 
        }
    ],
    "at-list": [ 
        ${TIME}
    ], 
    "inputs": [
        { 
            "inputName": "payloadTo", 
            "inputValue": "example@connectria.com"
        }, 
        { 
            "inputName": "payloadCc", 
            "inputValue": ""
        },
        { 
            "inputName": "payloadStuff", 
            "inputValue": "Twenty percent of all input forms filled out by people contain bad data. ~Dennis Ritchie"
        }
    ]
}
EOF

$ ./test-sched-email.sh 
HTTP/1.1 200 OK
Content-Length: 281
Content-Type: application/json; charset=utf-8
Date: Wed, 10 Feb 2021 17:35:26 GMT

{
    "atList": [
        1612978556
    ],
    "flowIDs": [
        {
            "flowDescription": "Send test email",
            "flowID": "79304e95-c077-4ffd-92a3-0d133a08bfd9",
            "flowName": "Test Email ",
            "funcName": "TestEmail"
        }
    ],
    "schedulerJobIDList": [
        48
    ],
    "status": "added jobs to scheduler",
    "triggerId": "79304e95-c077-4ffd-92a3-0d133a08bfd9"
}
```

### Immediately run a GoFlow

```bash
echo '{
    "triggerId": "1234578i-111-111-1111",
    "triggerName": "nightly chuck trigger",
    "flowIDs": [
        {
            "flowID": "1234567889",
            "funcName": "dailychuck",
            "flowDescription": "Nightly Chuck Norris joke email by cron",
            "flowName": "daily chuck flow"
        }
    ],
    "inputs": [
        {
            "inputName": "payload1",
            "inputValue": "example@connectria.com"
        },
        {
            "inputName": "payload2",
            "inputValue": "joe@connectria.com; shean@connectria.com"
        }
    ]
}' | http http://localhost:8083/api/runNow

HTTP/1.1 200 OK
Content-Length: 238
Content-Type: application/json; charset=utf-8
Date: Wed, 10 Feb 2021 17:39:30 GMT

{
    "flowIDs": [
        {
            "flowDescription": "Nightly Chuck Norris joke email by cron",
            "flowID": "1234567889",
            "flowName": "daily chuck flow",
            "funcName": "dailychuck"
        }
    ],
    "status": "Sent to goflows-processor for execution",
    "triggerId": "1234578i-111-111-1111"
}
```

### List all scheduled jobs

* one can use "?triggerId=[id]" or "?schedulerJobID=[id]" to list specific jobs*

```bash
$ http GET http://localhost:8083/api/list
HTTP/1.1 200 OK
Content-Length: 665
Content-Type: application/json; charset=utf-8
Date: Wed, 10 Feb 2021 17:43:29 GMT

{
    "scheduler-list": [
        {
            "flowIDs": [
                {
                    "flowDescription": "Send test email",
                    "flowID": "79304e95-c077-4ffd-92a3-0d133a08bfd9",
                    "flowName": "Test Email ",
                    "funcName": "TestEmail"
                }
            ],
            "inputs": [
                {
                    "inputName": "payloadTo",
                    "inputValue": "example@connectria.com"
                },
                {
                    "inputName": "payloadCc",
                    "inputValue": ""
                },
                {
                    "inputName": "payloadStuff",
                    "inputValue": "Twenty percent of all input forms filled out by people contain bad data. ~Dennis Ritchie"
                }
            ],
            "next": 1612979032,
            "nextHuman": "2021-02-10 17:43:52.81118459 +0000 UTC",
            "prev": -62135596800,
            "prevHuman": "0001-01-01 00:00:00 +0000 UTC",
            "schedulerJobID": 49,
            "triggerId": "79304e95-c077-4ffd-92a3-0d133a08bfd9",
            "triggerName": "test call back trigger"
        }
    ]
}

 http GET http://localhost:8083/api/list?triggerId=79304e95-c077-4ffd-92a3-0d
HTTP/1.1 200 OK
Content-Length: 760
Content-Type: application/json; charset=utf-8
Date: Wed, 10 Feb 2021 17:58:52 GMT

{
    "scheduler-list": [
        {
            "dutyCycle": "59 21 * * *",
            "flowIDs": [
                {
                    "flowDescription": "Send test email",
                    "flowID": "79304e95-c077-4ffd-92a3-0d133a08bfd9",
                    "flowName": "Test Email ",
                    "funcName": "TestEmail"
                }
            ],
            "inputs": [
                {
                    "inputName": "payloadTo",
                    "inputValue": "example@connectria.com"
                },
                {
                    "inputName": "payloadCc",
                    "inputValue": ""
                },
                {
                    "inputName": "payloadStuff",
                    "inputValue": "Twenty percent of all input forms filled out by people contain bad data. ~Dennis Ritchie"
                }
            ],
            "next": 1612994340,
            "nextHuman": "2021-02-10 21:59:00 +0000 UTC",
            "prev": -62135596800,
            "prevHuman": "0001-01-01 00:00:00 +0000 UTC",
            "repeat-until": 1613009905,
            "repeat-untilHuman": "2021-02-11 02:18:25 +0000 UTC",
            "schedulerJobID": 50,
            "triggerId": "79304e95-c077-4ffd-92a3-0d133a08bfd9",
            "triggerName": "test call back trigger"
        }
    ]
}
```

### Remove a scheduled job by scheduledJobID

```bash
$ curl --silent -X DELETE http://172.21.118.185:8083/api/remove/schedulerJobID/1  | jq 

{  "schedulerJobID": 1,
  "status": "removed from goflows-scheduler"
}
```

### Remove a scheduled job by triggerId

```bash
$ curl --silent -X DELETE http://172.21.118.185:8083/api/remove/triggerId/1234578i-111-111-1111 | jq

{  "status": "removed from goflows-scheduler",
  "triggerId": "1234578i-111-111-1111"
}
```

### Retrieve scheduler log history for specific schedulerJobID

```bash
$ http http://172.21.118.185:8083/api/history?schedulerJobID=1

HTTP/1.1 200 OK
Content-Length: 214
Content-Type: application/json; charset=utf-8
Date: Wed, 23 Dec 2020 00:44:19 GMT

{
    "scheduler-history": [
        {
            "LogData": {
                "function": "handleAdd()", 
                "level": "info", 
                "message": "Added 'nightly chuck trigger' at-job to scheduler", 
                "schedulerJobID": "1", 
                "time": 1608684080, 
                "triggerId": "1234578i-111-111-1111"
            }
        }
    ]
}
```

### Retrieve scheduler log history for specific triggerId

```bash
$ http http://172.21.118.185:8083/api/history?triggerId=1234578i-111-111-1111

HTTP/1.1 200 OK
Content-Length: 214
Content-Type: application/json; charset=utf-8
Date: Wed, 23 Dec 2020 00:44:39 GMT

{
    "scheduler-history": [
        {
            "LogData": {
                "function": "handleAdd()", 
                "level": "info", 
                "message": "Added 'nightly chuck trigger' at-job to scheduler", 
                "schedulerJobID": "1", 
                "time": 1608684080, 
                "triggerId": "1234578i-111-111-1111"
            }
        }
    ]
}
```

### Retrieve scheduler log history between specific time

* *To get logs "up to a specific time", omit the "startTime" parameter"*
* *To get logs "from a specific time", omit the "endTime" parameter"*
* *Time is specified in "UNIX time"*

``` bash
[centos@skyline-workflow goflows-scheduler]$ http http://172.21.118.185:8083/api/history?startTime=1607123038\&endTime=1607123212

HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Fri, 04 Dec 2020 23:08:06 GMT
Transfer-Encoding: chunked

{
    "scheduler-history": [
        {
            "LogData": {
                "ip": "172.21.118.185",
                "latency": 0.103789,
                "level": "info",
                "message": "Request",
                "method": "GET",
                "path": "/api/list",
                "status": 200,
                "time": 1607123038,
                "user-agent": "HTTPie/0.9.4"
            }
        },
        {
            "LogData": {
                "function": "handleRemove()",
                "level": "info",
                "message": "Requested jobID('2') not found",
                "time": 1607123054
            }
        },
        {
            "LogData": {
                "ip": "172.21.118.185",
                "latency": 0.918751,
                "level": "info",
                "message": "Request",
                "method": "DELETE",
                "path": "/api/remove/2",
                "status": 200,
                "time": 1607123054,
                "user-agent": "curl/7.29.0"
            }
        },
        {
            "LogData": {
                "function": "handleRemove()",
                "level": "info",
                "message": "removed jobID = '3' from goflows-scheduler",
                "time": 1607123058
            }
        },
        {
            "LogData": {
                "ip": "172.21.118.185",
                "latency": 1.209171,
                "level": "info",
                "message": "Request",
                "method": "DELETE",
                "path": "/api/remove/3",
                "status": 200,
                "time": 1607123058,
                "user-agent": "curl/7.29.0"
            }
        },
        {
            "LogData": {
                "ip": "172.21.118.185",
                "latency": 0.113181,
                "level": "info",
                "message": "Request",
                "method": "GET",
                "path": "/api/list",
                "status": 200,
                "time": 1607123063,
                "user-agent": "HTTPie/0.9.4"
            }
        },
        {
            "LogData": {
                "function": "func()",
                "level": "info",
                "message": "Current time ('2020-12-04 23:05:00 +0000 UTC') exceeds repeat-until ('2020-11-25 03:00:00 +0000 UTC'); removing jobID '4' from scheduler",
                "time": 1607123100
            }
        },
        {
            "LogData": {
                "function": "handleRemove()",
                "level": "info",
                "message": "removed jobID = '4' from goflows-scheduler",
                "time": 1607123110
            }
        },
        {
            "LogData": {
                "ip": "172.21.118.185",
                "latency": 1.436671,
                "level": "info",
                "message": "Request",
                "method": "DELETE",
                "path": "/api/remove/4",
                "status": 200,
                "time": 1607123110,
                "user-agent": "curl/7.29.0"
            }
        },
        {
            "LogData": {
                "function": "handleAdd()",
                "level": "info",
                "message": "added 'testscheduledflow1' to scheduler; jobID = 5, schedulerJobID = 5",
                "time": 1607123120
            }
        },
        {
            "LogData": {
                "ip": "172.21.118.185",
                "latency": 1.568592,
                "level": "info",
                "message": "Request",
                "method": "POST",
                "path": "/api/add",
                "status": 200,
                "time": 1607123120,
                "user-agent": "HTTPie/0.9.4"
            }
        },
        {
            "LogData": {
                "function": "handleRemove()",
                "level": "info",
                "message": "removed jobID = '5' from goflows-scheduler",
                "time": 1607123127
            }
        },
        {
            "LogData": {
                "ip": "172.21.118.185",
                "latency": 1.322414,
                "level": "info",
                "message": "Request",
                "method": "DELETE",
                "path": "/api/remove/5",
                "status": 200,
                "time": 1607123127,
                "user-agent": "curl/7.29.0"
            }
        },
        {
            "LogData": {
                "ip": "172.21.118.185",
                "latency": 3.612561,
                "level": "info",
                "message": "Request",
                "method": "GET",
                "path": "/api/history",
                "status": 200,
                "time": 1607123212,
                "user-agent": "HTTPie/0.9.4"
            }
        }
    ]
}
```

### validate cron-expression 

#### Sample script to validate various scenarios: 

```bash 
# test_cron_expression.sh - test the goflows-scheduler cron expression validator 
#
#!/bin/bash


echo "" 
echo "TEST 1: missing required field" 
cat <<-EOF | http POST :8083/api/validate | jq 
{
    "hello": "world"
}
EOF

echo "" 
echo "TEST 2: invalid format" 
cat <<-EOF | http POST :8083/api/validate | jq 
{
    "dutyCycle": "1 2 3 4 5 6 7 8"
}
EOF

echo "" 
echo "TEST 3: January 1 at 01:01"   
cat <<-EOF | http POST :8083/api/validate | jq 
{
    "dutyCycle": "01 01 01 1 * "
}
EOF

echo "" 
echo "TEST 4: Every Saturday at 13:54"   
cat <<-EOF | http POST :8083/api/validate | jq 
{
    "dutyCycle": "54 13 * * 6"
}
EOF

echo "" 
echo "TEST 5a: Every 5 mins on Monday, Wednesday and Friday"   
cat <<-EOF | http POST :8083/api/validate | jq 
{
    "dutyCycle": "*/5 * * * MON,WED,FRI"
}
EOF

echo "" 
echo "TEST 6: Every 5 mins 10 seconds from the time the job was submitted" 
cat <<-EOF | http POST :8083/api/validate | jq 
{
    "dutyCycle": "@every 5m10s"
}
EOF

echo "" 
echo "TEST 7: every month" 
cat <<-EOF | http POST :8083/api/validate | jq 
{
    "dutyCycle": "@monthly"
}
EOF

echo "" 
echo "TEST 8: garbage should fail" 
cat <<-EOF | http POST :8083/api/validate | jq 
{
    "dutyCycle": "a b 3 1 4"
}
EOF
```

#### Output from sample script to validate various scenarios: 

```bash

 ./test_cron_expression.sh 

TEST 1: missing required field
{
  "error": "missing required field: 'dutyCycle'"
}

TEST 2: invalid format
{
  "dutyCycle": "1 2 3 4 5 6 7 8",
  "error": "expected exactly 5 fields, found 8: [1 2 3 4 5 6 7 8]",
  "status": "invalid"
}

TEST 3: January 1 at 01:01
{
  "dutyCycle": "01 01 01 1 * ",
  "status": "valid"
}

TEST 4: Every Saturday at 13:54
{
  "dutyCycle": "54 13 * * 6",
  "status": "valid"
}

TEST 5a: Every 5 mins on Monday, Wednesday and Friday
{
  "dutyCycle": "*/5 * * * MON,WED,FRI",
  "status": "valid"
}

TEST 6: Every 5 mins 10 seconds from the time the job was submitted
{
  "dutyCycle": "@every 5m10s",
  "status": "valid"
}

TEST 7: every month
{
  "dutyCycle": "@monthly",
  "status": "valid"
}

TEST 8: garbage should fail
{
  "dutyCycle": "a b 3 1 4",
  "error": "failed to parse int from a: strconv.Atoi: parsing \"a\": invalid syntax",
  "status": "invalid"
}

For more formats, see https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format

```
