#!/bin/sh

# Copy in service RAML files 

cp ../core/data/raml/*.raml ./core/data/
cp ../core/metadata/raml/*.raml ./core/metadata/
cp ../core/command/raml/*.raml ./core/command/
cp ../support/logging/raml/*.raml ./support/logging/
cp ../export/client/raml/*.raml ./export/client/

# Build image (copying in documentation sources)

docker build -t doc-builder:latest -f Dockerfile.build .
if [ -d _build ]
then
  rm -rf _build
fi
mkdir _build

# Build documentation in container

docker run --name builder doc-builder:latest

# Copy built documentation out of container

docker cp builder:docbuild/_build/html _build/

# Clean up

rm ./core/data/*.raml
rm ./core/metadata/*.raml
rm ./core/command/*.raml
rm ./support/logging/*.raml
rm ./export/client/*.raml
docker rm builder
docker image prune -f
