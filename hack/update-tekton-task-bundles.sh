#!/bin/bash

# Use this script to update the Tekton Task Bundle references used in a Pipeline or a PipelineRun.
# update-tekton-task-bundles.sh .tekton/*.yaml

set -euo pipefail

FILES=$@

# Determine the flavor of yq and adjust yq commands accordingly
if [ -z "$(yq --version | grep mikefarah)" ]; then
   # Python yq
   YQ_FRAGMENT1='.. | select(type == "object" and has("resolver"))'
   YQ_FRAGMENT2='-r'
else
   # mikefarah yq
   YQ_FRAGMENT1='... | select(has("resolver"))'
   YQ_FRAGMENT2=''
fi

# Find existing image references
OLD_REFS="$(\
    yq "$YQ_FRAGMENT1 | .params // [] | .[] | select(.name == \"bundle\") | .value"  $FILES | \
    grep -v -- '---' | \
    sed 's/^"\(.*\)"$/\1/' | \
    sort -u \
)"

# Find updates for image references
for old_ref in ${OLD_REFS}; do
    repo_tag="${old_ref%@*}"
    new_digest="$(skopeo inspect --no-tags docker://${repo_tag} | yq $YQ_FRAGMENT2 '.Digest')"
    new_ref="${repo_tag}@${new_digest}"
    [[ $new_ref == $old_ref ]] && continue
    echo "New digest found! $new_ref"
    for file in $FILES; do
        sed -i -e "s!${old_ref}!${new_ref}!g" $file
    done
done
