package redis

// Redis commmands used in this project
// Reference: https://redis.io/commands
const (
	MULTI     = "MULTI"
	SET       = "SET"
	GET       = "GET"
	EXISTS    = "EXISTS"
	DEL       = "DEL"
	HSET      = "HSET"
	HGET      = "HGET"
	HEXISTS   = "HEXISTS"
	HDEL      = "HDEL"
	SADD      = "SADD"
	SREM      = "SREM"
	ZADD      = "ZADD"
	ZREM      = "ZREM"
	EXEC      = "EXEC"
	ZRANGE    = "ZRANGE"
	ZREVRANGE = "ZREVRANGE"
	MGET      = "MGET"
	ZCARD     = "ZCARD"
)
