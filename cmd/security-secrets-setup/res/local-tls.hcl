listener "tcp" { 
  address = "edgex-vault:8200" 
  tls_disable = "0" 
  cluster_address = "edgex-vault:8201"
  tls_min_version = "tls12"
  tls_client_ca_file ="/vault/config/pki/EdgeXFoundryCA/EdgeXFoundryCA.pem"
  tls_cert_file ="/vault/config/pki/EdgeXFoundryCA/edgex-vault.pem"
  tls_key_file = "/vault/config/pki/EdgeXFoundryCA/edgex-vault.priv.key"
}

backend "consul" {
  path = "vault/"
  address = "edgex-core-consul:8500"
  scheme = "http"
  redirect_addr = "https://edgex-vault:8200"
  cluster_addr = "https://edgex-vault:8201"
}

default_lease_ttl = "168h"
max_lease_ttl = "720h"
