#!/usr/bin/env groovy

def OASVERSION = '3.0.0'
def ISPRIVATE = true
def OWNER = 'EdgeXFoundry1'
def APIVERSION = ''
def APIFOLDER = ''
node {
    stage ('Checkout') {
        checkout scm
    }

    withCredentials([string(credentialsId: 'swaggerhub-api-key', variable: 'APIKEY')]) {
        APIVERSION = '1.1.1'
        APIFOLDER = 'v1.1.1'
        sh "sh toSwaggerHub.sh ${APIKEY} ${APIFOLDER} ${APIVERSION} ${OASVERSION} ${ISPRIVATE} ${OWNER}"
    }
    withCredentials([string(credentialsId: 'swaggerhub-api-key', variable: 'APIKEY')]) {
        APIVERSION = '2.0.0'
        APIFOLDER = 'v2'
        sh "sh toSwaggerHub.sh ${APIKEY} ${APIFOLDER} ${APIVERSION} ${OASVERSION} ${ISPRIVATE} ${OWNER}"
    }
}
