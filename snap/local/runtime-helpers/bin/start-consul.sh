#!/bin/bash -e

echo "$(date) deploying the default EdgeX configuration for Consul"
# the default Consul local configuration is applied to all cases no matter ACL is enabled or not
# note that Consul's DNS port is disabled based on the securing Consul ADR
# https://github.com/edgexfoundry/edgex-docs/blob/master/docs_src/design/adr/security/0017-consul-security.md#phase-1
cat > "$SNAP_DATA/consul/config/consul_default.json" <<EOF
{
    "enable_local_script_checks": true,
    "disable_update_check": true,
    "ports": {
      "dns": -1
    }
}
EOF

echo "$(date) ENABLE_REGISTRY_ACL = ${ENABLE_REGISTRY_ACL}"

# if feature flag ENABLE_REGISTRY_ACL is true, then we need to add additional configuration settings to Consul's ACL system
# according to the securing Consul ADR, we set the "default_policy" to "allow" in Phase 1
if [ "${ENABLE_REGISTRY_ACL}" == "true" ]; then
    echo "$(date) deploying additional ACL configuration for Consul"
    cat > "$SNAP_DATA/consul/config/consul_acl.json" <<EOF
{
    "acl": {
      "enabled": true,
      "default_policy": "allow",
      "enable_token_persistence": true
    }
}
EOF
fi

# start consul in the background
"$SNAP/bin/consul" agent \
    -data-dir="$SNAP_DATA/consul/data" \
    -config-dir="$SNAP_DATA/consul/config" \
    -server -client=0.0.0.0 -bind=127.0.0.1 -bootstrap -ui &

# loop trying to connect to consul, as soon as we are successful exit
# NOTE: ideally consul would be able to notify systemd directly, but currently 
# it only uses systemd's notify socket if consul is _joining_ another cluster
# and not when bootstrapping
# see https://github.com/hashicorp/consul/issues/4380

# to actually test if consul is ready, we simply check to see if consul 
# itself shows up in it's service catalog
# also note we don't have a timeout here because we use start-timeout for this
# daemon so systemd will kill us if we take too long waiting for this
CONSUL_URL=http://localhost:8500/v1/catalog/service/consul
until [ -n "$(curl -s $CONSUL_URL | jq -r '. | length')" ] && 
    [ "$(curl -s $CONSUL_URL | jq -r '. | length')" -gt "0" ] ; do
    sleep 1
done
