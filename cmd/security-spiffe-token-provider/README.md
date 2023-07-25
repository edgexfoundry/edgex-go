# Notes for testing out the spiffe-token-provider

Developers can utilize the docker-compose with edgex-security-test-client running together with spiffe related services.
To start the spiffe related services, please go to `edgex-compose` repository and use the `make run` utility under
the `compose-builder` directory.  The spiffe related services can be started with adding an option `delayed-start` after
`make run`; e.g. `make run dev delayed-start`.
Once spiffe related services are started, then use a command line console to do the following steps:

1. docker exec -ti edgex-security-test-client sh -l
1. apk update && apk --no-cache --update add curl
1. cd /tmp
1. spire-agent api fetch x509 -socketPath /tmp/edgex/secrets/spiffe/public/api.sock -write /tmp
1. curl -kiv -d service_key=security-test-client -d known_secret_names=redisdb --cert svid.0.pem --key svid.0.key -X POST "https://edgex-security-spiffe-token-provider:59841/api/v3/gettoken?raw_token=true"

If https request with curl command above runs successfully, we should see the secret store raw token returned.
