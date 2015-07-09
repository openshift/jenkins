#!/bin/bash

function obfuscate_password {
    local password="$1"
    local acegi_security_path=`find /tmp/war/WEB-INF/lib/ -name acegi-security-*.jar`
    local commons_codec_path=`find /tmp/war/WEB-INF/lib/ -name commons-codec-*.jar`

    java -classpath "${acegi_security_path}:${commons_codec_path}:/opt/openshift/password-encoder.jar" com.redhat.openshift.PasswordEncoder $password
}

if [ ! -e /var/lib/jenkins/configured ]; then
  mv /opt/openshift/configuration/* /var/lib/jenkins
  
  mkdir /tmp/war
  unzip -q /usr/lib/jenkins/jenkins.war -d /tmp/war
  admin_password_hash=`obfuscate_password ${JENKINS_PASSWORD:-password}`
  rm -rf /tmp/war

  sed -i "s,PASSWORD,$admin_password_hash,g" "/var/lib/jenkins/users/admin/config.xml"
  touch /var/lib/jenkins/configured
fi


# if `docker run` first argument start with `--` the user is passing jenkins launcher arguments
if [[ $# -lt 1 ]] || [[ "$1" == "--"* ]]; then
   exec java $JAVA_OPTS -jar /usr/lib/jenkins/jenkins.war $JENKINS_OPTS "$@"
fi

# As argument is not jenkins, assume user want to run his own process, for sample a `bash` shell to explore this image
exec "$@"

