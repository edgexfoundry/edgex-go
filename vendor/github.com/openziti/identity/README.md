Ziti Identity library
---------------------

This library is an attempt to normalize identity configuration for various ziti components. 

# Configuration

It is expected that identity configuration is stored in JSON format and mapped to `identity.IdentityConfig` type
```json
{
    "id": {
        "key": "file://{path}",
        "cert": "file://{path}",
        "server_cert": "file://{path}" // optional
        "ca": "file://{path}" // optional
    }
}
```

It allows different ways of specifying private keys and certificates
### Keys
* from file `"key": "file://{path to key PEM file}"`, or `"key": "{path to key PEM file}"`.
Note, latter version supports relative paths
* inline `"key": "pem:------BEGIN EC PRIVATE KEY-----...."`
* engine for HW token support `"key": "engine:{engine_id}?{engine options}"`

### Certificates
Applied to both ID/client and server certificates, as well as CA bundle config
* from file `"cert": "file://{path to cert PEM file}"`, or `"server_cert": "{path to key PEM file}"`.
                                                      Note, latter version supports relative paths
* inline `"cert": "pem:------BEGIN CERTIFICATE-----...."`

# Usage
Once `IdentityConfig` is loaded, it could be used to acquire actual TLS credentials
```go
idCfg := cfg.ID // load config from somewhere
id, err := identity.LoadIdentity(idCfg)

cltCert = id.Cert() // tls.Certificate
```
