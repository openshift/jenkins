#!/bin/bash

make plugins-list
DIFF=$(git diff 2/contrib/openshift/bundle-plugins.txt | wc -l)
if [[ ${DIFF} != 0 ]]; then
  echo "ERROR: Computed list of plugins in the bundle has changed since last commit"
  echo "ERROR: Please run 'make plugins-list' and ensure to commit and push  '2/contrib/openshift/bundle-plugins.txt'"
  exit 1
fi
