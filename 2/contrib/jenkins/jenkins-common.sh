#!/bin/sh

export JENKINS_HOME=/var/lib/jenkins
export CONFIG_PATH=${JENKINS_HOME}/config.xml
export OPENSHIFT_API_URL=https://openshift.default.svc.cluster.local
export KUBE_SA_DIR=/run/secrets/kubernetes.io/serviceaccount
export KUBE_CA=${KUBE_SA_DIR}/ca.crt
export AUTH_TOKEN=${KUBE_SA_DIR}/token
export JENKINS_PASSWORD KUBERNETES_SERVICE_HOST KUBERNETES_SERVICE_PORT
export ITEM_ROOTDIR="\${ITEM_ROOTDIR}" # Preserve this variable Jenkins has in config.xml

# Takes a password and an optional salt value, outputs the hashed password.
function obfuscate_password {
    local password="$1"
    local salt="$2"
    #local acegi_security_path=`find /tmp/war/WEB-INF/lib/ -name acegi-security-*.jar`
    #local commons_codec_path=`find /tmp/war/WEB-INF/lib/ -name commons-codec-*.jar`
    local jbcrypt_path=`find /tmp/war/WEB-INF/lib ${JENKINS_HOME}/war/WEB-INF/lib/ -name jbcrypt-*.jar 2> /dev/null | head -1`
    # source for password-encoder.jar is inside the jar.
    # acegi-security-1.0.7.jar is inside the jenkins war.
#    java -classpath "${acegi_security_path}:${commons_codec_path}:/opt/openshift/password-encoder.jar" com.redhat.openshift.PasswordEncoder $password $salt
     java -classpath "${jbcrypt_path}:/opt/openshift/password-encoder.jar" com.redhat.openshift.PasswordEncoder $password $salt
}

# Returns 0 if password matches 1 otherwise
function has_password_changed {
    local password="$1"
    local password_hash="$2"
    local jbcrypt_path=`find /tmp/war/WEB-INF/lib ${JENKINS_HOME}/war/WEB-INF/lib/ -name jbcrypt-*.jar 2> /dev/null | head -1`
    # source for password-encoder.jar is inside the jar.
     java -classpath "${jbcrypt_path}:/opt/openshift/password-encoder.jar" com.redhat.openshift.PasswordChecker $password $password_hash
}


