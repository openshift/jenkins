#!/bin/sh

export JENKINS_HOME=/var/lib/jenkins
export CONFIG_PATH=${JENKINS_HOME}/config.xml
export OPENSHIFT_API_URL=https://openshift.default.svc.cluster.local
export KUBE_SA_DIR=/run/secrets/kubernetes.io/serviceaccount
export KUBE_CA=${KUBE_SA_DIR}/ca.crt
export AUTH_TOKEN=${KUBE_SA_DIR}/token
export JENKINS_PASSWORD KUBERNETES_SERVICE_HOST KUBERNETES_SERVICE_PORT
export ITEM_ROOTDIR="\${ITEM_ROOTDIR}" # Preserve this variable Jenkins has in config.xml

# Jenkins LTS 2.516.1 removed the jbcrypt library from the WAR.
# Password hashing now uses Spring Security's BCryptPasswordEncoder instead.
#
# The Java source for the password utilities lives in jenkins-bcrypt-util/.
# At container startup, build_password_encoder_jar compiles these sources
# into password-encoder.jar using spring-security-crypto from the WAR.
#
# At runtime, the WAR lib directory is added to the classpath because
# BCryptPasswordEncoder also needs jcl-over-slf4j and slf4j-api for logging.
#
# ref: https://www.jenkins.io/changelog/2.516.1/
#      https://www.jenkins.io/changelog/2.509/

PASSWORD_ENCODER_SRC="/opt/openshift/jenkins-bcrypt-util"
PASSWORD_ENCODER_JAR="/opt/openshift/password-encoder.jar"

function build_password_encoder_jar {
    if [ -f "${PASSWORD_ENCODER_JAR}" ]; then
        return 0
    fi
    local spring_security_crypto=$(find /tmp/war/WEB-INF/lib ${JENKINS_HOME}/war/WEB-INF/lib -name "spring-security-crypto-*.jar" 2>/dev/null | head -1)
    javac -classpath "${spring_security_crypto}" \
        -d "${PASSWORD_ENCODER_SRC}" \
        "${PASSWORD_ENCODER_SRC}/com/redhat/openshift/PasswordEncoder.java" \
        "${PASSWORD_ENCODER_SRC}/com/redhat/openshift/PasswordChecker.java"
    jar cf "${PASSWORD_ENCODER_JAR}" -C "${PASSWORD_ENCODER_SRC}" com
}

# Takes a password, outputs the hashed password.
function obfuscate_password {
    local password="$1"
    local war_lib=$(find /tmp/war/WEB-INF/lib ${JENKINS_HOME}/war/WEB-INF/lib -maxdepth 0 -type d 2>/dev/null | head -1)
    
    build_password_encoder_jar
    
    java -classpath "${war_lib}/*:${PASSWORD_ENCODER_JAR}" com.redhat.openshift.PasswordEncoder "$password"
}

# Returns 0 if password matches 1 otherwise
function has_password_changed {
    local password="$1"
    local password_hash="$2"
    local war_lib=$(find /tmp/war/WEB-INF/lib ${JENKINS_HOME}/war/WEB-INF/lib -maxdepth 0 -type d 2>/dev/null | head -1)
    
    build_password_encoder_jar
    
    java -classpath "${war_lib}/*:${PASSWORD_ENCODER_JAR}" com.redhat.openshift.PasswordChecker "$password" "$password_hash"
}
