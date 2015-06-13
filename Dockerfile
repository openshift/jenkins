# openshift-jenkins
#
# This image provides a Jenkins server, primarily intended for integration with
# OpenShift v3.
#

FROM openshift/origin

ENV HOME /var/jenkins_home
ENV JENKINS_HOME /var/jenkins_home

RUN wget -O /etc/yum.repos.d/jenkins.repo http://pkg.jenkins-ci.org/redhat/jenkins.repo && \
  rpm --import http://pkg.jenkins-ci.org/redhat/jenkins-ci.org.key && \
  yum install -y zip unzip java-1.7.0-openjdk docker jenkins && yum clean all 

#RUN  curl -L -o /tmp/git.hpi http://updates.jenkins-ci.org/latest/git.hpi
#RUN  curl -L -o /tmp/git-client.hpi http://updates.jenkins-ci.org/latest/git-client.hpi
#RUN  curl -L -o /tmp/scm-api.hpi http://updates.jenkins-ci.org/latest/scm-api.hpi
#RUN  curl -L -o /tmp/ssh-credentials.hpi http://updates.jenkins-ci.org/latest/ssh-credentials.hpi
#RUN  curl -L -o /tmp/credentials.hpi http://updates.jenkins-ci.org/latest/credentials.hpi

RUN  usermod -m -d "$JENKINS_HOME" jenkins && \
  chown -R jenkins "$JENKINS_HOME"

COPY jenkins.sh /usr/local/bin/jenkins.sh

# for main web interface:
EXPOSE 8080

# will be used by attached slave agents:
EXPOSE 50000

USER jenkins

ENTRYPOINT ["/usr/local/bin/jenkins.sh"]
