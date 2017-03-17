编译jenkins-2-centos7镜像
make build VERSION=2

docker tag openshift/jenkins-2-centos7 hub.jucaicat.net/openshift/jenkins-2-centos7

docker push hub.jucaicat.net/openshift/jenkins-2-centos7



编译slave-gradle镜像
docker build -t hub.jucaicat.net/openshift/jenkins-slave-gradle-alpine .
docker push hub.jucaicat.net/openshift/jenkins-slave-gradle-alpine