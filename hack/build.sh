#!/bin/bash -e
# This script is used to build, test and squash the OpenShift Docker images.
#
# $1 - Specifies distribution - "rhel7" or "centos7"
# $2 - Specifies the image version - (must match with subdirectory in repo)# TEST_MODE - If set, build a candidate image and test it
# TEST_MODE - If set, build a candidate image and test it
# TAG_ON_SUCCESS - If set, tested image will be re-tagged as a non-candidate
#                  image, if the tests pass.
# VERSIONS - Must be set to a list with possible versions (subdirectories)

OS=$1
VERSION=$2

DOCKERFILE_PATH=""
BASE_IMAGE_NAME="openshift/jenkins"

# Cleanup the temporary Dockerfile created by docker build with version
trap "rm -f ${DOCKERFILE_PATH}.version" SIGINT SIGQUIT EXIT

# Perform docker build but append the LABEL with GIT commit id at the end
function docker_build_with_version {
  local dockerfile="$1"
  # Use perl here to make this compatible with OSX
  DOCKERFILE_PATH=$(perl -MCwd -e 'print Cwd::abs_path shift' $dockerfile)
  cp ${DOCKERFILE_PATH} "${DOCKERFILE_PATH}.version"
  git_version=$(git rev-parse --short HEAD)
  echo "LABEL io.openshift.builder-version=\"${git_version}\"" >> "${dockerfile}.version"
  docker build -t ${IMAGE_NAME} -f "${dockerfile}.version" .
  rm -f "${DOCKERFILE_PATH}.version"
}

# Versions are stored in subdirectories. You can specify VERSION variable
# to build just one single version. By default we build all versions
dirs=${VERSION:-$VERSIONS}

# enforce building of the slave-base image if we're building any of
# the slave images.  Note that we might build the slave-base
# twice if it was explicitly requested.  That's ok, it's 
# cheap to build it a second time.  The important thing
# is we have to build it before building any other
# slave image.
for dir in ${dirs}; do
  if [[ "$dir" =~ "slave" ]]; then
    dirs=( "slave-base ${dirs[@]}")
    break
  fi
done

for dir in ${dirs}; do
  IMAGE_NAME="${BASE_IMAGE_NAME}-${dir//./}-${OS}"

  if [[ ! -z "${TEST_MODE}" ]]; then
    IMAGE_NAME+="-candidate"
  fi

  echo "-> Building ${IMAGE_NAME} ..."

  pushd ${dir} > /dev/null
  if [ "$OS" == "rhel7" -o "$OS" == "rhel7-candidate" ]; then
    docker_build_with_version Dockerfile.rhel7
  else
    docker_build_with_version Dockerfile
  fi

  if [[ ! -z "${TEST_MODE}" ]]; then
    IMAGE_NAME=${IMAGE_NAME} test/run
    # always re-tag slave-base because we need it to build the other images even if we are just testing them.
    if [[ $? -eq 0 ]] && [[ "${TAG_ON_SUCCESS}" == "true" || "${dir}" == "slave-base" ]]; then
      echo "-> Re-tagging ${IMAGE_NAME} image to ${IMAGE_NAME%"-candidate"}"
      docker tag $IMAGE_NAME ${IMAGE_NAME%"-candidate"}
    fi

    if [[ ! -z "${REGISTRY}" ]]; then
      echo "-> Tagging image as" ${REGISTRY}/${IMAGE_NAME%"-candidate"}
      docker tag $IMAGE_NAME ${REGISTRY}/${IMAGE_NAME%"-candidate"}
    fi
  fi

  popd > /dev/null
done
