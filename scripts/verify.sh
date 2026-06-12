#!/bin/bash
set -euo pipefail

PLUGIN_MANAGER_VERSION="2.14.0"
PLUGIN_MANAGER_JAR="2/contrib/openshift/jenkins-plugin-manager.jar"
PLUGIN_MANAGER_URL="https://github.com/jenkinsci/plugin-installation-manager-tool/releases/download/${PLUGIN_MANAGER_VERSION}/jenkins-plugin-manager-${PLUGIN_MANAGER_VERSION}.jar"

if [ ! -f "$PLUGIN_MANAGER_JAR" ]; then
  printf "Downloading jenkins-plugin-manager %s ...\n" "$PLUGIN_MANAGER_VERSION"
  curl -sSfL -o "$PLUGIN_MANAGER_JAR" "$PLUGIN_MANAGER_URL"
fi

PLUGIN_DIR=$(mktemp -d)
trap 'rm -rf "$PLUGIN_DIR"' EXIT

printf "Checking for plugin vulnerabilities in 2/contrib/openshift/bundle-plugins.txt\n"
java -jar "$PLUGIN_MANAGER_JAR" --view-security-warnings -d "$PLUGIN_DIR" --plugin-file 2/contrib/openshift/bundle-plugins.txt --jenkins-version "$(cat 2/contrib/openshift/jenkins-version.txt)"