#!/bin/bash -e

if [ -z "$SNAPCRAFT_PROJECT_DIR" ]; then
    echo "Error: SNAPCRAFT_PROJECT_DIR not set!"
    exit 1
fi

pushd $SNAPCRAFT_PROJECT_DIR

if git describe ; then
    # get the short tag without commit count and hash when there are untagged commits
    TAG=$(git describe --tags --abbrev=0) 
    # drop the v prefix
    VERSION=$(echo $TAG | sed 's/v//')
    # count the number of commits after the tag
    COMMITS_COUNT=$(git rev-list --count $TAG..HEAD)
else
    VERSION="0.0.0"
fi

# When there are commits after the tag, add the number of commits as semver
# build metadata: https://semver.org/#spec-item-10
if [ "$COMMITS_COUNT" != "0" ]; then  
    VERSION="${VERSION}+${COMMITS_COUNT}"   
fi

echo "Version: $VERSION"

# set snap version
snapcraftctl set-version ${VERSION}

# write to file for the build using Makefile
echo $VERSION > ./VERSION

popd
