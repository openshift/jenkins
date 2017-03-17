#FROM openjdk:8-jdk-alpine
FROM hub.jucaicat.net/openshift/jenkins-2-centos7

ENV JAVA_HOME /usr/lib/jvm/java-1.8.0-openjdk
ENV HOME /root
ENV GRADLE_OPTS -Dorg.gradle.daemon=false
#USER root

COPY gradle /usr/local/gradle
#RUN userdel  jenkins \
#	&& rm -rf /var/lib/jenkin \
#	&& groupadd -g 10000 jenkins && useradd -c "Jenkins user" -d $HOME -u 10000 -g 10000 -m jenkins

ARG VERSION=3.7

RUN curl --create-dirs -sSLo /usr/share/jenkins/slave.jar https://repo.jenkins-ci.org/public/org/jenkins-ci/main/remoting/${VERSION}/remoting-${VERSION}.jar \
  && chmod 755 /usr/share/jenkins \
  && chmod 644 /usr/share/jenkins/slave.jar

#USER jenkins
#add ssh-key
#RUN mkdir /home/jenkins/.jenkins &&  mkdir /home/jenkins/.ssh && chmod 700 /home/jenkins/.ssh
COPY id_rsa /root/.ssh/id_rsa
COPY known_hosts  /root/.ssh/known_hosts
RUN chmod 600 /root/.ssh/*
#RUN chmod 700 /home/jenkins/.ssh/id_rsa 

#VOLUME /home/jenkins/.jenkins
WORKDIR /root
COPY jenkins-slave /usr/local/bin/jenkins-slave
RUN  chmod 755 /usr/local/bin/jenkins-slave \
     && chmod 755 /usr/local/gradle/bin/*

ENV PATH $PATH:/usr/local/gradle/bin:$JAVA_HOME/bin
#ENV PATH $PATH:/home/jenkins/tools/hudson.plugins.gradle.GradleInstallation/gradle3/gradle-3.2-rc-1/bin
ENTRYPOINT ["jenkins-slave"]
