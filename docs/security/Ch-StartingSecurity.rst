########################################
Starting security services within EdgeX
########################################

Similar to other EdgeX services, the security service can be started with Docker Compose. The security services can be started automatically with docker-compose up –d with proper docker compose file. An working sample docker compose file can be found from the edgex repo of Github at https://github.com/edgexfoundry/security-api-gateway/ . If the user prefers to start the security service manually, the commands are described below. 
docker-compose up -d volume
docker-compose up -d config-seed
docker-compose up -d consul
docker-compose up -d vault
docker-compose up -d vault-worker
docker-compose up -d kong-db
docker-compose up -d kong-migrations
docker-compose up -d kong
docker-compose up -d edgex-proxy

