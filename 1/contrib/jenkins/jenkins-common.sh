#!/bin/sh

export JENKINS_HOME=/var/lib/jenkins
export CONFIG_PATH=${JENKINS_HOME}/config.xml
export OPENSHIFT_API_URL=https://openshift.default.svc.cluster.local
export KUBE_CA=/run/secrets/kubernetes.io/serviceaccount/ca.crt
export AUTH_TOKEN=/run/secrets/kubernetes.io/serviceaccount/token
export JENKINS_PASSWORD KUBERNETES_SERVICE_HOST KUBERNETES_SERVICE_PORT
export ITEM_ROOTDIR="\${ITEM_ROOTDIR}" # Preserve this variable Jenkins has in config.xml

if [ -f "${AUTH_TOKEN}" ]; then
  export oc_auth="--token=$(cat $AUTH_TOKEN) --certificate-authority=${KUBE_CA}"
  export oc_cmd="oc --server=$OPENSHIFT_API_URL ${oc_auth}"
fi

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
