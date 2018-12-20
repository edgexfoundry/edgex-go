#!/bin/bash -e

# generate tls keys in vault's dir
cd "$SNAP_DATA/vault"
# check to see if the root certificate generated with pkisetup already exists, if so then don't generate new certs
# note that this assumes::
# * that the pkisetup-vault.json file exists and is valid json
# * that if the root ca file still exists then the other certificates still exists
# * that the root ca file name is located at $working_dir/$pki_setup_dir/$ca_name/$ca_name.pem (this is true for the current release)
CERT_DIR=$(jq -r '.working_dir' "$PKI_SETUP_VAULT_FILE")
CERT_SUBDIR=$(jq -r '.pki_setup_dir' "$PKI_SETUP_VAULT_FILE")
ROOT_NAME=$(jq -r '.x509_root_ca_parameters | .ca_name' "$PKI_SETUP_VAULT_FILE")
if [ ! -f "$CERT_DIR/$CERT_SUBDIR/$ROOT_NAME/$ROOT_NAME.pem" ]; then
     "$SNAP/bin/pkisetup" --config "$PKI_SETUP_VAULT_FILE"
     "$SNAP/bin/pkisetup" --config "$PKI_SETUP_KONG_FILE"
fi
