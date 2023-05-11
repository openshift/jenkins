#!/bin/bash

# Assumes the following directory structure exists:
# ~/projects/github.com/openshift/jenkins
# ~/projects/github.com/openshift/jenkins-sync-plugin
# ~/projects/github.com/openshift/jenkins-client-plugin
# ~/projects/github.com/openshift/jenkins-openshift-login-plugin

if [ -z "$QUAY_USER" ]; then
		printf "You MUST supply a QUAY_USER to push the image to\n"
		printf "Usage: QUAY_USER=<username> sh build.sh \n"
		exit 1
fi


# Create the directory to copy the plugin artifacts into
if [ ! -d jpi ]; then
	mkdir jpi
else
	rm -rf jpi/*
fi

# Build the base Jenkins agent image and tag it as quay.io/<username>/origin-jenkins-agent:latest
pushd ~/projects/github.com/openshift/jenkins || return
podman build slave-base -f slave-base/Dockerfile.rhel8 -t "quay.io/${QUAY_USER}/jenkins-agent-base:latest"
popd || return
podman push "quay.io/${QUAY_USER}/jenkins-agent-base:latest"

# Build the base Jenkins server image and tag it as quay.io/<username>/origin-jenkins:latest
pushd ~/projects/github.com/openshift/jenkins || return
podman build 2 -f 2/Dockerfile.rhel8 -t "quay.io/${QUAY_USER}/jenkins:server-base"
popd || return

if [ -z "$SKIP_BUILD_PLUGINS" ]; then
# Directory to name mapping for Jenkins plugins
PLUGINS=(
	"jenkins-sync-plugin:openshift-sync"
	"jenkins-client-plugin:openshift-client"
	"jenkins-openshift-login-plugin:openshift-login"
)

# Build each plugin and copy the artifact to the jpi folder
for plugin in "${PLUGINS[@]}"; do
	DIRECTORY=${plugin%%:*}
    NAME=${plugin#*:}

	pushd "${HOME}/projects/github.com/openshift/${DIRECTORY}" || exit
	mvn clean package
	popd || exit
	cp "${HOME}/projects/github.com/openshift/${DIRECTORY}/target/${NAME}.hpi" "./jpi/${NAME}.jpi"
done

# Create the needed Containerfile
cat > Containerfile << EOL
FROM quay.io/${QUAY_USER}/jenkins:server-base
USER root
COPY ./jpi /opt/openshift/plugins
EOL

# Build the quay.io/<username>/jenkins:latest image and push it to quay.io
podman build . -f Containerfile -t "quay.io/${QUAY_USER}/jenkins:latest"
else
podman tag "quay.io/${QUAY_USER}/jenkins:server-base" "quay.io/${QUAY_USER}/jenkins:latest"
fi
podman push "quay.io/${QUAY_USER}/jenkins:latest"

if [ -z "$NO_CLEANUP" ] || [ -n "$SKIP_BUILD_PLUGINS" ];then
	rm -rf jpi
	rm Containerfile
fi
