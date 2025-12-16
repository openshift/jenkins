#!/bin/env bash

#
# Run the Jenkins image on your local workstation using podman or similar
# so that you get an interactive command prompt, something like
# $ podman run -it --entrypoint /bin/bash <image>
#
# You can then paste the following command at the command prompt
# which will download this script file and run it in the container
#
# NOTE: Don't forget to replace <commit sha> with the SHA from
#       github.com/openshift/jenkins that you would like to test
#       the image against
# curl https://raw.githubusercontent.com/openshift/jenkins/master/scripts/verify-jenkins.sh | SHA=<commit sha> sh
#
# NOTE: If you want to run the script a second time, you will need to exit the container and
#       re-run it using the `podman run -it --entrypoint /bin/bash <image>` command
#       because the /usr/libexec/s2i/run command moves the Jenkins plugins to a different
#       location and all of the plugin exists and plugin version verifications will fail
#
set -o pipefail

# Default Github Jenkins base url
GITHUB_JENKINS_BASE_URL='https://raw.githubusercontent.com/openshift/jenkins'

# Default file/folder locations
PLUGINS_DIR=${PLUGINS_DIR:-/opt/openshift/plugins}
BUNDLE_PLUGINS_LOCATION=${BUNDLE_PLUGINS_LOCATION:-/opt/openshift/bundle-plugins.txt}
JENKINS_WAR_LOCATION=${JENKINS_LOCATION:-/usr/lib/jenkins/jenkins.war}
JENKINS_VERSION_LOCATION=${JENKINS_VERSION_LOCATION:-/opt/openshift/jenkins-version.txt}
JENKINS_START_SCRIPT_LOCATION=/usr/libexec/s2i/run
COMMIT_JENKINS_VERSION_LOCATION=/tmp/jenkins-version.txt
COMMIT_BUNDLE_PLUGINS_LOCATION=/tmp/bundle-plugins.txt

# Location of the log file for failed verifications
FAILED_LOG_LOCATION=/tmp/failed.log
# Ensure that the log file does not exist from a previous run
rm -rf $FAILED_LOG_LOCATION

# ERROR_STRINGS
#
# An array of error strings to check for in the Jenkins log file
ERROR_STRINGS=(
			"Failed to extract the bundled plugin"
			"Failed Loading plugin"
			)

# ensure_sha_env_var
#
# Ensure that the environment variable SHA is set
ensure_sha_env_var() {
	if [ -z "$SHA" ]; then
		printf "You MUST supply a SHA to verify against\n"
		printf "Usage: curl https://raw.githubusercontent.com/openshift/jenkins/master/scripts/verify-jenkins.sh | SHA=<commit sha> sh \n"
		exit 1
	fi
}

# The base url for the files that we need from the code repository
GITHUB_JENKINS_COMMIT_URL="$GITHUB_JENKINS_BASE_URL/$SHA"

# verify_file()
#
# verifies that $local_file_path/$file_name exists and that the contents match
# those of $remote_file_path/$file_name at the provided SHA
#
# Example: verify_file "file.txt" "/my/local/path" "/my/remote/path"
#
# @param file_name; The file name to use
# @param local_file_path; The path to the file in the container
# @param remote_file_path; The path to the file in the github repository
#
# @return void
verify_file(){
	local file_name=$1
	local local_file_path="${2}/${file_name}"
	local remote_file_path="${GITHUB_JENKINS_COMMIT_URL}/${3}/${file_name}"

	# The temporary file where the contents of the remote file are stored
	local tmp_file_path="/tmp/${file_name}"

	printf "[VERIFYING] %s\n" "${local_file_path}"

	# Download the remote file to a temporary file, verifying the http_status code for the request
	printf "\t%-50s" "- downloading ... "
	http_status_code=$(curl -s -o "${tmp_file_path}" -w "%{http_code}\n" "${remote_file_path}" 2> /dev/null)

	if [ "$http_status_code" == 200 ]; then
		printf "PASS\n"
	else
		printf "FAIL\n"
		printf "failed to download %s, http status code: %s" "${remote_file_path}" "${http_status_code}" >> $FAILED_LOG_LOCATION
	fi

	# Check if the local file exists
	printf "\t%-50s" "- exists ... "
	if test -f "${local_file_path}"; then
		printf "PASS\n"
	else
		printf "FAIL\n"
		echo "${local_file_path} does not exist but should" >> $FAILED_LOG_LOCATION
	fi

	# Check if the contents of the local file match those of the remote file
	printf "\t%-50s" "- contents match ... "
	if cmp -s "${local_file_path}" "${tmp_file_path}"; then
		printf "PASS\n"
	else
		printf "FAIL\n"
		printf "%s comparison failed\n" "${file_name}" >> $FAILED_LOG_LOCATION
		printf "%s\n" "$(diff "${local_file_path}" "${tmp_file_path}")" >> $FAILED_LOG_LOCATION
	fi
}

# verify_plugins()
#
# verifies that the plugins from the bundle-plugins.txt file exist and that they
# are the at version specified by the bundle-plugins.txt file in the code repository
# at the specified SHA
#
verify_plugins() {
	while IFS= read -r plugin
	do
		if [[ "$plugin" =~ "Generated" ]]; then
			continue
		fi
		name="${plugin%%:*}"
		version="${plugin##*:}"

		printf "[VERIFYING] %s.jpi:%s\n" "${name}" "${version}"
		printf "\t%-50s" "- exists ... "
		if test -f "${PLUGINS_DIR}/${name}.jpi"; then
			printf "PASS\n"
		else
			printf "FAIL\n"
			echo "${PLUGINS_DIR}/${name}.jpi does not exist but should" >> $FAILED_LOG_LOCATION
		fi

		found_version=$(unzip -p "${PLUGINS_DIR}/${name}.jpi" META-INF/MANIFEST.MF | grep "^Plugin-Version: " | sed -e 's#^Plugin-Version: ##')
		found_version=${found_version%%[[:space:]]}
		printf "\t%-50s" "- version matches ... "
		if [ "$version" == "$found_version" ]; then
			printf "PASS\n"
		else
			printf "FAIL\n"
			echo " ${PLUGINS_DIR}/${plugin} should be ${version} but was ${found_version} instead" >> $FAILED_LOG_LOCATION
		fi
	done < "$COMMIT_BUNDLE_PLUGINS_LOCATION"
}

# verify_jenkins_war()
#
# verifies that the /usr/lib/jenkins/jenkins.war file exists and is at the
# same version specified by the jenkins-version.txt file at the provided SHA
#
verify_jenkins_war() {
	jenkins_version=$(cat "$COMMIT_JENKINS_VERSION_LOCATION")
	printf "[VERIFYING] %s:%s\n" "$JENKINS_WAR_LOCATION" "${jenkins_version}"
	jenkins_version_found=$(/usr/lib/jvm/java-21/bin/java -jar "$JENKINS_WAR_LOCATION" --version 2> /dev/null)
	printf "\t%-50s" "- exists ... "
		if test -f "$JENKINS_WAR_LOCATION"; then
			printf "PASS\n"
		else
			printf "FAIL\n"
			printf "%s does not exist" "$JENKINS_WAR_LOCATION" >> $FAILED_LOG_LOCATION
		fi
	printf "\t%-50s" "- version matches ... "
	if [[ "$jenkins_version_found" == "$jenkins_version" ]]; then
		printf "PASS\n"
	else
		printf "FAIL\n"
		echo "wanted Jenkins version ${jenkins_version} but found ${jenkins_version_found} instead" >> $FAILED_LOG_LOCATION
	fi
}

# verify_jenkins_startup()
#
# verifies that Jenkins starts up without any plugin conflicts
# Jenkins will ultimately fail to start correctly but enough of the process
# completes that we can check for this specific set of issues
#
verify_jenkins_startup() {
	if test -f $JENKINS_START_SCRIPT_LOCATION; then
		printf "Starting Jenkins with /usr/libexec/s2i/bin\n"
		printf "This could take up to 60 seconds\n"
		$JENKINS_START_SCRIPT_LOCATION &> /tmp/jenkins.log &
		for i in {1..60}
		do
			printf "."
			sleep 1
		done
		printf "\n[VERIFYING] Jenkins startup\n"
		for e in "${ERROR_STRINGS[@]}"; do
			printf "\t%-50s" "- $e ..."
			FOUND=$(grep /tmp/jenkins.log -e "$e" -c)
			if [ "$FOUND" == 0 ]; then
				printf "PASS\n"

			else
				printf "FAIL\n"
				printf "found %s occurences of '%s'\n" "$FOUND" "$e" >> $FAILED_LOG_LOCATION
			fi
		done
	else
		printf "%s does not exist, but should\n" "$JENKINS_START_SCRIPT_LOCATION" >> $FAILED_LOG_LOCATION
	fi
}

# check_failed_log()
#
# checks if the /tmp/failed.log file exists, if so there were some errors during verification
# that need to be fixed
#
check_failed_log() {
	if [[ -f $FAILED_LOG_LOCATION ]]; then
		echo -e "\nSome errors were encountered:\n$(<"$FAILED_LOG_LOCATION")" >&2
		exit 1
	else
		echo -e "\nAll tests succeeded!"
	fi
}

verify_installed_packages() {
	http_status_code=$(curl -s -o "/tmp/Dockerfile.rhel8" -w "%{http_code}\n" "${GITHUB_JENKINS_BASE_URL}/${SHA}/2/Dockerfile.rhel8-multi-arch" 2> /dev/null)
	printf "[VERIFYING] Installed packages ...\n"

	# Download the remote file to a temporary file, verifying the http_status code for the request
	printf "\t%-50s" "- downloading Dockerfile.rhel8 ... "
	if [ "$http_status_code" == 200 ]; then
		printf "PASS\n"
	else
		printf "FAIL\n"
		printf "failed to download %s, http status code: %s" "${remote_file_path}" "${http_status_code}" >> $FAILED_LOG_LOCATION
	fi

	 INSTALL_PKGS=$(grep "INSTALL_PKGS=" /tmp/Dockerfile.rhel8)
	 INSTALL_PKGS=${INSTALL_PKGS//INSTALL_PKGS=\"/}
	 INSTALL_PKGS=${INSTALL_PKGS//\" && \\/}
	 INSTALL_PKGS=$(echo "${INSTALL_PKGS}" | xargs)

	for e in ${INSTALL_PKGS[@]}; do
			printf "\t%-50s" "- checking $e ..."
			if ! yum list installed "${e}" &> /dev/null; then
				printf "FAIL\n"
				printf "package %s is not installed" "${e}\n" >> $FAILED_LOG_LOCATION
			else
				printf "PASS\n"
			fi
	done
}

main() {
	ensure_sha_env_var
	verify_installed_packages
	verify_file "jenkins-version.txt" "/opt/openshift" "2/contrib/openshift"
	verify_file "bundle-plugins.txt" "/opt/openshift" "2/contrib/openshift"
	verify_plugins
	verify_jenkins_war
	verify_jenkins_startup
	check_failed_log
}

main
