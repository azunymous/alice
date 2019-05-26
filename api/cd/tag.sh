#!/bin/bash
SCRIPTPATH="$( cd "$(dirname "$0")" ; pwd -P )"
cd ${SCRIPTPATH}
cd ../

VERSION=`git describe --tags --abbrev=0 | awk -F. '{$NF+=1; OFS="."; print $0}'`
git tag $VERSION
git push origin --tags