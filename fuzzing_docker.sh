#!/usr/bin/env bash

EDGEX_PROJECT_NAME=${1}
echo $EDGEX_PROJECT_NAME
SWAGGER_FILE_PATH=${2}
echo $SWAGGER_FILE_PATH
SWAGGER_FILE_NAME=${3}
echo $SWAGGER_FILE_NAME

echo "--compile from swagger file"
./restler_bin/restler/Restler compile --api_spec /$SWAGGER_FILE_PATH/$SWAGGER_FILE_NAME

echo "--test the grammar"
./restler_bin/restler/Restler test --grammar_file ./Compile/grammar.py --dictionary_file ./Compile/dict.json --settings ./Compile/engine_settings.json --no_ssl

# assuming edgex service is already running on host
echo "--run fuzz-lean"
./restler_bin/restler/Restler fuzz-lean --grammar_file ./Compile/grammar.py --dictionary_file ./Compile/dict.json --settings ./Compile/engine_settings.json --no_ssl

echo "--copy result logs into $EDGEX_PROJECT_NAME"
mkdir -p /fuzz_result/$EDGEX_PROJECT_NAME
cp -r ./Test/RestlerResults/ /fuzz_result/$EDGEX_PROJECT_NAME/
