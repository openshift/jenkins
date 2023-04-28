#!/bin/bash

printf "Checking for plugin vulnerabilities in 2/contrib/openshift/bundle-plugins.txt\n"
java -jar 2/contrib/openshift/jenkins-plugin-manager.jar --view-security-warnings --plugin-file 2/contrib/openshift/bundle-plugins.txt --jenkins-version $(cat 2/contrib/openshift/jenkins-version.txt)