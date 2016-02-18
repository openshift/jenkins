#!/bin/sh

source /usr/local/bin/jenkins-common.sh

# Since OpenShift runs this Docker image under random user ID, we have to assign
# the 'jenkins' user name to this UID. For that we use nss_wrapper and
# passwd.template.
# If you adding more layers to this Docker image and the layer you adding add
# more system users, make sure you update the passwd.template.
generate_passwd_file `id -u` `id -g`

set -eu
cmd="$1"; shift
exec $cmd "$@"
