# Notes for developers regarding to use different Redis Access Control List (ACL)

Currently, the `security-bootstrapper` configureRedis produces the ACL configuration file for Redis' default user.
Should using different ACL rules call for a debugging needs, developers could override this built-in configuration behavior as follows:

 1. Currently, the default ACL file path inside the redis.conf is pointing to the path with the file name `edgex_redis_acl.conf`.  A developer can always provide his own redis config file containing the different file name (eg. developer-acl.conf) for ACL rules like adding some `dangerous` commands such as `INFO, MONITOR, BGSAVE, and FLUSHD` inside his own ACL file using `+` directive. eg.:

```text
    user default on allkeys +@all -@dangerous #_{{.HashedRedisPwd}}_ +INFO +MONITOR +BGSAVE + FLUSHDB
```

  and use his own config file on developer modified redis' entrypoint script to start the redis server like:

```sh
      exec /usr/local/bin/docker-entrypoint.sh redis-server developer_redis.conf
```

  on `database` service of a docker-compose file.

  Note that the HashedRedisPwd still needs to be come from the original dynamically created redis.conf file as it is read from secretstore Vault.

  A developer can also just modified the ACL file `edgex_redis_acl.conf` directly and then use `ACL LOAD` or `ACL SAVE` commands to change ACL rules assuming he/she has the right permissions to update that file.

 2. For snap, a developer can just change `CONFIG_FILE` environment variable of snap `redis` service to point to his own above-mentioned configuration file, `developer_redis.conf` (assuming developer is putting his configuration file under the same directory eg. `$SNAP_DATA/redis/conf`; creating a new mounted file system and directory inside snapcraft is beyond the scope of this topic).
