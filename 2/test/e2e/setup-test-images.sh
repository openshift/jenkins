#!/bin/bash
#
# Imports the Jenkins and Jenkins Agent Base images into the openshift
# namespace ImageStreams so they are available to the e2e tests and to
# any template-based deployments on the cluster.
#
# Usage:
#   JENKINS_IMAGE=registry.redhat.io/ocp-tools-4/jenkins-rhel9:v4.17.0 \
#   JENKINS_AGENT_BASE_IMAGE=registry.redhat.io/ocp-tools-4/jenkins-agent-base-rhel9:v4.17.0 \
#   JENKINS_IMAGE_OLD=registry.redhat.io/ocp-tools-4/jenkins-rhel9:v4.17.0-1750848396 \
#   ./setup-test-images.sh
#
set -euo pipefail

JENKINS_IMAGE="${JENKINS_IMAGE:?JENKINS_IMAGE must be set}"
JENKINS_AGENT_BASE_IMAGE="${JENKINS_AGENT_BASE_IMAGE:?JENKINS_AGENT_BASE_IMAGE must be set}"
JENKINS_IMAGE_OLD="${JENKINS_IMAGE_OLD:?JENKINS_IMAGE_OLD must be set}"

echo "=== Setting up test images in openshift namespace ImageStreams ==="
echo "  Jenkins image       : ${JENKINS_IMAGE}"
echo "  Agent base image    : ${JENKINS_AGENT_BASE_IMAGE}"
echo "  Old Jenkins image version: ${JENKINS_IMAGE_OLD}"
echo

# ------ Jenkins ImageStream (jenkins:2 and jenkins:latest) ------
echo "Tagging ${JENKINS_IMAGE} -> openshift/jenkins:2"
oc tag --source=docker "${JENKINS_IMAGE}" openshift/jenkins:2

echo "Tagging ${JENKINS_IMAGE} -> openshift/jenkins:latest"
oc tag --source=docker "${JENKINS_IMAGE}" openshift/jenkins:latest

# ------ Jenkins Agent Base ImageStream (jenkins-agent-base:latest) ------
echo "Tagging ${JENKINS_AGENT_BASE_IMAGE} -> openshift/jenkins-agent-base:latest"
oc tag --source=docker "${JENKINS_AGENT_BASE_IMAGE}" openshift/jenkins-agent-base:latest

# ------ Old Jenkins ImageStream (jenkins:oldversion) for upgrade tests ------
echo "Tagging ${JENKINS_IMAGE_OLD} -> openshift/jenkins:oldversion"
oc tag --source=docker "${JENKINS_IMAGE_OLD}" openshift/jenkins:oldversion

# ------ Wait for imports ------
echo
echo "Waiting for image imports to resolve..."

wait_for_istag() {
    local istag="$1"
    local attempts=24  # 24 x 5s = 2 minutes
    for i in $(seq 1 ${attempts}); do
        ref=$(oc get istag "${istag}" -n openshift \
              -o jsonpath='{.image.dockerImageReference}' 2>/dev/null || true)
        if [ -n "${ref}" ]; then
            echo "  ${istag} -> ${ref}"
            return 0
        fi
        sleep 5
    done
    echo "  ERROR: ${istag} did not resolve after 2 minutes"
    return 1
}

wait_for_istag "jenkins:2"
wait_for_istag "jenkins:latest"
wait_for_istag "jenkins-agent-base:latest"
wait_for_istag "jenkins:oldversion"

echo
echo "=== Image setup complete ==="
