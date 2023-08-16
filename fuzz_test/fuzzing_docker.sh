#!/usr/bin/env bash
# /*******************************************************************************
#  * Copyright 2023 Intel Corporation.
#  *
#  * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
#  * in compliance with the License. You may obtain a copy of the License at
#  *
#  * http://www.apache.org/licenses/LICENSE-2.0
#  *
#  * Unless required by applicable law or agreed to in writing, software distributed under the License
#  * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
#  * or implied. See the License for the specific language governing permissions and limitations under
#  * the License.
#  *******************************************************************************/


EDGEX_PROJECT_NAME=${1}
echo $EDGEX_PROJECT_NAME
SWAGGER_FILE_NAME_PATH=${2}
echo $SWAGGER_FILE_NAME_PATH

SWAGGER_FILE_PATH="/restler-fuzzer/openapi"

usage()
{
  echo "Usage:"
  echo "./fuzzing_docker.sh <EDGEX_PROJECT_NAME> <SWAGGER_FILE_NAME_PATH>"
  echo
  echo "<EDGEX_PROJECT_NAME> is required, options: all|core-data|core-command|core-metadata|support-notifications|support-scheduler"
  echo "<SWAGGER_FILE_NAME_PATH> is required for NOT \"all\" EDGEX_PROJECT_NAME, it is the path and filename of a project swagger file"
  exit 1
}

runFuzzLeanPerSwagger() {
    echo "--compile from swagger file: $2"
    ./restler_bin/restler/Restler compile --api_spec "$2"

    echo "--test the grammar"
    ./restler_bin/restler/Restler test --grammar_file ./Compile/grammar.py --dictionary_file ./Compile/dict.json --settings ./Compile/engine_settings.json --no_ssl

    # assuming edgex service is already running on host
    echo "--run fuzz-lean"
    ./restler_bin/restler/Restler fuzz-lean --grammar_file ./Compile/grammar.py --dictionary_file ./Compile/dict.json --settings ./Compile/engine_settings.json --no_ssl

    echo "--copy result logs into $1"
    mkdir -p /fuzz_results/"$1"
    cp -r ./FuzzLean/ /fuzz_results/"$1"/
}

if [ "$EDGEX_PROJECT_NAME" == "" ]
then
    echo "Please provide a valid project name."
    usage
fi
if [ "$EDGEX_PROJECT_NAME" == "all" ]
then
    echo "fuzz-lean for all swagger files"

    for swagger in "$SWAGGER_FILE_PATH"/*
    do
        projectname=$(basename "$swagger" .yaml)
        echo "$projectname"
        echo "$swagger"
        if [[ "$projectname" != *"."* ]]
        then
            runFuzzLeanPerSwagger $projectname $swagger
        fi
    done
else
    echo "fuzz-lean a specific swagger file only"
    runFuzzLeanPerSwagger $EDGEX_PROJECT_NAME $SWAGGER_FILE_NAME_PATH
fi

