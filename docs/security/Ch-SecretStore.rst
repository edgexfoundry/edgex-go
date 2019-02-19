###################
Secret Store
###################

There are all kinds of secrets used within EdgeX Foundry micro services, such as tokens, passwords, certificates etc. The secret store serves as the central repository to keep these secrets. The developers of other EdgeX Foundry micro services utilize the secret store to create, store and retrieve secrets relevant to their corresponding micro service. The communications the between secret store and other micro services are secured by TLS. 

Currently the EdgeX Foundry secret store is implemented with `Vault <https://www.vaultproject.io/>`_, a HashiCorp open source software product.

Vault is a tool for securely accessing secrets. A secret is anything that you want to tightly control access to, such as API keys, passwords, database credentials, service credentials, or certificates. Vault provides a unified interface to any secret, while providing tight access control and multiple authentication mechanisms (token, LDAP, etc.). Vault adds on key rolling, revocation rules, time-to-live access tokens, secure storage, Shamir Secret Sharing based unlocking mechanism, high availability and detailed auditing.

Vault can use several backend systems (filesystem, databases, Consul, Etcd, S3, Azure, etc.) to securely store every sensitive asset. The current EdgeX Foundry implementation of Vault is using `Consul <https://www.consul.io/>`_, another HashiCorp open source software product. Consul is a distributed service mesh to connect, secure, and configure services across any runtime platform and public or private cloud. Consul uses a consensus protocol to provide Consistency as defined by `CAP <https://en.wikipedia.org/wiki/CAP_theorem>`_. The consensus protocol is based on `"Raft: In search of an Understandable Consensus Algorithm" <https://ramcloud.stanford.edu/wiki/download/attachments/11370504/raft.pdf>`_. For a visual explanation of Raft, see The `Secret Lives of Data <http://thesecretlivesofdata.com/raft>`_.

The seamless integration of Vault and Consul provides a strong yet simple infrastructure to setup a reliable high availability architecture (Vault failover nodes, Consul Clustering) for the EdgeX Foundry Security services in production. 

The key features of Vault are:

    * **Secure Secret Storage:** Arbitrary key/value secrets can be stored in Vault. Vault encrypts these secrets prior to writing them to persistent storage, so gaining access to the raw storage isn't enough to access your secrets. Authentication mechanisms (internal and/or external) and authorizations based upon policies provide access management. 

    * **Dynamic Secrets:** Vault can generate secrets on-demand for some systems, automatically revoking them after the lease is up.

    * **Data Encryption:** Vault can encrypt and decrypt data without storing it. 

    * **Leasing and Renewal:** All secrets in Vault have a lease associated with them (automatic revocation). However, clients can renew leases via built-in renew APIs.

    * **Revocation:** Vault has built-in support for secret revocation. Single secrets, but also a tree of secrets. Revocation assists in key rolling as well as locking down systems in the case of an intrusion.

======================
Start the Secret Store
======================

Start the Secret Store with Docker Compose and a Docker Compose manifest file. The Docker Compose file named docker-compose-delhi-0.7.0.yml can be found at these two locations:

    * https://github.com/edgexfoundry/security-secret-store/blob/delhi/docker-compose-delhi-0.7.0.yml
    * https://github.com/edgexfoundry/developer-scripts/blob/master/compose-files/security/docker-compose-delhi-0.7.0.yml

This Compose file starts the entire EdgeX Foundry platform including the security services. 

The command to start EdgeX Foundry platform including the Secret Store and API gateway related services is:

.. code-block:: shell

    sh> docker-compose up -d

For a pristine run, it is strongly recommended to thoroughly clean up any Docker artifacts remaining from the current run.

.. code-block:: shell

    sh> docker-compose down -v

The "down" operation will remove containers, network and adding the -v option it will also remove persistent volumes:

.. code-block:: shell
    :linenos:
    :emphasize-lines: 1,9,10

    Stopping edgex-vault       ... done
    Stopping edgex-core-consul ... done
    Stopping edgex-files       ... done
    Removing edgex-vault-worker ... done
    Removing edgex-vault        ... done
    Removing edgex-config-seed  ... done
    Removing edgex-core-consul  ... done
    Removing edgex-files        ... done
    Removing network security-secret-store_edgex-network
    Removing volume security-secret-store_db-data
    Removing volume security-secret-store_log-data
    Removing volume security-secret-store_consul-config
    Removing volume security-secret-store_consul-data
    Removing volume security-secret-store_vault-config
    Removing volume security-secret-store_vault-file
    Removing volume security-secret-store_vault-logs

Troubleshooting steps
---------------------

For debugging purpose, the Secret Store services can be started individually with these commands. A pre-requisite being to carefully follow the invocation sequence in order to avoid any dependency failure. Lines starting with # (hashtag) are contextual comments to explicit the purpose of the above corresponding command:

Clean-up
^^^^^^^^

.. code-block:: shell
    :linenos:
    :emphasize-lines: 4

    sh> cd <path-to-EdgeX-Foundry-Secret-Store>
    # <...>/security-secret-store/

    sh> docker-compose down -v

    sh> docker-compose ps
    # Check no previous container is running 

    sh> docker volume ls
    # Check and remove any previous persistent and/or unused volumes
    sh> docker volume prune
    sh> docker volume rm <volume-name>

    sh> docker network ls
    # Check and remove the previous EdgeX Foundry Docker network
    sh> docker network rm edgex-network

Start the first service: volume (platform volume initializations)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: shell

    sh> docker-compose up -d volume

Sample output:

.. code-block:: shell
    :linenos:
    :emphasize-lines: 1,9

    Creating network "security-secret-store_edgex-network" with driver "bridge"
    Creating volume "security-secret-store_db-data" with default driver
    Creating volume "security-secret-store_log-data" with default driver
    Creating volume "security-secret-store_consul-config" with default driver
    Creating volume "security-secret-store_consul-data" with default driver
    Creating volume "security-secret-store_vault-config" with default driver
    Creating volume "security-secret-store_vault-file" with default driver
    Creating volume "security-secret-store_vault-logs" with default driver
    Creating edgex-files ... done

Start the second service: consul (Consul is Vault store backend)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: shell

    sh> docker-compose up -d consul

Sample output:

.. code-block:: shell
    :linenos:
    :emphasize-lines: 2

    edgex-files is up-to-date
    Creating edgex-core-consul ... done

Display and inspect consul service logs: important lines are highlighted

.. code-block:: shell

    sh> docker-compose logs consul

Sample output:

.. code-block:: none
    :linenos:
    :emphasize-lines: 6,32

    Attaching to edgex-core-consul
    edgex-core-consul  | ==> Starting Consul agent...
    edgex-core-consul  | ==> Consul agent running!
    edgex-core-consul  |            Version: 'v1.1.0'
    edgex-core-consul  |            Node ID: '371cbce6-02a8-65f6-ddea-6df5c40a4c50'
    edgex-core-consul  |          Node name: 'edgex-core-consul'
    edgex-core-consul  |         Datacenter: 'dc1' (Segment: '<all>')
    edgex-core-consul  |             Server: true (Bootstrap: false)
    edgex-core-consul  |        Client Addr: [0.0.0.0] (HTTP: 8500, HTTPS: -1, DNS: 8600)
    edgex-core-consul  |       Cluster Addr: 127.0.0.1 (LAN: 8301, WAN: 8302)
    edgex-core-consul  |            Encrypt: Gossip: false, TLS-Outgoing: false, TLS-Incoming: false
    edgex-core-consul  |
    edgex-core-consul  | ==> Log data will now stream in as it occurs:
    edgex-core-consul  |
    edgex-core-consul  |     2019/01/13 13:25:06 [DEBUG] agent: Using random ID "371cbce6-02a8-65f6-ddea-6df5c40a4c50" as node ID
    edgex-core-consul  |     2019/01/13 13:25:06 [INFO] raft: Initial configuration (index=1): [{Suffrage:Voter ID:371cbce6-02a8-65f6-ddea-6df5c40a4c50 Address:127.0.0.1:8300}]
    edgex-core-consul  |     2019/01/13 13:25:06 [INFO] raft: Node at 127.0.0.1:8300 [Follower] entering Follower state (Leader: "")
    edgex-core-consul  |     2019/01/13 13:25:06 [INFO] serf: EventMemberJoin: edgex-core-consul.dc1 127.0.0.1
    edgex-core-consul  |     2019/01/13 13:25:06 [INFO] serf: EventMemberJoin: edgex-core-consul 127.0.0.1
    edgex-core-consul  |     2019/01/13 13:25:06 [INFO] consul: Adding LAN server edgex-core-consul (Addr: tcp/127.0.0.1:8300) (DC: dc1)
    edgex-core-consul  |     2019/01/13 13:25:06 [INFO] consul: Handled member-join event for server "edgex-core-consul.dc1" in area "wan"
    edgex-core-consul  |     2019/01/13 13:25:06 [INFO] agent: Started DNS server 0.0.0.0:8600 (tcp)
    edgex-core-consul  |     2019/01/13 13:25:06 [INFO] agent: Started DNS server 0.0.0.0:8600 (udp)
    edgex-core-consul  |     2019/01/13 13:25:06 [INFO] agent: Started HTTP server on [::]:8500 (tcp)
    edgex-core-consul  |     2019/01/13 13:25:06 [INFO] agent: started state syncer
    edgex-core-consul  |     2019/01/13 13:25:06 [WARN] raft: Heartbeat timeout from "" reached, starting election
    edgex-core-consul  |     2019/01/13 13:25:06 [INFO] raft: Node at 127.0.0.1:8300 [Candidate] entering Candidate state in term 2
    edgex-core-consul  |     2019/01/13 13:25:06 [DEBUG] raft: Votes needed: 1
    edgex-core-consul  |     2019/01/13 13:25:06 [DEBUG] raft: Vote granted from 371cbce6-02a8-65f6-ddea-6df5c40a4c50 in term 2. Tally: 1
    edgex-core-consul  |     2019/01/13 13:25:06 [INFO] raft: Election won. Tally: 1
    edgex-core-consul  |     2019/01/13 13:25:06 [INFO] raft: Node at 127.0.0.1:8300 [Leader] entering Leader state
    edgex-core-consul  |     2019/01/13 13:25:06 [INFO] consul: cluster leadership acquired


Start the third service: config-seed (platform configuration initializations)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: shell

    sh> docker-compose up -d config-seed

Sample output:

.. code-block:: shell
    :linenos:
    :emphasize-lines: 3

    edgex-files is up-to-date
    edgex-core-consul is up-to-date
    Creating edgex-config-seed ... done

Display and inspect the created container states: important lines are highlighted

.. code-block:: shell

    sh> docker-compose ps

Sample output:

.. code-block:: none
    :linenos:
    :emphasize-lines: 3

        Name                     Command               State                                      Ports
    ---------------------------------------------------------------------------------------------------------------------------------------
    edgex-config-seed   /bin/sh -c /edgex/cmd/conf ...   Exit 0
    edgex-core-consul   docker-entrypoint.sh agent ...   Up       8300/tcp, 8301/tcp, 8301/udp, 8302/tcp, 8302/udp, 0.0.0.0:8400->8400/tcp,
                                                                0.0.0.0:8500->8500/tcp, 0.0.0.0:8600->8600/tcp, 8600/udp
    edgex-files         /bin/sh -c /usr/bin/tail - ...   Up


.. note::  Line 3: ``edgex-config-seed`` service has exited after successful processing (exit code 0)

Start the fourth service: vault (Vault tool)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. note:: Vault will be **uninitialized** and **unsealed** upon success. The ``vault-worker`` service will process the initialization and unsealing tasks.


.. code-block:: shell

    sh> docker-compose up -d vault

Sample output:

.. code-block:: shell
    :linenos:
    :emphasize-lines: 3

    edgex-files is up-to-date
    edgex-core-consul is up-to-date
    Creating edgex-vault ... done

Display and inspect "vault" service logs: important lines are highlighted

.. code-block:: shell

    sh> docker-compose logs vault

Sample output:

.. code-block:: none
    :linenos:
    :emphasize-lines: 4,7,10

    Attaching to edgex-vault
    edgex-vault        | ==> Vault server configuration:
    edgex-vault        |
    edgex-vault        |              Api Address: https://edgex-vault:8200
    edgex-vault        |                      Cgo: disabled
    edgex-vault        |          Cluster Address: https://edgex-vault:8201
    edgex-vault        |               Listener 1: tcp (addr: "edgex-vault:8200", cluster address: "edgex-vault:8201", tls: "enabled")
    edgex-vault        |                Log Level: info
    edgex-vault        |                    Mlock: supported: true, enabled: true
    edgex-vault        |                  Storage: consul (HA available)
    edgex-vault        |                  Version: Vault v0.10.2
    edgex-vault        |              Version Sha: 3ee0802ed08cb7f4046c2151ec4671a076b76166
    edgex-vault        |
    edgex-vault        | ==> Vault server started! Log data will stream in below:
    edgex-vault        |


.. note:: Line 4 & 7: Vault **API** endpoint on port 8200 (lines 4 and 7).

          Line 7: Vault has **TLS** enabled.

          Line 10: Vault backend storage is **Consul**.

Start the fifth service: vault-worker (Vault init/unseal process and setups)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: shell

    sh> docker-compose up -d vault-worker

Sample output:

.. code-block:: shell
    :linenos:
    :emphasize-lines: 4

    edgex-files is up-to-date
    edgex-core-consul is up-to-date
    edgex-vault is up-to-date
    Creating edgex-vault-worker ... done

Display and inspect "vault-worker" service logs: important lines are highlighted

.. code-block:: shell

    sh> docker-compose logs vault-worker

Sample output:

.. code-block:: none
    :linenos:
    :emphasize-lines: 4,5,19,23

    Attaching to edgex-vault-worker
    edgex-vault-worker | INFO: 2019/01/13 13:35:42 successful loading the rootCA cert.
    edgex-vault-worker | INFO: 2019/01/13 13:35:43 {"keys":["564b9444eebe28b393c21a4dca1e32835b7dc27f5da03b73d22b666cb20224a9"],"keys_base64":["VkuURO6+KLOTwhpNyh4yg1t9wn9doDtz0itmbLICJKk="],"recovery_keys":null,"recovery_keys_base64":null,"root_token":"01dbbae4-353a-8cdf-8189-4d50e5535a6f"}
    edgex-vault-worker | INFO: 2019/01/13 13:35:43 Vault has been initialized successfully.
    edgex-vault-worker | INFO: 2019/01/13 13:35:43 Vault has been unsealed successfully.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Vault Health Check HTTP Status: 200 OK (StatusCode: 200)
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Verifying Admin policy file hash (SHA256).
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Vault policy file checksum (SHA256): 5ce8d58cf7d931735f6532742f677c109a91a263bcefe9aef73ab2a69f4b43d3
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Reading Admin policy file.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Importing Vault Admin policy.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Import Policy Successfull.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Reading Kong policy file.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Importing Vault Kong policy.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Import Policy Successfull.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Creating Vault Admin token.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Create Token Successfull.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Creating Vault Kong token.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Create Token Successfull.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Successful on reading certificate from v1/secret/edgex/pki/tls/edgex-kong.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Cert&key are not in the secret store yet, will need to upload them.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Load cert&key pair from volume successfully, now will upload to secret store.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Trying to upload cert&key to secret store.
    edgex-vault-worker | INFO: 2019/01/13 13:35:48 Successful to add certificate to the secret store.


Display and inspect "vault" service logs: important lines are highlighted

.. code-block:: shell

    sh> docker-compose logs vault

Sample output:

.. code-block:: none
    :linenos:
    :emphasize-lines: 18,35,44,60,61

    Attaching to edgex-vault
    edgex-vault        | ==> Vault server configuration:
    edgex-vault        |
    edgex-vault        |              Api Address: https://edgex-vault:8200
    edgex-vault        |                      Cgo: disabled
    edgex-vault        |          Cluster Address: https://edgex-vault:8201
    edgex-vault        |               Listener 1: tcp (addr: "edgex-vault:8200", cluster address: "edgex-vault:8201", tls: enabled")
    edgex-vault        |                Log Level: info
    edgex-vault        |                    Mlock: supported: true, enabled: true
    edgex-vault        |                  Storage: consul (HA available)
    edgex-vault        |                  Version: Vault v0.10.2
    edgex-vault        |              Version Sha: 3ee0802ed08cb7f4046c2151ec4671a076b76166
    edgex-vault        |
    edgex-vault        | ==> Vault server started! Log data will stream in below:
    edgex-vault        |
    edgex-vault        | 2019-01-13T13:35:42.549Z [INFO ] core: security barrier not initialized
    edgex-vault        | 2019-01-13T13:35:42.551Z [INFO ] core: security barrier not initialized
    edgex-vault        | 2019-01-13T13:35:42.554Z [INFO ] core: security barrier initialized: shares=1 threshold=1
    edgex-vault        | 2019-01-13T13:35:42.575Z [INFO ] core: post-unseal setup starting
    edgex-vault        | 2019-01-13T13:35:42.583Z [INFO ] core: loaded wrapping token key
    edgex-vault        | 2019-01-13T13:35:42.583Z [INFO ] core: successfully setup plugin catalog: plugin-directory=
    edgex-vault        | 2019-01-13T13:35:42.584Z [INFO ] core: no mounts; adding default mount table
    edgex-vault        | 2019-01-13T13:35:42.585Z [INFO ] core: successfully mounted backend: type=kv path=secret/
    edgex-vault        | 2019-01-13T13:35:42.586Z [INFO ] core: successfully mounted backend: type=cubbyhole path=cubbyhole/
    edgex-vault        | 2019-01-13T13:35:42.586Z [INFO ] core: successfully mounted backend: type=system path=sys/
    edgex-vault        | 2019-01-13T13:35:42.586Z [INFO ] core: successfully mounted backend: type=identity path=identity/
    edgex-vault        | 2019-01-13T13:35:42.593Z [INFO ] core: restoring leases
    edgex-vault        | 2019-01-13T13:35:42.593Z [INFO ] rollback: starting rollback manager
    edgex-vault        | 2019-01-13T13:35:42.594Z [INFO ] expiration: lease restore complete
    edgex-vault        | 2019-01-13T13:35:42.596Z [INFO ] identity: entities restored
    edgex-vault        | 2019-01-13T13:35:42.597Z [INFO ] identity: groups restored
    edgex-vault        | 2019-01-13T13:35:42.597Z [INFO ] core: post-unseal setup complete
    edgex-vault        | 2019-01-13T13:35:42.597Z [INFO ] core: core/startClusterListener: starting listener: listener_address=172.19.0.4:8201
    edgex-vault        | 2019-01-13T13:35:42.597Z [INFO ] core: core/startClusterListener: serving cluster requests: cluster_listen_address=172.19.0.4:8201
    edgex-vault        | 2019-01-13T13:35:42.600Z [INFO ] core: root token generated
    edgex-vault        | 2019-01-13T13:35:42.600Z [INFO ] core: pre-seal teardown starting
    edgex-vault        | 2019-01-13T13:35:42.600Z [INFO ] core: stopping cluster listeners
    edgex-vault        | 2019-01-13T13:35:42.600Z [INFO ] core: shutting down forwarding rpc listeners
    edgex-vault        | 2019-01-13T13:35:42.600Z [INFO ] core: forwarding rpc listeners stopped
    edgex-vault        | 2019-01-13T13:35:43.099Z [INFO ] core: rpc listeners successfully shut down
    edgex-vault        | 2019-01-13T13:35:43.099Z [INFO ] core: cluster listeners successfully shut down
    edgex-vault        | 2019-01-13T13:35:43.100Z [INFO ] rollback: stopping rollback manager
    edgex-vault        | 2019-01-13T13:35:43.100Z [INFO ] core: pre-seal teardown complete
    edgex-vault        | 2019-01-13T13:35:43.105Z [INFO ] core: vault is unsealed
    edgex-vault        | 2019-01-13T13:35:43.105Z [INFO ] core: entering standby mode
    edgex-vault        | 2019-01-13T13:35:43.109Z [INFO ] core: acquired lock, enabling active operation
    edgex-vault        | 2019-01-13T13:35:43.134Z [INFO ] core: post-unseal setup starting
    edgex-vault        | 2019-01-13T13:35:43.135Z [INFO ] core: loaded wrapping token key
    edgex-vault        | 2019-01-13T13:35:43.135Z [INFO ] core: successfully setup plugin catalog: plugin-directory=
    edgex-vault        | 2019-01-13T13:35:43.137Z [INFO ] core: successfully mounted backend: type=kv path=secret/
    edgex-vault        | 2019-01-13T13:35:43.137Z [INFO ] core: successfully mounted backend: type=system path=sys/
    edgex-vault        | 2019-01-13T13:35:43.137Z [INFO ] core: successfully mounted backend: type=identity path=identity/
    edgex-vault        | 2019-01-13T13:35:43.137Z [INFO ] core: successfully mounted backend: type=cubbyhole path=cubbyhole/
    edgex-vault        | 2019-01-13T13:35:43.141Z [INFO ] core: restoring leases
    edgex-vault        | 2019-01-13T13:35:43.142Z [INFO ] rollback: starting rollback manager
    edgex-vault        | 2019-01-13T13:35:43.142Z [INFO ] expiration: lease restore complete
    edgex-vault        | 2019-01-13T13:35:43.143Z [INFO ] identity: entities restored
    edgex-vault        | 2019-01-13T13:35:43.143Z [INFO ] identity: groups restored
    edgex-vault        | 2019-01-13T13:35:43.144Z [INFO ] core: post-unseal setup complete
    edgex-vault        | 2019-01-13T13:35:43.144Z [INFO ] core: core/startClusterListener: starting listener: listener_address=172.19.0.4:8201
    edgex-vault        | 2019-01-13T13:35:43.144Z [INFO ] core: core/startClusterListener: serving cluster requests: cluster_listen_address=172.19.0.4:8201


.. note::  Line 18: Vault **initialization** successful.

           Line 35: Vault **root token** generated.

           Line 44: Vault **unsealing** successful.

           Line 50: Vault key/value store *secret* successfully **mounted**.

           Line 60 & 61: Vault successfully **started**.


Display and inspect the created container states: important lines are highlighted

.. code-block:: shell

    sh> docker-compose ps

Sample output:

.. code-block:: none
    :linenos:
    :emphasize-lines: 9

        Name                     Command               State                                     Ports
    ---------------------------------------------------------------------------------------------------------------------------------------
    edgex-config-seed    /bin/sh -c /edgex/cmd/conf ...   Exit 0
    edgex-core-consul    docker-entrypoint.sh agent ...   Up       8300/tcp, 8301/tcp, 8301/udp, 8302/tcp, 8302/udp,
                                                                0.0.0.0:8400->8400/tcp, 0.0.0.0:8500->8500/tcp, 0.0.0.0:8600->8600/tcp,
                                                                8600/udp
    edgex-files          /bin/sh -c /usr/bin/tail - ...   Up
    edgex-vault          docker-entrypoint.sh serve ...   Up       0.0.0.0:8200->8200/tcp
    edgex-vault-worker   ./edgex-vault-worker --ini ...   Exit 0


.. note::  Line 9: ``edgex-vault-worker`` service has exited after successful processing (exit code 0)

Display and inspect the created container volumes: important lines are highlighted

.. code-block:: shell
    :linenos:
    :emphasize-lines: 1

    sh> docker volume ls
    DRIVER              VOLUME NAME
    local               security-secret-store_consul-config
    local               security-secret-store_consul-data
    local               security-secret-store_db-data
    local               security-secret-store_log-data
    local               security-secret-store_vault-config
    local               security-secret-store_vault-file
    local               security-secret-store_vault-logs

Display and inspect the container network (``security-secret-store_edgex-network``): important lines are highlighted

.. code-block:: shell
    :linenos:
    :emphasize-lines: 1,6

    sh> docker network ls
    NETWORK ID          NAME                                  DRIVER              SCOPE
    63227826fbc7        bridge                                bridge              local
    60763abffde3        host                                  host                local
    1d236ab1dbbd        none                                  null                local
    0a7f7266d102        security-secret-store_edgex-network   bridge              local


Using Consul Web UI
-------------------

For learning and verification purposes one might use the Consul Web UI interface to gather and double check specific Vault informations.

Consul Web UI endpoint port is exposed by the Docker compose file ``docker-compose-delhi-0.7.0.yml``. EdgeX Foundry platform uses the Consul default port number 8500. It is normally not recommended to expose Consul UI port number in production, at least the UI should not be accessible from outside the platform environment. However, because all the Vault secrets are encrypted before being transmitted and stored in the Consul backend, having access to Consul is not sufficient to access any secrets, the vault data encryption/decryption key would be absolutely necessary. 

Open a Web browser on ``http://<EdgeX Consul Server>:8500/ui``.

On the screenshot below, after selecting ``SERVICES`` and ``Vault``, the UI will show the various Vault status (heartbeat and init/unseal states), coloring the boxes in green, orange or red depending on the level of importance (info, warning, error). By clicking each of the right side status indicators, more information will be accessible in order to better inspect any situation.

.. image:: EdgeX-UI-Consul-01.png

As a practical example, we are going to navigate the Consul structure for Vault in order to check if the API Gateway (Kong) TLS certificate and private key were fetched and stored accordingly during the ``vault-worker`` process. 

First select ``KEY/VALUE`` menu, and then select ``vault`` root structure:

.. image:: EdgeX-UI-Consul-02.png

We are now going to navigate deeper in the ``vault`` tree structure to reach and display the EdgeX Kong TLS assets. Continue by selecting ``logical/``:

.. image:: EdgeX-UI-Consul-03.png
    :width: 348px
    :align: center
    :height: 287px

Then select ``d7809b...`` an arbitrary UID generated and created by Consul during Vault registration:

.. image:: EdgeX-UI-Consul-04.png
    :width: 377px
    :align: center
    :height: 216px

Select ``edgex/``:

.. image:: EdgeX-UI-Consul-05.png
    :width: 419px
    :align: center
    :height: 228px

Select ``pki/``:

.. image:: EdgeX-UI-Consul-06.png
    :width: 417px
    :align: center
    :height: 222px

Select ``tls/``:

.. image:: EdgeX-UI-Consul-07.png
    :width: 418px
    :align: center
    :height: 237px

Select ``edgex-kong/``:

.. image:: EdgeX-UI-Consul-08.png
    :width: 423px
    :align: center
    :height: 233px

And we are now finally able to display the encrypted Vault secret containing the API Gateway (Kong) TLS server certificate and its corresponding private key. As you can see on the screenshot below the Vault key/value is encrypted and totally opaque to Consul, the Vault data encrytion key (DEK) would be necessary to decrypt these secrets. Each Vault secret is encrypted before being transmitted to Consul node(s). 

.. image:: EdgeX-UI-Consul-09.png


Shell Access to Consul Container and Using Consul CLI
-----------------------------------------------------

.. code-block:: json
    :linenos:
    :emphasize-lines: 5,9,12,14

    sh> docker exec -it -e PS1='\u@\h:\w \$ ' edgex-core-consul sh

    root@edgex-core-consul:/ # consul members
    Node               Address         Status  Type    Build  Protocol  DC   Segment
    edgex-core-consul  127.0.0.1:8301  alive   server  1.1.0  2         dc1  <all>

    root@edgex-core-consul:/ # consul catalog nodes
    Node               ID        Address    DC
    edgex-core-consul  e49af36a  127.0.0.1  dc1

    root@edgex-core-consul:/ # consul catalog services
    consul
    edgex-mongo
    vault

.. note::  Line 5: Shows the Consul node status ``alive`` (1 node in EdgeX default configuration).

           Line 9: Shows the Consul nodes (1 node in EdgeX default configuration).

           Lines 12-14: Show the Consul registered services.

============================
Configuring the Secret Store
============================

Vault server configuration is essentially concentrated in one JSON file named ``local.json``. This file was prepared during the Vault Docker image build process. In the eventuality of a change, the Vault server container should be accessed to then modify the JSON file. The absolute path being ``/vault/config/local.json``. To reload the new configuration simply send Vault PID a HUP signal to trigger a configuration reload. 

Sample Vault server configuration file:

.. code-block:: json
    :linenos:
    :emphasize-lines: 2,5,11,13

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

The ``listener`` clause refers to Vault server process (port, TLS and server name), the ``backend`` clause refers to the storage backend (i.e. Consul).

To modify this configuration file, execute a shell session in the running Vault container: 

.. code-block:: json
    :linenos:
    :emphasize-lines: 1

    sh> docker exec -it -e PS1='\u@\h:\w \$ ' -e VAULT_CAPATH='/vault/config/pki/EdgeXFoundryCA/EdgeXFoundryCA.pem' edgex-vault sh

    root@edgex-vault:/vault # ls -l
    total 12
    drwxr-xr-x    4 vault    vault         4096 Jan 13 13:34 config
    drwxr-xr-x    2 vault    vault         4096 Jun  7  2018 file
    drwxr-xr-x    2 vault    vault         4096 Jun  7  2018 logs

Pay attention to the ``VAULT_CAPATH`` environment variable passed to the session. This is necessary in order to run succesful Vault CLI command. Every Vault CLI command is a wrapper of the Vault HTTP API. The Vault server is configured with TLS using X.509 PKI materials generated and signed by a local self-signed CA (EdgeXFoundryCA). Therefore, in order for each Vault CLI command (or to that extent cURL commands) to verify the Vault server TLS certificate, the self-signing CA root certificate would have to be known by the CLI command interpreter. This ``VAULT_CAPATH`` variable is checked by every Vault CLI commands, alternatively each Vault CLI command can specify an option with the same certificate path if the variable is not set. 

The self-signed Root CA certificate path can be found in the Vault configuration file (see above local.json), with parameter ``tls_client_ca_file ="/vault/config/pki/EdgeXFoundryCA/EdgeXFoundryCA.pem"``.

The ``local.json`` configuration file can be read and modified within the running container:

.. code-block:: json
    :linenos:
    :emphasize-lines: 1

    root@edgex-vault:/vault # cat config/local.json
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


A sample Vault CLI command to check Vault status:

.. code-block:: json
    :linenos:
    :emphasize-lines: 1

    root@edgex-vault:/vault # vault status
    Key             Value
    ---             -----
    Seal Type       shamir
    Sealed          false
    Total Shares    1
    Threshold       1
    Version         0.10.2
    Cluster Name    vault-cluster-57b3c4ed
    Cluster ID      fe6d18bf-fa9c-0d52-3278-bca0390af023
    HA Enabled      true
    HA Cluster      https://edgex-vault:8201
    HA Mode         active

All the X.509 PKI materials including the self-signing CA are located under ``/vault/config/pki/EdgeXFoundryCA``. 

.. code-block:: json
    :linenos:
    :emphasize-lines: 1

    root@edgex-vault:/vault # ls -l config/pki/EdgeXFoundryCA/
    total 24
    -rw-r--r--    1 vault    vault          956 Dec  5 14:05 EdgeXFoundryCA.pem
    -r--------    1 vault    vault          306 Dec  5 14:05 EdgeXFoundryCA.priv.key
    -rw-r--r--    1 vault    vault          989 Dec  5 14:05 edgex-kong.pem
    -rw-------    1 vault    vault          306 Dec  5 14:05 edgex-kong.priv.key
    -rw-r--r--    1 vault    vault         1001 Dec  5 14:05 edgex-vault.pem
    -rw-------    1 vault    vault          306 Dec  5 14:05 edgex-vault.priv.key

.. note::  Line 3: self-signing root CA certificate.

           Line 4: self-signing root CA private key.

           Line 5: API Gateway (Kong) TLS server certificate.

           Line 6: API Gateway (Kong) TLS server certificate private key.

           Line 7: Vault TLS server certificate.

           Line 8: Vault TLS server certificate private key.

The CA name (EdgeXFoundryCA) was defined by the `pkisetup <https://github.com/edgexfoundry/security-secret-store/tree/master/pkisetup>`_ tool during the Vault image build process. This tool is also responsible for all the TLS server configuration and creation tasks. 

If you are willing to change any of the Vault X.509 PKI assets or configuration parameters you will have to modify the `pkisetup-vault.json <https://github.com/edgexfoundry/security-secret-store/blob/master/pkisetup/pkisetup-vault.json>`_ file and rebuild a new Vault Docker image. 

Similarly to Vault, each EdgeX Foundry service having a TLS server certificate and private key had its X.509 PKI assets generated and signed during the Vault Docker image build process. Therefore, the API Gateway (Kong) configuration file named `pkisetup-kong.json <https://github.com/edgexfoundry/security-secret-store/blob/master/pkisetup/pkisetup-kong.json>`_ would have to be modified accordingly. A new Vault Docker image would have to be built.

The Vault `Dockerfile <https://github.com/edgexfoundry/security-secret-store/blob/master/Dockerfile.vault>`_ contains the ``pkisetup`` executions, see below for a corresponding excerpt (highlighted lines):

.. code-block:: json
    :linenos:
    :emphasize-lines: 7,9

    # Create assets folder (needed for unseal key/s, root token and tmp)
    # Run CA/Vault and Kong PKI/TLS setups and peform housekeeping tasks
    RUN mkdir /vault/config/assets && \
        chown -R vault:vault /vault && \
        chmod 644 /vault/config/local.json && \
        chmod 744 pkisetup* && \
        ./pkisetup --config pkisetup-vault.json && \
        echo "" && \
        ./pkisetup --config pkisetup-kong.json && \
        chown -R vault:vault /vault/config/pki && \
    rm -f /vault/pkisetup /vault/pkisetup-vault.json /vault/pkisetup-kong.json


EdgeX Foundry Docker environment implements a basic Vault/Consul architecture that does not provide high availability guaranties. Only one Consul server and one Vault server will be running. In a more sophisticated production environment it would be possible to build a reliable high availability infrastructure regarding Consul and Vault. To facilitate the setup of a minimal failover architecture the ``security-secret-store`` repository provides a sample folder named `Full-Architecture-Prototype <https://github.com/edgexfoundry/security-secret-store/tree/master/Full-Architecture-Prototype>`_ that contains necessary materials (scripts, helpers, configurations, etc.) to achieve that goal. 

These samples describe an architecture design with two Vault servers in failover mode (active/standby), using each one a Consul client, which subsequently connects to a Consul cluster of 3 nodes (minimal Raft concensus quorum). The Consul clients and servers (nodes) have redundant paths. 

======================
Using the Secret Store
======================

1st alternative: executing a shell session in the active Vault container to run Vault CLI commands.
---------------------------------------------------------------------------------------------------

See paragraph **Configuring the Secret Store** to have more details on the ``VAULT_CAPATH`` environment variable.

See HashiCorp Vault API documentation for further details on syntax and usage (https://www.vaultproject.io/api/).

Execute a shell session in the running Vault container:

.. code-block:: shell
    :linenos:
    :emphasize-lines: 7,9

    sh> docker exec -it -e PS1='\u@\h:\w \$ ' -e VAULT_CAPATH='/vault/config/pki/EdgeXFoundryCA/EdgeXFoundryCA.pem' edgex-vault sh

Locate the ``asssets`` folder, and the ``resp-init.json`` file:

.. code-block:: shell
    :linenos:
    :emphasize-lines: 5

    root@edgex-vault:/vault # ls -l config/assets/
    total 12
    -rw-r--r--    1 root     root           366 Jan 13 13:35 admin-token.json
    -rw-r--r--    1 root     root           365 Jan 13 13:35 kong-token.json
    -rw-r--r--    1 root     root           241 Jan 13 13:35 resp-init.json

Inspect the ``resp-init.json`` file to grab the Vault Root Token:

.. code-block:: shell
    :linenos:
    :emphasize-lines: 7

    root@edgex-vault:/vault # cat config/assets/resp-init.json
    {
    "keys":["564b9444eebe28b393c21a4dca1e32835b7dc27f5da03b73d22b666cb20224a9"],
    "keys_base64":["VkuURO6+KLOTwhpNyh4yg1t9wn9doDtz0itmbLICJKk="],
    "recovery_keys":null,
    "recovery_keys_base64":null,
    "root_token":"01dbbae4-353a-8cdf-8189-4d50e5535a6f"
    }

Login to Vault using Vault CLI and the gathered Root Token:

.. code-block:: shell
    :linenos:
    :emphasize-lines: 8,12

    root@edgex-vault:/vault # vault login 01dbbae4-353a-8cdf-8189-4d50e5535a6f
    Success! You are now authenticated. The token information displayed below
    is already stored in the token helper. You do NOT need to run "vault login"
    again. Future Vault requests will automatically use this token.

    Key                Value
    ---                -----
    token              01dbbae4-353a-8cdf-8189-4d50e5535a6f
    token_accessor     4d5eabf7-8710-81b1-b6a4-9ba17fdfdeb7
    token_duration     ∞
    token_renewable    false
    token_policies     [root]

Perform an introspection ``lookup`` on the current token login:

.. code-block:: shell
    :linenos:
    :emphasize-lines: 7,15,16

    root@edgex-vault:/vault # vault token lookup
    Key                 Value
    ---                 -----
    accessor            4d5eabf7-8710-81b1-b6a4-9ba17fdfdeb7
    creation_time       1547386542
    creation_ttl        0
    display_name        root
    entity_id           n/a
    expire_time         <nil>
    explicit_max_ttl    0
    id                  01dbbae4-353a-8cdf-8189-4d50e5535a6f
    meta                <nil>
    num_uses            0
    orphan              true
    path                auth/token/root
    policies            [root]
    ttl                 0

.. note:: Lines 9 & 10: the Root Token is the only token that has no expiration enforcement rules (Time to Live TTL counter).

Perform a check on the current token login to display the corresponding capabilities (policies):

.. code-block:: shell
    :linenos:
    :emphasize-lines: 2

    root@edgex-vault:/vault # vault token capabilities 01dbbae4-353a-8cdf-8189-4d50e5535a6f
    root

Perform a ``list`` request to display the currently mounted secret backends:

.. code-block:: shell
    :linenos:
    :emphasize-lines: 6

    root@edgex-vault:/vault # vault secrets list
    Path          Type         Accessor              Description
    ----          ----         --------              -----------
    cubbyhole/    cubbyhole    cubbyhole_ad070930    per-token private secret storage
    identity/     identity     identity_5397dc2f     identity store
    secret/       kv           kv_2362c227           key/value secret storage
    sys/          system       system_410e4276       system endpoints used for control, policy and debugging

.. note:: Line 5: EdgeX Foundry platform is using the Key/Value secret storage named ``secret``

Let's drill down into the ``secret`` k/v storage and walk through a predefined hierarchical tree structure (path).

.. note:: the ``pkisetup`` tool used during the Vault Docker image build process generates all the related X.509 TLS materials. The ``vault-worker`` service is storing each service materials into Vault using arbitrary paths, setting up access policies accordingly.

For example, the API Gateway (Kong) service X.509 TLS materials:

.. code-block:: shell
    :linenos:
    :emphasize-lines: 4,9,14,19

    root@edgex-vault:/vault # vault list secret
    Keys
    ----
    edgex/

    root@edgex-vault:/vault # vault list secret/edgex
    Keys
    ----
    pki/

    root@edgex-vault:/vault # vault list secret/edgex/pki
    Keys
    ----
    tls/

    root@edgex-vault:/vault # vault list secret/edgex/pki/tls
    Keys
    ----
    edgex-kong

Displaying the API gateway (Kong) service X.509 TLS materials (TLS certificate ``cert`` & corresponding private key ``key``):

.. code-block:: shell
    :linenos:
    :emphasize-lines: 5,22

    root@edgex-vault:/vault # vault read secret/edgex/pki/tls/edgex-kong
    Key                 Value
    ---                 -----
    refresh_interval    168h
    cert                -----BEGIN CERTIFICATE-----
    MIICrjCCAjWgAwIBAgIQDvZxhmU3nyG4cwXlQesMFDAKBggqhkjOPQQDAzB7MQsw
    CQYDVQQGEwJVUzELMAkGA1UECBMCQ0ExFjAUBgNVBAcTDVNhbiBGcmFuY2lzY28x
    FzAVBgNVBAoTDkVkZ2VYRm91bmRyeUNBMRUwEwYDVQQLEwxFZGdlWEZvdW5kcnkxt
    FzAVBgNVBAMTDkVkZ2VYRm91bmRyeUNBMB4XDTE4MTIwNTE0MDUyOFoXDTI4MTIw
    NTE0MDUyOFowazELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAkNBMRYwFAYDVQQHEw1T
    YW4gRnJhbmNpc2NvMRMwEQYDVQQKEwplZGdleC1rb25nMQ0wCwYDVQQLEwRLb25n
    MRMwEQYDVQQDEwplZGdleC1rb25nMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAE2dnb
    EboXET1TjzmWKFv3A0wklwNbs9t9JLT0ecpQr64a277UnTAQhgCv2e2/x9EP4eta
    gSlz5PCqdAykWW0URIEPSwUKWmx4x1DBwyUD2oDOPsFrywIVEC3DlqQAL6huo4GN
    MIGKMA4GA1UdDwEB/wQEAwIFoDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMB
    Af8EAjAAMB8GA1UdIwQYMBaAFFX63XbmPNpLceOJYyt2Y+LfW/gxMDQGA1UdEQQt
    MCuCCmVkZ2V4LWtvbmeCEGVkZ2V4LWtvbmcubG9jYWyBC2FkbWluQGxvY2FsMAoG
    CCqGSM49BAMDA2cAMGQCMCaH3sSKq6nlr6hBJx82wYEiK4slMbySiQZg5mLcwrsQ
    tIPGcQ2lgBdQYzI3ymOS5gIwNhpQmo/p3hkoFzA4rxIAZx/GUgZan51JlXW0rpgz
    4HerRLe55EmvF10mF7VCGOXe
    -----END CERTIFICATE-----
    key                 -----BEGIN PRIVATE KEY-----
    MIG2AgEAMBAGByqGSM49AgEGBSuBBAAiBIGeMIGbAgEBBDC6BRUXqkJbey765+8b
    Oib2qG/jbai2rzp0+NQyJv4ijAyYjJlxhVGggZqPPBy8baqhZANiAATZ2dsRuhcR
    PVOPOZYoW/cDTCSXA1uz230ktPR5ylCvrhrbvtSdMBCGAK/Z7b/H0Q/h61qBKXPk
    8Kp0DKRZbRREgQ9LBQpabHjHUMHDJQPagM4+wWvLAhUQLcOWpAAvqG4=
    -----END PRIVATE KEY-----

.. note:: These two key values are in PEM format.

2nd alternative: using the Vault Web UI.
----------------------------------------

Open a browser session on ``https://<EdgeX Vault Server>:8200``, accept the self-signed TLS server certificate and sign-in with the Root Token (see above **1st alternative** to learn how to fetch this token):

.. image:: EdgeX-UI-Vault-01.png
    :width: 606px
    :align: center
    :height: 504px

Upper left corner of the current Vault UI session, the sign-out menu displaying the current token name:

.. image:: EdgeX-UI-Vault-02.png
    :width: 275px
    :align: center
    :height: 156px

Select the Vault secret backend:

.. image:: EdgeX-UI-Vault-03.png
    :align: center

Navigate the API Gateway (Kong) service X.509 TLS materials path (edgex/pki/tls/edgex-kong):

.. image:: EdgeX-UI-Vault-04.png
    :align: center

The Vault UI also allows entering Vault CLI commands (see above **1st alternative**) using an embedded console:

.. image:: EdgeX-UI-Vault-05.png
    :align: center


3rd alternative: directly using the Vault HTTP API with cURL commands.
----------------------------------------------------------------------

See paragraph **Configuring the Secret Store** to have more details on the ``--cacert`` option (identical purpose as the ``VAULT_CAPATH`` environment variable for Vault CLI).

See paragraph **Using the Secret Store** to have more details on gathering the Vault Root Token (ref: ``/vault/config/assets/resp-init.json``).

See HashiCorp Vault API documentation for further details on syntax and usage (https://www.vaultproject.io/api/).

Displaying (GET) the API gateway (Kong) service X.509 TLS materials (TLS certificate cert & corresponding private key key):

.. code-block:: shell
    :linenos:
    :emphasize-lines: 1,3,5

    curl -s --cacert /vault/config/pki/EdgeXFoundryCA/EdgeXFoundryCA.pem \
        --location \
        --header "X-Vault-Token: 01dbbae4-353a-8cdf-8189-4d50e5535a6f" \
        --request GET \
        https://edgex-vault:8200/v1/secret/edgex/pki/tls/edgex-kong | jq

.. note:: Line 2: the ``--location`` option allows following a redirection (necessary when using a Vault cluster)

          Line 5: the Vault API path prefix ``/v1/secret`` and the API Gateway X.509 TLS materials k/v ``/edgex/pki/tls/edgex-kong``.

          Line 5: the ``jq`` tool is a lightweight and flexible command-line JSON processor (https://stedolan.github.io/jq/) allowing JSON pretty printing in the terminal.

Sample JSON returned:

.. code-block:: json
    :linenos:
    :emphasize-lines: 7,8

    {
    "request_id": "eaa80a1b-0d31-8d11-6ce1-8d9aa3ac6a19",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 604800,
    "data": {
        "cert": "-----BEGIN CERTIFICATE-----\nMIICrjCCAjWgAwIBAgIQDvZxhmU3nyG4cwXlQesMFDAKBggqhkjOPQQDAzB7MQsw\nCQYDVQQGEwJVUzELMAkGA1UECBMCQ0ExFjAUBgNVBAcTDVNhbiBGcmFuY2lzY28x\nFzAVBgNVBAoTDkVkZ2VYRm91bmRyeUNBMRUwEwYDVQQLEwxFZGdlWEZvdW5kcnkx\nFzAVBgNVBAMTDkVkZ2VYRm91bmRyeUNBMB4XDTE4MTIwNTE0MDUyOFoXDTI4MTIw\nNTE0MDUyOFowazELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAkNBMRYwFAYDVQQHEw1T\nYW4gRnJhbmNpc2NvMRMwEQYDVQQKEwplZGdleC1rb25nMQ0wCwYDVQQLEwRLb25n\nMRMwEQYDVQQDEwplZGdleC1rb25nMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAE2dnb\nEboXET1TjzmWKFv3A0wklwNbs9t9JLT0ecpQr64a277UnTAQhgCv2e2/x9EP4eta\ngSlz5PCqdAykWW0URIEPSwUKWmx4x1DBwyUD2oDOPsFrywIVEC3DlqQAL6huo4GN\nMIGKMA4GA1UdDwEB/wQEAwIFoDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMB\nAf8EAjAAMB8GA1UdIwQYMBaAFFX63XbmPNpLceOJYyt2Y+LfW/gxMDQGA1UdEQQt\nMCuCCmVkZ2V4LWtvbmeCEGVkZ2V4LWtvbmcubG9jYWyBC2FkbWluQGxvY2FsMAoG\nCCqGSM49BAMDA2cAMGQCMCaH3sSKq6nlr6hBJx82wYEiK4slMbySiQZg5mLcwrsQ\ntIPGcQ2lgBdQYzI3ymOS5gIwNhpQmo/p3hkoFzA4rxIAZx/GUgZan51JlXW0rpgz\n4HerRLe55EmvF10mF7VCGOXe\n-----END CERTIFICATE-----\n",
        "key": "-----BEGIN PRIVATE KEY-----\nMIG2AgEAMBAGByqGSM49AgEGBSuBBAAiBIGeMIGbAgEBBDC6BRUXqkJbey765+8b\nOib2qG/jbai2rzp0+NQyJv4ijAyYjJlxhVGggZqPPBy8baqhZANiAATZ2dsRuhcR\nPVOPOZYoW/cDTCSXA1uz230ktPR5ylCvrhrbvtSdMBCGAK/Z7b/H0Q/h61qBKXPk\n8Kp0DKRZbRREgQ9LBQpabHjHUMHDJQPagM4+wWvLAhUQLcOWpAAvqG4=\n-----END PRIVATE KEY-----\n"
    },
    "wrap_info": null,
    "warnings": null,
    "auth": null
    }

.. note:: Lines 7 & 8: the two key values (TLS certificate ``cert`` & corresponding private key ``key``) are in PEM format (https://tools.ietf.org/html/rfc1421).

Displaying (LIST) the root key path in the Vault ``secret`` backend for the EdgeX Foudry platform (``edgex``):

.. code-block:: shell
    :linenos:
    :emphasize-lines: 1,3,5

    curl -s --cacert /vault/config/pki/EdgeXFoundryCA/EdgeXFoundryCA.pem \
        --location \
        --header "X-Vault-Token: 01dbbae4-353a-8cdf-8189-4d50e5535a6f" \
        --request LIST \
        https://edgex-vault:8200/v1/secret | jq

Sample JSON returned:

.. code-block:: json
    :linenos:
    :emphasize-lines: 7,8

    {
    "request_id": "0e0ea024-176d-21b3-73cb-99f17729b230",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 0,
    "data": {
        "keys": [
        "edgex/"
        ]
    },
    "wrap_info": null,
    "warnings": null,
    "auth": null
    }

Displaying (GET) the Vault seal status (``API path: /v1/sys/seal-status``): 

.. code-block:: shell
    :linenos:
    :emphasize-lines: 1,3,5

    curl -s --cacert /vault/config/pki/EdgeXFoundryCA/EdgeXFoundryCA.pem \
        --location \
        --header "X-Vault-Token: 01dbbae4-353a-8cdf-8189-4d50e5535a6f" \
        --request GET \
        https://edgex-vault:8200/v1/sys/seal-status | jq

Sample JSON returned:

.. code-block:: json
    :linenos:
    :emphasize-lines: 3,4,5

    {
    "type": "shamir",
    "sealed": false,
    "t": 1,
    "n": 1,
    "progress": 0,
    "nonce": "",
    "version": "0.10.2",
    "cluster_name": "vault-cluster-57b3c4ed",
    "cluster_id": "fe6d18bf-fa9c-0d52-3278-bca0390af023"
    }

.. note:: Line 3: Vault is **unsealed** therefore available and ready for requests.

          Line 4 & 5: Vault Shamir Secret Sharing default configuration for EdgeX Foundry: 1 share with threshold 1 (no sharding). 