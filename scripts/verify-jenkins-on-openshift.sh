#!/bin/bash

# This script spins up a container from provided jenkins (CPaas) image 
# & runs `verify_jenkins.sh` in it.
# If all tests pass, it deploys a jenkins pod with the same image on an 
# OpenShift cluster & informs once the pod is available to run e2e tests against it.
# 
# Prerequisites:
# Ensure a login to OpenShift cluster, either via user:password, token
# OR
# KUBECONFIG env var refers to a valid kubeconfig
# 
# Usage: ./verify-jenkins-on-openshift.sh [-i <jenkins_image>] [-s <commit_sha>]
# 

set -o pipefail

# returns the details about command usage
usage() { 
    printf "Usage: ./verify-jenkins-on-openshift.sh [-i <jenkins_image>] [-s <commit_sha>]\n"
    printf "Pre-requisite: Ensure that a user is already logged in the cluster OR have a valid KUBECONFIG populated\n" 
    exit 1; 
}

# Assigns the variable values as per the flags passed.
while getopts "i:s:*:" flag
do
    case "${flag}" in
        i) JENKINS_IMAGE=${OPTARG};;
        s) COMMIT_SHA=${OPTARG};;
        *) usage
    esac
done

# Verifies & sets the required values for variables.
# 
vars(){   

    if [ -z "${JENKINS_IMAGE}" ] || [ -z "${COMMIT_SHA}" ]
    then
        usage
    fi

    # The image pushed will have the respective commit's sha assigned as the tag
    QUAY_IMAGE="quay.io/pipeline-integrations/openshift-ose-jenkins:${COMMIT_SHA}"
}

# deploy_on_openshift()
# Deploys the pod with the provided CPaas image.
# 
deploy_on_openshift(){

    # ensure oc exists
    oc version > /dev/null
    if [ $? != 0 ]
    then
        printf "\n>ERR || Please install 'oc' client to continue. https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/\n"
        printf ">ERR || And/OR ensure that the user has been logged in to the cluster.\n"
        exit 1
    fi

    # if oc exists, continue
    oc delete project test-jenkins-latest
    oc new-project test-jenkins-latest
    oc import-image --confirm jenkins-image --from="${QUAY_IMAGE}"
    oc new-app jenkins-ephemeral -p NAMESPACE="$(oc project -q)" -p JENKINS_IMAGE_STREAM_TAG=jenkins-image:latest

    # wait for jenkins pod to switch to 'Ready' & 'Running` state:
    sleep 10
    printf "\n-------------------------------------------------------\n>INFO || Waiting for pod to be ready\n"
    OUTPUT=$(oc wait pod --for=condition=Ready -l name=jenkins,deploymentconfig=jenkins --timeout=120s)

    if [ -z "${OUTPUT}" ]
    then
        exit 1
    else
        IFS=" " read -ra ELEMENTS <<< "${OUTPUT}"
        IFS="/" read -r POD POD_NAME <<< "${ELEMENTS[0]}"
    fi

    printf "\n-------------------------------------------------------\n>INFO || Pod %s is now available. Any additional tests can be performed against it.\n" "${POD_NAME}"
}

# get_image_ready() 
# pulls the brew image locally, tags & pushes it to quay.
# 
get_image_ready(){
    printf "\n-------------------------------------------------------\n>INFO || Pulling the Jenkins Image\n"
    IMAGE_ID=$(podman pull "${JENKINS_IMAGE}")
    if [ -z "${IMAGE_ID}" ]
    then
        printf "\n>ERR || Image Pull Failed\n"
        exit 1
    else
        podman tag "${JENKINS_IMAGE}" "${QUAY_IMAGE}"
        printf "\n-------------------------------------------------------\n>INFO || Pushing the image to 'quay.io/pipeline-integrations'\n"
        podman push "${QUAY_IMAGE}"
        printf "\n-------------------------------------------------------\n>INFO || Image %s pushed successful\n" "${QUAY_IMAGE}"
    fi

    # now that image is available, deploy a pod on OpenShift cluster using same image.
    deploy_on_openshift 
}

# Creates '/tmp/download_script.sh' script which is executed within the container.
# The `download_script.sh` downloads & executes the verify_jenkins.sh inside the container. 
# 
execute_verification_script(){

    # Create a script, that'll execute the `verify-jenkins.sh` within the pod
    # read -e -p "Enter the relevant commit's COMMIT_SHA: " COMMIT_SHA
    cat > /tmp/download_script.sh << EOL
curl https://raw.githubusercontent.com/openshift/jenkins/master/scripts/verify-jenkins.sh | SHA=${COMMIT_SHA} sh
EOL

    # Run jenkins verification script inside the pod & redirect the output to a file.
    printf "\n-------------------------------------------------------\n>INFO || Running a container with image: %s\n" "${JENKINS_IMAGE}"
    # in case any container with same name exists, deletes the container.
    podman rm "jenkins-$(hostname)" -f
    podman run -dt --name="jenkins-$(hostname)" --entrypoint /bin/bash "${JENKINS_IMAGE}"
    printf "\n-------------------------------------------------------\n>INFO || Executing the verify-jenkins.sh within the %s container (Takes ~100 seconds)\n" "jenkins-$(hostname)"
    podman exec -i "jenkins-$(hostname)" bash -s < /tmp/download_script.sh &> /tmp/result.out &
    for i in {1..100}
    do
        printf "."
        sleep 1
    done

    # If "All tests succeeded", move ahead with deploying the pod on OpenShift, else exit.
    if [ "$(grep "All tests succeeded" /tmp/result.out)" == "" ]
    then
        grep "Jenkins startup" /tmp/result.out -A 4
        printf "\n>>> ERR || Check /tmp/result.out on local machine for details\n"
        exit 1
    else
        printf "=== All tests succeeded ==="
        printf "\n-------------------------------------------------------\n>INFO || Deploying Jenkins on OpenShift\n"
        get_image_ready
    fi
}

main(){
    vars
    execute_verification_script
}

main