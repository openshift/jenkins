#!/bin/sh

export JENKINS_HOME=/var/lib/jenkins
export CONFIG_PATH=${JENKINS_HOME}/config.xml
export OPENSHIFT_API_URL=https://openshift.default.svc.cluster.local
export KUBE_SA_DIR=/run/secrets/kubernetes.io/serviceaccount
export KUBE_CA=${KUBE_SA_DIR}/ca.crt
export AUTH_TOKEN=${KUBE_SA_DIR}/token
export JENKINS_PASSWORD KUBERNETES_SERVICE_HOST KUBERNETES_SERVICE_PORT
export ITEM_ROOTDIR="\${ITEM_ROOTDIR}" # Preserve this variable Jenkins has in config.xml

# Takes a password, outputs the hashed password.
# Jenkins 2.509+ and LTS 2.516.1 removed jbcrypt; password-encoder.jar now uses Spring Security's
# BCryptPasswordEncoder, which needs spring-security-crypto and its transitive
# dependencies (spring-jcl) from the WAR lib at runtime.
# 
# ref: https://www.jenkins.io/changelog/2.516.1/
#      https://www.jenkins.io/changelog/2.509/
function obfuscate_password {
    local password="$1"
    local war_lib=$(find /tmp/war/WEB-INF/lib ${JENKINS_HOME}/war/WEB-INF/lib -maxdepth 0 -type d 2>/dev/null | head -1)
    java -classpath "${war_lib}/*:/opt/openshift/password-encoder.jar" com.redhat.openshift.PasswordEncoder "$password"
}

# Returns 0 if password matches 1 otherwise
function has_password_changed {
    local password="$1"
    local password_hash="$2"
    local war_lib=$(find /tmp/war/WEB-INF/lib ${JENKINS_HOME}/war/WEB-INF/lib -maxdepth 0 -type d 2>/dev/null | head -1)
    java -classpath "${war_lib}/*:/opt/openshift/password-encoder.jar" com.redhat.openshift.PasswordChecker "$password" "$password_hash"
}


