# Hacking

The basics to build this project and test it locally

## Locally 

## Building locally
Jenkins images can be built using  `podman`, `buildah` or `docker`. For historical reasons, the default build runtime is `docker`

To build on podman:
```
make VERSIONS="2 maven-3.5" BUILD_COMMAND="podman build --no-cache"
```


## Deploying
To deploy your Jenkins built images refer to the section [ Deploying on an OpenShift Cluster ]


## OpenShift

Be sure that you are logged-in with an OpenShift cluster or that KUBECONFIG environment variable points to a valid kubeconfig.

### Building on an OpenShift cluster

```
oc new-build https://github.com/origin/jenkins.git#your-branch-name --context-dir=2/
```

### Deploying on an OpenShift Cluster
```
oc new-app jenkins-persistent -p NAMESPACE=$(oc project -q) -p JENKINS_IMAGE_STREAM_TAG=jenkins:latest
```



