#!/bin/bash

cp ../2/contrib/openshift/base-plugins.txt .
docker build -f ./Dockerfile-plugin-lister -t openshift/jenkins-with-local-plugin-list:latest .
rm base-plugins.txt
