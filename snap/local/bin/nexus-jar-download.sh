# example usage:
# $ download_jar_file support-scheduler 0.7.0 staging
download_jar_file() 
{
    export SERVICE=$1
    export VERSION=$2
    export REPO=$3

    export JAR_FILE=$SERVICE-$VERSION.jar

    # download the jar file and the md5 sum for it
    curl https://nexus.edgexfoundry.org/content/repositories/$REPO/org/edgexfoundry/$SERVICE/$VERSION/$JAR_FILE \
        -s -S -f -o $JAR_FILE 
    curl https://nexus.edgexfoundry.org/content/repositories/$REPO/org/edgexfoundry/$SERVICE/$VERSION/$JAR_FILE.md5 \
        -s -S -f -o $JAR_FILE.md5

    # confirm that the sum matches - note we can't use `md5sum -c` here because 
    # the md5sum file is just the hash, not the filename and so md5sum doesn't accept it
    sum=$(md5sum $JAR_FILE | awk '{print $1}')
    if [ "$sum" = "$(cat $JAR_FILE.md5)" ]; then
        # posix sh can't do "!=", so we just have an empty block here
        echo ""
    else
        echo "invalid md5sum for file $JAR_FILE"
        exit 1
    fi

    # install the jar file into the part install directory
    install -d "$SNAPCRAFT_PART_INSTALL/jar/$SERVICE"
    mv "$JAR_FILE" "$SNAPCRAFT_PART_INSTALL/jar/$SERVICE/$SERVICE.jar"
}
