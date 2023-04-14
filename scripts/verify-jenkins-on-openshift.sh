#!/bin/bash

# This script deploys the jenkins pod with provided image (from CPaas) on an 
# OpenShift cluster & executes verify_jenkins.sh within the newly deployed pod.
set -o pipefail

# Takes input & stores in the respective variables
# 
get_input(){
    read -p "OpenShift's API URL: " API_SERVER
    read -p "OpenShift's User Name (kubeadmin): " USER_NAME
    read -p "OpenShift's User Password: " USER_PASSWORD
    read -p "Jenkins Image built by CPaas: " JENKINS_IMAGE  
}

vars(){

    get_input
    IFS=":" read -r IMG TAG <<< ${JENKINS_IMAGE}
    QUAY_IMAGE="quay.io/pipeline-integrations/openshift-ose-jenkins:${TAG}"

    # if username is not passed, pick `kubeadmin` as default user
    if [ -z ${USER_NAME} ]
    then
        USER_NAME="kubeadmin"
    fi
}

# get_image_ready() 
# pulls the brew image locally, tags & pushes it to quay.
# 
get_image_ready(){
    echo -e "\n-------------------------------------------------------\n>INFO || Pulling the Jenkins Image\n"
    IMAGE_ID=$(podman pull ${JENKINS_IMAGE})
    if [ -z $IMAGE_ID ]
    then
        echo -e "\n>ERR =======> Image Pull Failed\n"
        exit 1
    else
        podman tag ${JENKINS_IMAGE} ${QUAY_IMAGE}
        echo -e "\n-------------------------------------------------------\n>INFO || Pushing the image to quay.io/pipeline-integrations\n"
        podman push ${QUAY_IMAGE}
        echo -e "\n-------------------------------------------------------\n>INFO || Image push successful\n"
    fi
}

# deploy_on_openshift()
# Deploys the pod with the provided CPaas image, 
# 
deploy_on_openshift(){

    # ensure oc exists
    oc version > /dev/null
    if [ $? != 0 ]
    then
        echo "\n-------------------------------------------------------> Please install `oc` client to continue. https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/"
        exit 1
    fi

    # if oc exists, continue
    echo -e "\n-------------------------------------------------------\n>INFO || Logging in as " ${USER_NAME} " to " ${API_SERVER} "\n"
    oc login -u ${USER_NAME} -p ${USER_PASSWORD} ${API_SERVER}
    oc delete project test-jenkins-latest
    oc new-project test-jenkins-latest
    oc import-image --confirm jenkins-image --from=${QUAY_IMAGE}
    oc new-app jenkins-ephemeral -p NAMESPACE=$(oc project -q) -p JENKINS_IMAGE_STREAM_TAG=jenkins-image:latest

    # wait for jenkins pod to switch to 'Ready' & 'Running` state:
    sleep 10
    echo -e "\n-------------------------------------------------------\n>INFO || Waiting for pod to be ready\n"
    OUTPUT=$(oc wait pod --for=condition=Ready -l name=jenkins,deploymentconfig=jenkins --timeout=120s)

    if [ -z "${OUTPUT}" ]
    then
        exit 1
    else
        IFS=" " read -ra ELEMENTS <<< ${OUTPUT}
        IFS="/" read -r POD POD_NAME <<< ${ELEMENTS[0]}
    fi

    echo -e "\n-------------------------------------------------------\n>INFO || Pod ${POD_NAME} is available now\n"    
}

# Creates '/tmp/download_script.sh' script which is executed within the pod.
# The `download_script.sh` downloads & executes the verify_jenkins.sh inside the pod. 
# 
execute_verification_script(){

    # Create a script, that'll execute the `verify-jenkins.sh` within the pod
    read -e -p "Enter the relevant commit's SHA: " COMMIT_SHA
    cat > /tmp/download_script.sh << EOL
curl https://raw.githubusercontent.com/openshift/jenkins/master/scripts/verify-jenkins.sh | SHA=$COMMIT_SHA sh
EOL

    # Run jenkins verification script inside the pod & redirect the output to a file.
    echo -e "\n-------------------------------------------------------\n>INFO || Running the verify-jenkins.sh within the pod\n"
    oc exec -i $POD_NAME -- bash -s < /tmp/download_script.sh &> /tmp/result.out

    # Check if "All tests succeeded" string exists in the file.
    if [ "$(grep "All tests succeeded" /tmp/result.out)" == "" ]
    then
        grep "Jenkins startup" /tmp/result.out -A 3
        echo -e "\n>>> ERR || Check /tmp/result.out on local machine for details\n"
        exit 1
    else
        echo -e "=== All tests succeeded ==="
    fi
}

main(){
    vars
    get_image_ready
    deploy_on_openshift
    execute_verification_script
}

main