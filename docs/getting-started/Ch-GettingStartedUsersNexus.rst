#################################################
Getting Docker Images from Edgex Nexus Repository
#################################################

In some cases, it may be necessary to get your EdgeX container images from the Nexus Repository (managed by the Linux Foundation) as opposed to the Docker Hub repository.  Nexus is a repository that contains all the staging and development containers and images for the EdgeX Foundry project.  Containers destined for Docker Hub actually move through the Nexus Repository on their way to Docker Hub.

Some reasons why you might want to use container images from the Nexus Repos might include when:

a) the container is not available from Docker Hub (or Docker Hub is down temporarily)

b) you need the latest development build container.

c) you are working in a Windows or non-Linux environment and you are unable to build a container without some issues (Docker shell scripts may not work in Docker For Windows due to CR-LF on Git in Windows).

In order to get containers from the Nexus Repository, follow these steps:

**Pull the container(s)**

Pull the container(s) from Nexus into your local environment.  In the example below, the docker-core-config-seed container image is pulled from Nexus.  Note the host name (nexus3.edgexfoundry.org) and port (10004) for pulling

$ docker pull nexus3.edgexfoundry.org:10004/docker-core-config-seed

**Replace the image(s) in docker-compose.yml**

A Docker Compose file that pulls the latest EdgeX container images from Nexus is available here:  https://github.com/edgexfoundry/developer-scripts/blob/master/compose-files/docker-compose-nexus.yml.
If you are creating your own Docker Compose file or want to use and existing EdgeX Docker Compose file but selectively use Nexus images, replace the name/location of the Docker image in your docker-compose.yml file for the containers you want to get from Nexus versus Docker Hub.  For example, the config-seed item in docker-compose.yml might ordinarily look like this:

::

   config-seed:
     **image: edgexfoundry/docker-core-config-seed**
    
     ports:
         \- "8400:8400"\
         \- "8500:8500"\
         \- "8600:8600"\
    
     container_name: edgex-config-seed
    
     hostname: edgex-core-config-seed
    
     networks:
       edgex-network:
         aliases:
             \- edgex-core-consul\
     volumes_from:
       \- volume\
     depends_on:
       \- volume\

Change the "image" field to point to the Nexus Repos

::

   config-seed:
     **image: nexus3.edgexfoundry.org:10004/docker-core-config-seed**
      ports:
         \- "8400:8400"\
         \- "8500:8500"\
         \- "8600:8600"\
    
     container_name: edgex-config-seed
    
     hostname: edgex-core-config-seed
  
     networks:
       edgex-network:
         aliases:
             \- edgex-core-consul\
     volumes_from:
       \- volume\
     depends_on:
       \- volume\

Save the Docker Compose YAML file after you make any changes.

**Use the image(s)**

Now start your container(s) as you would normally with Docker Compose

$ docker-compose up -d [container_name]

