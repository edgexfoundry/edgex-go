pipeline {
    agent { label 'ubuntu18.04-docker-8c-8g' }
    stages {
        stage('EdgeX Build') {
            stages {
                stage('Build amd64') {
                    environment {
                        ARCH = 'x86_64'
                    }
                    stages {
                        stage('Prep') {
                            steps {
                                script {
                                    sh "docker pull nexus3.edgexfoundry.org:10003/edgex-devops/edgex-golang-base:1.20-alpine"
                                    sh "docker tag nexus3.edgexfoundry.org:10003/edgex-devops/edgex-golang-base:1.20-alpine ci-base-image-${env.ARCH}"
                                    docker.image("ci-base-image-${env.ARCH}").inside('-u 0:0') { sh 'go version' }
                                }
                            }
                        }

                        stage('Test') {
                            steps {
                                script {
                                    docker.image("ci-base-image-${env.ARCH}").inside('-u 0:0 -v /var/run/docker.sock:/var/run/docker.sock') {
                                        // fixes permissions issues due new Go 1.18 buildvcs checks
                                        // sh 'git config --global --add safe.directory $WORKSPACE'

                                        sh 'make test'
                                    }
                                }
                            }
                        }

                        stage('Build') {
                            environment {
                                BUILDER_BASE = "ci-base-image-${env.ARCH}"
                            }
                            steps {
                                sh 'make -j4 docker'
                            }
                        }
                    }
                }
            }
        }
    }
}