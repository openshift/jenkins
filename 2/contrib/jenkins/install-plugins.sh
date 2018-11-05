#! /bin/bash -eu
#
# Originally copied from https://github.com/jenkinsci/docker
# You can set JENKINS_UC to change the default URL to Jenkins update center
#
# Usage:
#
# FROM openshift/jenkins-2-centos7
# COPY plugins.txt /plugins.txt
# RUN /usr/local/bin/install-plugins.sh /plugins.txt
#
# The format of 'plugins.txt. is:
#
# pluginId:pluginVersion

set -o pipefail

# BEGIN - From https://raw.githubusercontent.com/jenkinsci/docker/master/jenkins-support
# compare if version1 < version2
versionLT() {
    local v1; v1=$(echo "$1" | cut -d '-' -f 1 )
    local q1; q1=$(echo "$1" | cut -s -d '-' -f 2- )
    local v2; v2=$(echo "$2" | cut -d '-' -f 1 )
    local q2; q2=$(echo "$2" | cut -s -d '-' -f 2- )
    if [ "$v1" = "$v2" ]; then
        if [ "$q1" = "$q2" ]; then
            return 1
        else
            if [ -z "$q1" ]; then
                return 1
            else
                if [ -z "$q2" ]; then
                    return 0
                else
                    [  "$q1" = "$(echo -e "$q1\n$q2" | sort -V | head -n1)" ]
                fi
            fi
        fi
    else
        [  "$v1" = "$(echo -e "$v1\n$v2" | sort -V | head -n1)" ]
    fi
}

# returns a plugin version from a plugin archive
get_plugin_version() {
    local archive; archive=$1
    local version; version=$(unzip -p "$archive" META-INF/MANIFEST.MF | grep "^Plugin-Version: " | sed -e 's#^Plugin-Version: ##')
    version=${version%%[[:space:]]}
    echo "$version"
}

# Copy files from /usr/share/jenkins/ref into $JENKINS_HOME
# So the initial JENKINS-HOME is set with expected content.
# Don't override, as this is just a reference setup, and use from UI
# can then change this, upgrade plugins, etc.
copy_reference_file() {
    f="${1%/}"
    b="${f%.override}"
    rel="${b:23}"
    version_marker="${rel}.version_from_image"
    dir=$(dirname "${b}")
    local action;
    local reason;
    local container_version;
    local image_version;
    local marker_version;
    local log; log=false
    if [[ ${rel} == plugins/*.jpi ]]; then
        container_version=$(get_plugin_version "$JENKINS_HOME/${rel}")
        image_version=$(get_plugin_version "${f}")
        if [[ -e $JENKINS_HOME/${version_marker} ]]; then
            marker_version=$(cat "$JENKINS_HOME/${version_marker}")
            if versionLT "$marker_version" "$container_version"; then
                action="SKIPPED"
                reason="Installed version ($container_version) has been manually upgraded from initial version ($marker_version)"
                log=true
            else
                if [[ "$image_version" == "$container_version" ]]; then
                    action="SKIPPED"
                    reason="Version from image is the same as the installed version $image_version"
                else
                    if versionLT "$image_version" "$container_version"; then
                        action="SKIPPED"
                        log=true
                        reason="Image version ($image_version) is older than installed version ($container_version)"
                    else
                        action="UPGRADED"
                        log=true
                        reason="Image version ($image_version) is newer than installed version ($container_version)"
                    fi
                fi
            fi
        else
            if [[ -n "$TRY_UPGRADE_IF_NO_MARKER" ]]; then
                if [[ "$image_version" == "$container_version" ]]; then
                    action="SKIPPED"
                    reason="Version from image is the same as the installed version $image_version (no marker found)"
                    # Add marker for next time
                    echo "$image_version" > "$JENKINS_HOME/${version_marker}"
                else
                    if versionLT "$image_version" "$container_version"; then
                        action="SKIPPED"
                        log=true
                        reason="Image version ($image_version) is older than installed version ($container_version) (no marker found)"
                    else
                        action="UPGRADED"
                        log=true
                        reason="Image version ($image_version) is newer than installed version ($container_version) (no marker found)"
                    fi
                fi
            fi
        fi
        if [[ ! -e $JENKINS_HOME/${rel} || "$action" == "UPGRADED" || $f = *.override ]]; then
            action=${action:-"INSTALLED"}
            log=true
            mkdir -p "$JENKINS_HOME/${dir:23}"
            # if done on rhel, we may need to override a link to /usr/lib/jenkins, so include --remove-destination
            cp --remove-destination -r "${f}" "$JENKINS_HOME/${rel}";
            # pin plugins on initial copy
            touch "$JENKINS_HOME/${rel}.pinned"
            echo "$image_version" > "$JENKINS_HOME/${version_marker}"
            reason=${reason:-$image_version}
        else
            action=${action:-"SKIPPED"}
        fi
    else
        if [[ ! -e $JENKINS_HOME/${rel} || $f = *.override ]]
        then
            action="INSTALLED"
            log=true
            mkdir -p "$JENKINS_HOME/${dir:23}"
            # if done on rhel, we may need to override a link to /usr/lib/jenkins, so include --remove-destination
            cp --remove-destination -r "${f}" "$JENKINS_HOME/${rel}";
        else
            action="SKIPPED"
        fi
    fi
    if [[ -n "$VERBOSE" || "$log" == "true" ]]; then
        if [ -z "$reason" ]; then
            echo "$action $rel" >> "$COPY_REFERENCE_FILE_LOG"
        else
            echo "$action $rel : $reason" >> "$COPY_REFERENCE_FILE_LOG"
        fi
    fi
}
# END - From https://raw.githubusercontent.com/jenkinsci/docker/master/jenkins-support

REF_DIR=${REF:-/opt/openshift/plugins}
FAILED="$REF_DIR/failed-plugins.txt"

JENKINS_WAR=/usr/lib/jenkins/jenkins.war

INCREMENTAL_BUILD_ARTIFACTS_DIR="/tmp/artifacts"

function getLockFile() {
    echo -n "$REF_DIR/${1}.lock"
}

function getArchiveFilename() {
    echo -n "$REF_DIR/${1}.jpi"
}

function download() {
    local plugin originalPlugin version lock ignoreLockFile
    plugin="$1"
    version="${2:-latest}"
    ignoreLockFile="${3:-}"
    lock="$(getLockFile "$plugin")"

    if [[ $ignoreLockFile ]] || mkdir "$lock" &>/dev/null; then
        if ! doDownload "$plugin" "$version"; then
            # some plugin don't follow the rules about artifact ID
            # typically: docker-plugin
            originalPlugin="$plugin"
            plugin="${plugin}-plugin"
            if ! doDownload "$plugin" "$version"; then
                echo "Failed to download plugin: $originalPlugin or $plugin" >&2
                echo "Not downloaded: ${originalPlugin}" >> "$FAILED"
                return 1
            fi
        fi

        if ! checkIntegrity "$plugin"; then
            echo "Downloaded file is not a valid ZIP: $(getArchiveFilename "$plugin")" >&2
            echo "Download integrity: ${plugin}" >> "$FAILED"
            return 1
        fi

        resolveDependencies "$plugin"
    fi
}

function doDownload() {
    local plugin version url jpi
    plugin="$1"
    version="$2"
    jpi="$(getArchiveFilename "$plugin")"

    # If plugin already exists and is the same version do not download
    if test -f "$jpi" && unzip -p "$jpi" META-INF/MANIFEST.MF | tr -d '\r' | grep "^Plugin-Version: ${version}$" > /dev/null; then
        echo "Using provided plugin: $plugin"
        return 0
    fi

    # Check if the plugin is cached and in correct version; if so, use it instead of downloading
    # Some plugins do not follow the naming conventions and include the "-plugin" suffix; both cases need to be covered
    for pluginFilename in "$plugin.jpi" "$plugin-plugin.jpi"; do
        local cachedPlugin="$INCREMENTAL_BUILD_ARTIFACTS_DIR/plugins/$pluginFilename"
        if test -f "$cachedPlugin" && [[ $(get_plugin_version "$cachedPlugin") == "$version" ]]; then
            echo "Copying plugin from a cache created by s2i: $cachedPlugin"
            cp "$cachedPlugin" "$jpi"
            return 0
        fi
    done

    if [[ "$version" == "latest" && -n "$JENKINS_UC_LATEST" ]]; then
        # If version-specific Update Center is available, which is the case for LTS versions,
        # use it to resolve latest versions.
        url="$JENKINS_UC_LATEST/latest/${plugin}.hpi"
    elif [[ "$version" == "experimental" && -n "$JENKINS_UC_EXPERIMENTAL" ]]; then
        # Download from the experimental update center
        url="$JENKINS_UC_EXPERIMENTAL/latest/${plugin}.hpi"
    else
        JENKINS_UC_DOWNLOAD=${JENKINS_UC_DOWNLOAD:-"$JENKINS_UC/download"}
        url="$JENKINS_UC_DOWNLOAD/plugins/$plugin/$version/${plugin}.hpi"
    fi

    echo "Downloading plugin: $plugin from $url"
    curl --connect-timeout "${CURL_CONNECTION_TIMEOUT:-20}" --retry "${CURL_RETRY:-5}" --retry-delay "${CURL_RETRY_DELAY:-0}" --retry-max-time "${CURL_RETRY_MAX_TIME:-60}" -s -f -L "$url" -o "$jpi"
    return $?
}

function checkIntegrity() {
    local plugin jpi
    plugin="$1"
    jpi="$(getArchiveFilename "$plugin")"

    zip -T "$jpi" >/dev/null
    return $?
}

function resolveDependencies() {
    local plugin jpi dependencies
    plugin="$1"
    jpi="$(getArchiveFilename "$plugin")"

    set +o pipefail
    dependencies="$(unzip -p "$jpi" META-INF/MANIFEST.MF | tr -d '\r' | tr '\n' '|' | sed -e 's#| ##g' | tr '|' '\n' | grep "^Plugin-Dependencies: " | sed -e 's#^Plugin-Dependencies: ##')"
    set -o pipefail

    if [[ ! $dependencies ]]; then
        echo " > $plugin has no dependencies"
        return
    fi

    echo " > $plugin depends on $dependencies"

    IFS=',' read -a array <<< "$dependencies"

    for d in "${array[@]}"
    do
        plugin="$(cut -d':' -f1 - <<< "$d")"
        #
        # Note, matrix-auth plugin notes cloudbees-folder as optional in the archive, but then failed to
        # load, citing a dependency that is too old, during testing ... so we will download optional dependencies
        #
        local versionFromPluginParam
        if [[ $d == *"resolution:=optional"* ]]; then
            echo "Examining optional dependency $plugin"
            optional_jpi="$(getArchiveFilename "$plugin")"
            if [ ! -f "${optional_jpi}" ]; then
                echo "Optional dependency $plugin not installed already, skipping"
                continue
            fi
            echo "Optional dependency $plugin already installed, need to determine if it is at a sufficient version"
            versionFromPluginParam="$(cut -d';' -f1 - <<< "$d")"
        else
            versionFromPluginParam=$d
        fi
        local pluginInstalled
        local minVersion; minVersion=$(versionFromPlugin "${versionFromPluginParam}")

    set +o pipefail
    local filename; filename=$(getArchiveFilename "$plugin")
    local previouslyDownloadedVersion; previouslyDownloadedVersion=$(get_plugin_version $filename)
    set -o pipefail
    
        # ${bundledPlugins} checks for plugins bundled in the jenkins.war file; per 
        # https://wiki.jenkins-ci.org/display/JENKINS/Bundling+plugins+with+Jenkins this is getting
        # phased out, but we are keeping this check in for now while that transition bakes a bit more    
        if pluginInstalled="$(echo "${bundledPlugins}" | grep "^${plugin}:")"; then
            pluginInstalled="${pluginInstalled//[$'\r']}"
            # get the version of the plugin bundled
            local versionInstalled; versionInstalled=$(versionFromPlugin "${pluginInstalled}")
            # if the bundled plugins is older than the minimum version needed for the dependency,
            # download the dependence; passing "true" is needed for "download" to replace the existing dependency
            if versionLT "${versionInstalled}" "${minVersion}"; then
                echo "Upgrading bundled dependency $d ($minVersion > $versionInstalled)"
                download "$plugin" "$minVersion" "true"
            else
                echo "Skipping already bundled dependency $d ($minVersion <= $versionInstalled)"
            fi
            # bypass further processing if a bundled plugin
            continue
        fi

        # if the dependency plugin has yet to be downloaded (hence the var is not set) download
        if [[ -z "${previouslyDownloadedVersion:-}" ]]; then
            echo "Downloading dependency plugin $plugin version $minVersion that has yet to be installed"
            download "$plugin" "$minVersion"
        else
            # get the version of the dependency plugin already downloaded; if not recent enough, download
            # the minimum version required; the "true" parameter is need for "download" to overwrite the existing
            # version of the plugin
            if versionLT "${previouslyDownloadedVersion}" "${minVersion}"; then
                echo "Upgrading previously downloaded plugin $plugin at $previouslyDownloadedVersion to $minVersion"
                download "$plugin" "$minVersion" "true"
            fi
        fi
    done
    wait
}

function bundledPlugins() {
    if [ -f $JENKINS_WAR ]
    then
        TEMP_PLUGIN_DIR=/tmp/plugintemp.$$
        for i in $(jar tf $JENKINS_WAR | egrep '[^detached-]plugins.*\..pi' | sort)
        do
            rm -fr $TEMP_PLUGIN_DIR
            mkdir -p $TEMP_PLUGIN_DIR
            PLUGIN=$(basename "$i"|cut -f1 -d'.')
            (cd $TEMP_PLUGIN_DIR;jar xf "$JENKINS_WAR" "$i";jar xvf "$TEMP_PLUGIN_DIR/$i" META-INF/MANIFEST.MF >/dev/null 2>&1)
            VER=$(egrep -i Plugin-Version "$TEMP_PLUGIN_DIR/META-INF/MANIFEST.MF"|cut -d: -f2|sed 's/ //')
            echo "$PLUGIN:$VER"
        done
        rm -fr $TEMP_PLUGIN_DIR
    else
        echo "ERROR file not found: $JENKINS_WAR"
        exit 1
    fi
}

function versionFromPlugin() {
    local plugin=$1
    if [[ $plugin =~ .*:.* ]]; then
        echo "${plugin##*:}"
    else
        echo "latest"
    fi

}

function installedPlugins() {
    for f in "$REF_DIR"/*.jpi; do
        echo "$(basename "$f" | sed -e 's/\.jpi//'):$(get_plugin_version "$f")"
    done
}

function jenkinsMajorMinorVersion() {
    if [[ -f "$JENKINS_WAR" ]]; then
        local version major minor
        version="$(/etc/alternatives/java -jar $JENKINS_WAR --version)"
        major="$(echo "$version" | cut -d '.' -f 1)"
        minor="$(echo "$version" | cut -d '.' -f 2)"
        echo "$major.$minor"
    else
        echo "ERROR file not found: $JENKINS_WAR"
        return 1
    fi
}

main() {
    local plugin version

    mkdir -p "$REF_DIR" || exit 1

    for file in $@; do
        # clean up any dos file injected carriage returns
        sed -i 's/\r$//' $file
    done

    # Create lockfile manually before first run to make sure any explicit version set is used.
    echo "Creating initial locks..."
    for plugin in `cat $@ | grep -v ^#`; do
        if [ -z $plugin ]; then
            continue
        fi
        echo "Locking $plugin"
        mkdir "$(getLockFile "${plugin%%:*}")"
    done

    echo -e "\nAnalyzing war..."
    bundledPlugins="$(bundledPlugins)"
    
    # Check if there's a version-specific update center, which is the case for LTS versions
    jenkinsVersion="$(jenkinsMajorMinorVersion)"
    if curl -fsL -o /dev/null "$JENKINS_UC/$jenkinsVersion"; then
        JENKINS_UC_LATEST="$JENKINS_UC/$jenkinsVersion"
        echo "Using version-specific update center: $JENKINS_UC_LATEST..."
    else
        JENKINS_UC_LATEST=
    fi

    echo -e "\nDownloading plugins..."
    for plugin in `cat $@ | grep -v ^#`; do
        if [ -z $plugin ]; then
            continue
        fi
        version=""
        if [[ $plugin =~ .*:.* ]]; then
            version=$(versionFromPlugin "${plugin}")
            plugin="${plugin%%:*}"
        fi

        download "$plugin" "$version" "true"
    done
    wait

    echo
    echo "WAR bundled plugins:"
    echo "${bundledPlugins}"
    echo
    echo "Installed plugins:"
    installedPlugins

    if [[ -f $FAILED ]]; then
        echo -e "\nSome plugins failed to download!\n$(<"$FAILED")" >&2
        exit 1
    fi

    echo -e "\nCleaning up locks"
    rm -rf "$REF_DIR"/*.lock
}

main "$@"
