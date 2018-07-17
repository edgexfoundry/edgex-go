#!/bin/sh
set -e

# Copy in service RAML files 

cp ../api/raml/core-data.raml ./core/data/
cp ../api/raml/core-metadata.raml ./core/metadata/
cp ../api/raml/core-command.raml ./core/command/
cp ../api/raml/support-logging.raml ./support/logging/
cp ../export/client/raml/*.raml ./export/client/

# Build image (copying in documentation sources)

docker build -t doc-builder:latest -f Dockerfile.build .
rm -rf _build
mkdir _build

# Build documentation in container

docker run --rm -v "$(pwd)"/_build:/docbuild/_build doc-builder:latest

# Clean up

rm ./core/data/*.raml
rm ./core/metadata/*.raml
rm ./core/command/*.raml
rm ./support/logging/*.raml
rm ./export/client/*.raml
