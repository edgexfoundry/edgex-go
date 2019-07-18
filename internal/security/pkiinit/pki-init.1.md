% pki-init(1) Version 1.0 | PKI initialization for EdgeX Foundry secret management subsystem

NAME
====

**pki-init** — Creates an on-device public-key infrastructure (PKI) to secure microservice secret management

SYNOPSIS
========

| **pki-init** \[**-generate**|**-cache**|**-cacheca**|**-import**]
| **pki-init** \[**-h**|**--help**|**]

DESCRIPTION
===========

The Vault secret management component of EdgeX Foundry requires TLS encryption of secrets over the wire via a pre-created PKI.  In Docker-based installations of EdgeX Foundry, this component carries out the installation of a PKI that is for transport-level encryption of application secrets and is designed to run in its own microservice. pki-init supports several modes of operation as defined in the Options section.

As the PKI is security-sensitive this tool takes a number of precautions to safeguard the PKI:
* The underlying `pkisetup` utility is invoked to generate the PKI into a subfolder of $XDG_RUNTIME_DIR, which is a non-persistent tmpfs volume.
* The resulting PKI is deployed on a per-service basis to /run/edgex/secrets/pki/_service-name_, which is assumed to also be a non-persistent tmpfs volume.
* The CA private key is shred (securely erased) prior to caching, if caching is wanted.

Options
-------

-h, --help

:   Prints brief usage information.

-generate

:   Causes a PKI to be generated afresh every time (typically, this will be whenever the framework is started).

-cache

:   Causes a PKI to be generated exactly once and then copied to a designated cache location for future use.  The private key of the CA certificate is not cached to deter unauthorized creation of additional end-entity TLS certificates.

-cacheca

:   Causes a PKI to be generated exactly once and then copied to a designated cache location for future use.  The private key of the CA certificate is cached as well to support the possibility of after-the-fact creation of new TLS end-entity certificates.

-import

:   Suppresses all PKI generation functionality and tells pki-init to assume that there is a Docker volume that the cache volume contains a pre-populated PKI such as a Kong certificate signed by an external certificate authority or TLS keys signed by an offline enterprise certificate authority.

FILES
=====

*/run/edgex/secrets/***

:   Target deployment folder for the PKI secrets. Populated with subdirectories named after EdgeX services (e.g. `edgex-vault`) and contains typically two files: `server.crt` for a PEM-encoded end-entity TLS certificate and the corresponding private key in `server.key` as well as a sentinel value `.pki-init.complete`.  The special service name "ca" is reserved for the CA certificate and private key.  

*pkisetup*

:   EdgeX utility application to generate a CA and TLS leaf certificates.  Expected to be in the working directory where `pki-init` is launched.

*pkisetup-vault.json*

:   Configuration file for `pkisetup` to generate Vault's PKI.  Expected to be in the working directory where `pki-init` is launched.

ENVIRONMENT
===========

**XDG_RUNTIME_DIR**

:   A runtime scratch area used for PKI generation.  The `pkisetup` utility is invoked into `$XDG_RUNTIME_DIR/edgex/pki-init/setup` and the reformatted and stored in `$XDG_RUNTIME_DIR/edgex/pki-init/generated`

**PKI_CACHE**

:   If unspecified defaults to `/etc/edgex/pki` and should be mapped to a persistent Docker volume.  The `-cache` option caches PKI into here, and `-import` deploys a PKI from here.

NOTES
=====

As pki-init is a helper utility to ensure that a PKI is created on first launch, it is intended that pki-init is always invoked with the same operation flag, such as `-generate` or `-cache` or `-import`.   It is not possible to change from `-cache` to `-cacheca` or vice-versa (the tool will detect this and output a warning).  Changing from `-cache` to `-generate` will cause the cache to be ignored when deploying a PKI and changing it back will cause a reversion to a stale CA.  Changing from `-cache` to `-import` mode of operation is not noticeable by the tool--the PKI that is in the cache will be the one deployed.  To force regeneration of the PKI cache after the first launch, the PKI cache must be manually cleaned.

BUGS
====

See GitHub Issues: <https://github.com/edgexfoundry/security-secret-store/issues>

SEE ALSO
========

**pkisetup(1)**