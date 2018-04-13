#!/bin/sh
set -e

# Copy in service RAML files 

cp ../core/data/raml/*.raml ./core/data/
cp ../core/metadata/raml/*.raml ./core/metadata/
cp ../core/command/raml/*.raml ./core/command/
cp ../support/logging/raml/*.raml ./support/logging/
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
