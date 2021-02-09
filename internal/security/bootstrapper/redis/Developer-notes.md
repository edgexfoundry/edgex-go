# Notes for developers regarding to use different Redis Access Control List (ACL)

Currently, the `security-bootstrapper` configureRedis produces the ACL configuration file for Redis' default user.
Should using different ACL rules call for a debugging needs, developers could override this built-in configuration behavior as follows:

 1. A developer can provide his own config file containing the ACL rules for add some `dangerous` commands like `INFO, MONITOR, BGSAVE, and FLUSHD` inside his own config file using `+` directive. eg.:

```text
    user default on allkeys +@all -@dangerous >_{{.RedisPwd}}_ +INFO +MONITOR +BGSAVE + FLUSHDB
```

  and use his own config file on developer modified redis' entrypoint script to start the redis server like:

```sh
      exec /usr/local/bin/docker-entrypoint.sh redis-server developer_redis.conf
```

  on `database` service of a docker-compose file.

  Note that the RedisPwd still needs to be come from the original dynamically created redis.conf file as it is read from secretstore Vault.

 2. For snap, a developer can just change `CONFIG_FILE` environment variable of snap `redis` service to point to his own above-mentioned configuration file, `developer_redis.conf` (assuming developer is putting his configuration file under the same directory eg. `$SNAP_DATA/redis/conf`; creating a new mounted file system and directory inside snapcraft is beyond the scope of this topic).
