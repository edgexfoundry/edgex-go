#!/bin/bash

apiKey=$1
apiFolder=$2
apiVersion=$3
oasVersion=$4
isPrivate=$5
owner=$6
apiPath="api/openapi/"${apiFolder}

for file in ${apiPath}/*.yaml; do
    apiName="$(echo $file | cut -d "/" -f 4 | cut -d "." -f 1)"
    echo "API_Name:"$apiName
    apiContent="$(cat ${apiPath}/${apiName}.yaml)"
    
    curl -X POST "https://api.swaggerhub.com/apis/${owner}/${apiName}?isPrivate=${isPrivate}&version=${apiVersion}&oas=${oasVersion}&force=true" -H "accept:application/json" -H "Authorization:${apiKey}" -H "Content-Type:application/yaml" -d "${apiContent}"
done
