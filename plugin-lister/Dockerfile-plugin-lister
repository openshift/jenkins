FROM quay.io/openshift/origin-jenkins:v4.0
USER root
COPY base-plugins.txt /opt/openshift/configuration
RUN /usr/local/bin/install-plugins.sh /opt/openshift/configuration/base-plugins.txt