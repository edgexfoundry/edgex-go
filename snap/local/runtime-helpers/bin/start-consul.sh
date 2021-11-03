#!/bin/bash -e

echo "$(date) deploying the default EdgeX configuration for Consul"
# the default Consul local configuration is applied to all cases no matter ACL is enabled or not
# note that Consul's DNS port is disabled based on the securing Consul ADR
# https://github.com/edgexfoundry/edgex-docs/blob/master/docs_src/design/adr/security/0017-consul-security.md#phase-1
cat > "$SNAP_DATA/consul/config/consul_default.json" <<EOF
{
    "node_name": "edgex-core-consul",
    "enable_local_script_checks": true,
    "disable_update_check": true,
    "ports": {
      "dns": -1
    }
}
EOF

acls=${EDGEX_SECURITY_SECRET_STORE:-true}
logger "start-consul.sh: acls=$acls"

if [ "$acls" == "true" ]; then
    echo "$(date) deploying additional ACL configuration for Consul"
    cat > "$SNAP_DATA/consul/config/consul_acl.json" <<EOF
{
      "acl": {
      "enabled": true,
      "default_policy": "deny",
      "enable_token_persistence": true
    }
}
EOF
fi

# start consul in the background
"$SNAP/bin/consul" agent \
    -data-dir="$SNAP_DATA/consul/data" \
    -config-dir="$SNAP_DATA/consul/config" \
    -server -bind=127.0.0.1 -bootstrap -ui &

# loop trying to connect to consul, as soon as we are successful exit
# NOTE: ideally consul would be able to notify systemd directly, but currently 
# it only uses systemd's notify socket if consul is _joining_ another cluster
# and not when bootstrapping
# see https://github.com/hashicorp/consul/issues/4380

# Note: we no longer loop trying to connect to consul here as it is already
# taken care by security-consul-bootstrapper, in which it actually waits for
# the consul leader being elected
# see details in https://github.com/edgexfoundry/edgex-go/blob/master/internal/security/bootstrapper/command/setupacl/command.go#L117-L131
# this is to avoid the chicken-and-egg problem when it is running in "deny" policy mode
# as the consul token being required for the service checking API to be able to talk to consul
# in tandem with non-blocking startup of security-consul-bootstrapper
