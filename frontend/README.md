# GoFlows API Frontend 
The GoFlows API frontend serves as orchestration layer for the GoFlows service.  The API provides CRUD for flows, triggers and jobs.  The service makes backend calls to goflows-scheduler when a trigger is submitted.  There is also a flows - runNow endpoint that allows immedidate execution of the flow.  After a trigger has been submitted, the backend service uses the jobs endpoint to keep a running status of the status of the flow in the jobs endpoint. 

<p><strong>Full API Documentation can be found at:</strong>API documentation is compiled using <i>/static/swagger.yaml</i> or <i>/static/swagger.json</i>.  Both files include examples
that can be viewed at <url>/swagger

<p>The <i>params.ini</i> contains the configuration parameters for RabbitMQ, MongoDB, and the URL for the backend scheduler.  TBD Example ini file (really should replace with a .env file anyhow)</i></p>

#Release Notes
2021-12-08 - Added Stringifyed Inputs to flows and triggers. Now when you attempt to run a flow or trigger and pass as an input value a valid JSON string it will automatically be converted to stringifyied. 


<h4>Deployment</h4>
<pre>
####################
 DB SETUP
####################
1. Log into Mongo Server:

$ mongo --authenticationDatabase "admin" -u "admin" -p
$ use admin;
$ db.createUser({ user: "frontend", pwd: "<password>", roles: [{ role: "readWrite", db: "goflow-mock"}, { role: "readWrite", db: "godata-models" }] })

####################
 MQ SETUP
####################

1. Log into RabbitMQ Server

Download rabbitadmin tool:
$ wget http://localhost:15672/cli/rabbitmqadmin
$ chmod +x rabbitmqadmin
$ mv rabbitmqadmin /usr/local/bin/
$ rabbitmqadmin -u full_access declare exchange --vhost=GoFlows name=GoFlow.CRUD.Feed type=fanout -p <password> durable=false

####################
 APP Deployment
####################

$ sudo yum -y install epel-release git python3
$ sudo yum -y install nginx
$ sudo systemctl enable nginx

$ cd /opt/
$ git clone git@github.com:Connectria/goflows_frontend.git
$ cd goflows_frontend
$ pip install -r requirements.txt
$ cp -ar Notes/Deployment/systemd/* /etc/systemd/system/
$ sudo systemctl daemon-reload
$ cp Notes/Deployment/params.ini.example /opt/goflows_frontend/params.ini
$ cp Notes/Deployment/nginx_config /etc/nginx/conf.d/   # Be sure to remove the original server{} block from /etc/nginx/nginx.conf

2. Change passwords/IP's in params.ini file to reflect the environment.
</pre>

<pre>
As of Close of Sprint 9 5/29/2021 Mergeing Support for mulitiaple types of triggers mainly event and task
Task is a cron or at list
	- Some validation was pushed deeper and will have to be resurfaced, mainly some cron and at checks
Event is one triggered by an external system, in this case Opsgenie
	- At this time there are no checks for syntax of the "event-rules"

</pre>
