#!/usr/bin/env bash
docker build -t hub.jucaicat.net/openshift/jenkins-slave-gradle-alpine .
docker push hub.jucaicat.net/openshift/jenkins-slave-gradle-alpine