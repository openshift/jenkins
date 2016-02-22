#!/bin/sh
#
# This file provides functions to automatically discover suitable image streams
# that the Kubernetes plugin will use to create "slave" pods.
# The image streams has to have label "role" set to "jenkins-slave".
#
# The Jenkins container also need permissions to access the OpenShift API to
# list image streams. You have to run this command to allow that:
#
# $ oc policy add-role-to-user edit system:serviceaccount:ci:default -n ci
#
# (where the 'ci' is the namespace where Jenkins runs)

export DEFAULT_SLAVE_DIRECTORY=/opt/app-root/jenkins
export SLAVE_LABEL="jenkins-slave"

# The project name equals to the namespace name where the container with jenkins
# runs. You can override it by setting the PROJECT_NAME variable.
# If there is no environment variable and this container does not run in
# kubernetes, the default value "ci" is used.
if [ -z "${PROJECT_NAME}" ]; then
  if [ -f "${KUBE_SA_DIR}/namespace" ]; then
    export PROJECT_NAME=$(cat "${KUBE_SA_DIR}/namespace")
  else
    export PROJECT_NAME="ci"
  fi
else
  export PROJECT_NAME
fi

export JENKINS_PASSWORD KUBERNETES_SERVICE_HOST KUBERNETES_SERVICE_PORT
export K8S_PLUGIN_POD_TEMPLATES=""
export PATH=$PATH:${JENKINS_HOME}/.local/bin

function has_service_account() {
  [ -f "${AUTH_TOKEN}" ]
}

if has_service_account; then
  export oc_auth="--token=$(cat $AUTH_TOKEN) --certificate-authority=${KUBE_CA}"
  export oc_cmd="oc --server=$OPENSHIFT_API_URL ${oc_auth}"
fi

# get_imagestream_names returns a list of image streams that match the
# SLAVE_LABEL
function get_is_names() {
  [ -z "$oc_cmd" ] && return
  $oc_cmd get is -n "${PROJECT_NAME}" -l role=${SLAVE_LABEL} -o template --template "{{range .items}}{{.metadata.name}} {{end}}"
}

# convert_is_to_slave converts the OpenShift imagestream to a Jenkins Kubernetes
# Plugin slave configuration.
function convert_is_to_slave() {
  [ -z "$oc_cmd" ] && return
  local name=$1
  local template_file=$(mktemp)
  local template="
  <org.csanchez.jenkins.plugins.kubernetes.PodTemplate>
    <name>{{.metadata.name}}</name>
    <image>{{.status.dockerImageRepository}}</image>
    <privileged>false</privileged>
    <command></command>
    <args></args>
    <instanceCap>5</instanceCap>
    <volumes/>
    <remoteFs>{{if index .metadata.annotations \"slave-directory\"}}{{index .metadata.annotations \"slave-directory\"}}{{else}}${DEFAULT_SLAVE_DIRECTORY}{{end}}</remoteFs>
    <label>{{if index .metadata.annotations \"slave-label\"}}{{index .metadata.annotations \"slave-label\"}}{{else}}${name}{{end}}</label>
  </org.csanchez.jenkins.plugins.kubernetes.PodTemplate>
  "
  echo "${template}" > ${template_file}
  $oc_cmd get -n "${PROJECT_NAME}" is/${name} -o templatefile --template ${template_file}
  rm -f ${template_file} &>/dev/null
}

# generate_kubernetes_config generates a configuration for the kubernetes plugin
function generate_kubernetes_config() {
    [ -z "$oc_cmd" ] && return
    local slave_templates=""
    if has_service_account; then
      for name in $(get_is_names); do
        slave_templates+=$(convert_is_to_slave ${name})
      done
    else
      return
    fi
    [ -z "${slave_templates}" ] && return
    echo "
    <org.csanchez.jenkins.plugins.kubernetes.KubernetesCloud plugin=\"kubernetes@0.5\">
      <name>openshift</name>
      <templates>
        ${slave_templates}
      </templates>
      <serverUrl>https://${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT}</serverUrl>
      <skipTlsVerify>true</skipTlsVerify>
      <namespace>${PROJECT_NAME}</namespace>
      <jenkinsUrl>http://${JENKINS_SERVICE_HOST}:${JENKINS_SERVICE_PORT}</jenkinsUrl>
      <credentialsId>1a12dfa4-7fc5-47a7-aa17-cc56572a41c7</credentialsId>
      <containerCap>10</containerCap>
      <retentionTimeout>5</retentionTimeout>
    </org.csanchez.jenkins.plugins.kubernetes.KubernetesCloud>
    "
}

# generate_kubernetes_credentials generates the credentials entry for the
# kubernetes service account.
function generate_kubernetes_credentials() {
  echo "<entry>
      <com.cloudbees.plugins.credentials.domains.Domain>
        <specifications/>
      </com.cloudbees.plugins.credentials.domains.Domain>
      <java.util.concurrent.CopyOnWriteArrayList>
        <org.csanchez.jenkins.plugins.kubernetes.ServiceAccountCredential plugin=\"kubernetes@0.4.1\">
          <scope>GLOBAL</scope>
          <id>1a12dfa4-7fc5-47a7-aa17-cc56572a41c7</id>
          <description></description>
        </org.csanchez.jenkins.plugins.kubernetes.ServiceAccountCredential>
      </java.util.concurrent.CopyOnWriteArrayList>
    </entry>
    "
}
