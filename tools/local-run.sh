#!/bin/bash

# Asssumes you have a default service account type setup
CATTLE_AGENT_LOCALHOST_REPLACE="10.0.3.2" CATTLE_ACCESS_KEY="service" CATTLE_SECRET_KEY="servicepass" CATTLE_URL=http://192.168.11.212:8080/v1 ./rancher-mesos-scheduler
