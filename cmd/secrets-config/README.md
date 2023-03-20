% secrets-config-proxy(1) User Manuals secrets-config-proxy(1)

# NAME

secrets-config-proxy – Configure EdgeX API gateway service

# SYNOPSIS

**secrets-config proxy** SUBCOMMAND [OPTIONS]

# DESCRIPTION

Configures the EdgeX API gateway service.

This command is used to configure the TLS certificate for external connections, create authentication tokens for inbound proxy access, and other related utility functions.

Proxy configuration commands (listed below) require access to the secret store master key in order to generate temporary secret store access credentials.

# OPTIONS

  * **--configDir** _/path/to/directory/with/configuration.yaml_ (optional)

    Points to directory containing a configuration.yaml file.

!!! edgey "EdgeX 3.0"
    The `--confdir` command line option is replaced by `--configDir` in EdgeX 3.0.

# SUBCOMMANDS

  * **tls**

    Configure inbound TLS certificate.
    This command will replace the default TLS certificate created with EdgeX is started for the first time.
    Requires additional arguments:

    * **--inCert** _/path/to/certchain_ (required)

      Path to TLS leaf certificate (PEM-encoded x.509) (the file extension is arbitrary).
      If intermediate certificates are required to chain to a certificate authority,
      these should also be included.
      The root certificate authority should not be included.

    * **--inKey** _/path/to/private\_key_ (required)

      Path to TLS private key (PEM-encoded).

    * **--keyFilename** _filename_ (optional)

    	Filename of private key file (on target (default "nginx.key")

    * **--targetFolder** _directory-path_ (optional)

    	Path to TLS key file (default "/etc/ssl/nginx")

  * **adduser**

    Create an API gateway user by creating a user identity the EdgeX secret store.
    Requires additional arguments:

    * **--user** _username_ (required)

      Username of the user to add.

    * **--jwtTTL** _duration_ (optional)

    	JWT created by vault identity provider lasts this long (_s, _m, _h, or _d, seconds if no unit) (default "1h")

      Clients have up to `tokenTTL` time available to exchange the secret store token for a signed JWT.
      The validity period of that JWT is governed by `jwtTTL`.

    * **--tokenTTL** _duration_ (optional)

    	Vault token created as a result of vault login lasts this long  (_s, _m, _h, or _d, seconds if no unit) (default "1h")

      The `adduser` command creates a credential that enables a use to request a token for the secret store.
      The intended purpose of this token is to exchange it for a signed JWT.
      The duration specified here governs the time period within which a signed JWT can be requested.

      Note that although these tokens are renewable, there is nothing to be done with the token
      except for requesting a JWT. Thus, the token renew endpoint is not currently exposed externally.

    * **--useRootToken** (optional)

      Normally, `secrets-config` uses a service token in the secret store token file.
      As this token expires from inactivity an hour after it is created,
      it is possible to point `secrets-config` at a `resp-init.json`
      and a root token will be created afresh from the key shares in that file.
      The `--useRootToken` flag is used to tell `secrets-config`
      to use this authentication method to talk to the EdgeX secret store.

    Upon completion, `adduser` returns a JSON object with a random `password` field set.
    This password is generated from the kernel random source and overwrites any previous password set on the account.

    A sample shell script to turn this into an token that can be used for API gateway authentication is as follows:

    ```shell
    password=password-from-above

    vault_token=$(curl -ks "http://localhost:8200/v1/auth/userpass/login/${username}" -d "{\"password\":\"${password}\"}" | jq -r '.auth.client_token')

    id_token=$(curl -ks -H "Authorization: Bearer ${vault_token}" "http://localhost:8200/v1/identity/oidc/token/${username}" | jq -r '.data.token')

    echo "${id_token}"
    ```

    It is expected the the username/password returned from adduser will be saved for later use.
    However, if the password is lost, adduser can be run a second time to reset the password.

  * **deluser**

    Delete a API gateway user. Requires additional arguments:

    * **--user** _username_ (required)

      Username of the user to delete.


  * **jwt**

!!! edgey "EdgeX 3.0"
    The `jwt` sub-command is no longer supported in EdgeX 3.0.


# CONFIGURATION

# ENVIRONMENT

  * **IKM\_HOOK**

    Enables decryption of an encrypted secret store master key by pointing at an executable that returns an encryption seed that is formatted as a hex-encoded (typically 32-byte) string to its stdout.
    This optional feature, if enabled, requires pointing at the same executable that was used
    by security-secretstore-setup to provision and unlock the EdgeX the secret store.

# SEE ALSO

secrets-config(1)

EdgeX Foundry Last change: 2023
