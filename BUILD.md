编译jenkins-2-centos7镜像
make build VERSION=2

docker tag openshift/jenkins-2-centos7 hub.shandiancat.net/openshift/jenkins-2-centos7
docker push hub.shandiancat.net/openshift/jenkins-2-centos7



编译slave-gradle镜像
docker build -t hub.shandiancat.net/openshift/jenkins-slave-gradle-alpine .
docker push hub.shandiancat.net/openshift/jenkins-slave-gradle-alpine