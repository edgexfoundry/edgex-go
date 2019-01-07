#!/bin/bash

# TODO: it would be better if this func first tried SIGTERM,
# and fell back to SIGKILL, however this script is not meant
# to be long-lived, deprecated once all of the security services
# are independent daemons and systemd will do this for us
kill_service() {
	echo "sending SIGKILL to $2 ($1) service"
	kill -9 "$1" || true
}

# send SIGINT to consul
int_service() {
	echo "sending SIGINT to $2 ($1) service"
	kill -INT "$1" || true
}

# stop kong - note this might fail, so we just force it to always succeed with "|| true"
${SNAP}/bin/kong-wrapper.sh stop -p ${SNAP_DATA}/kong || true

# killing nginx is tricky, as we need to make sure we don't inadvertantly kill any other nginx
# processes, so we use pgrep with a more specific pattern
pid=`pgrep -f "nginx.*${SNAP_DATA}/kong"` || true
if [ ! -z $pid ] && [ $pid != "" ] ; then
    int_service $pid "nginx parent"
fi

# finally for kong also remove the kong_env file so that file doesn't stop the start script from working again
rm -f ${SNAP_DATA}/kong/.kong_env

# kill vault-worker, as it might get stuck in an infinite loop trying to unseal the vault
pid=`pgrep -f "bin/vault-worker.sh"` || true
if [ ! -z $pid ] && [ $pid != "" ] ; then
    kill_service $pid "vault worker"
fi

# send sigint to vault to shut it down
pid=`pgrep -f "bin/vault.*config/security-secret-store/vault-config.json"` || true
if [ ! -z $pid ] && [ $pid != "" ] ; then
    int_service $pid "vault"
fi
