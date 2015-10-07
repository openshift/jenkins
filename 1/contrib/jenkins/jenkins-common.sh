#!/bin/sh

export DEFAULT_SLAVE_DIRECTORY=/opt/app-root/jenkins
export JENKINS_HOME=/var/lib/jenkins
export CONFIG_PATH=${JENKINS_HOME}/config.xml
export PROJECT_NAME=${PROJECT_NAME:-ci}
export OPENSHIFT_API_URL=https://openshift.default.svc.cluster.local
export KUBE_CA=/run/secrets/kubernetes.io/serviceaccount/ca.crt
export AUTH_TOKEN=/run/secrets/kubernetes.io/serviceaccount/token
export JENKINS_PASSWORD KUBERNETES_SERVICE_HOST KUBERNETES_SERVICE_PORT
export ITEM_ROOTDIR="\${ITEM_ROOTDIR}" # Preserve this variable Jenkins has in config.xml
export K8S_PLUGIN_POD_TEMPLATES=""

export oc_auth="--token=$(cat $AUTH_TOKEN) --certificate-authority=${KUBE_CA}"
export oc_cmd="oc -n ${PROJECT_NAME} --server=$OPENSHIFT_API_URL ${oc_auth}"

# get_imagestream_names returns a list of imagestreams names that contains
# label 'role=jenkins-slave'
function get_is_names() {
  $oc_cmd get is -l role=jenkins-slave -o template -t "{{range .items}}{{.metadata.name}} {{end}}"
}

# convert_is_to_slave converts the OpenShift imagestream to a Jenkins Kubernetes
# Plugin slave configuration.
function convert_is_to_slave() {
  local name=$1
  local template_file=$(mktemp)
  local template="
  <org.csanchez.jenkins.plugins.kubernetes.PodTemplate>
    <name>{{.metadata.name}}</name>
    <image>{{.status.dockerImageRepository}}</image>
    <privileged>false</privileged>
    <remoteFs>{{if index .metadata.annotations \"slave-directory\"}}{{index .metadata.annotations \"slave-directory\"}}{{else}}${DEFAULT_SLAVE_DIRECTORY}{{end}}</remoteFs>
    <instanceCap>5</instanceCap>
    <label>{{if index .metadata.annotations \"slave-label\"}}{{index .metadata.annotations \"slave-label\"}}{{else}}${name}{{end}}</label>
  </org.csanchez.jenkins.plugins.kubernetes.PodTemplate>
  "
  echo "${template}" > ${template_file}
  $oc_cmd get is/${name} -o templatefile -t ${template_file}
  rm -f ${template_file} &>/dev/null
}

# Generate passwd file based on current uid
function generate_passwd_file() {
  export USER_ID=$1
  export GROUP_ID=$2
  envsubst < /opt/openshift/passwd.template > /opt/openshift/passwd
  export LD_PRELOAD=libnss_wrapper.so
  export NSS_WRAPPER_PASSWD=/opt/openshift/passwd
  export NSS_WRAPPER_GROUP=/etc/group
}

function obfuscate_password {
    local password="$1"
    local acegi_security_path=`find /tmp/war/WEB-INF/lib/ -name acegi-security-*.jar`
    local commons_codec_path=`find /tmp/war/WEB-INF/lib/ -name commons-codec-*.jar`

    java -classpath "${acegi_security_path}:${commons_codec_path}:/opt/openshift/password-encoder.jar" com.redhat.openshift.PasswordEncoder $password
}


