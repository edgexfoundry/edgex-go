package redis

// Redis commmands used in this project
// Reference: https://redis.io/commands
const (
	MULTI     = "MULTI"
	SET       = "SET"
	EXISTS    = "EXISTS"
	HSET      = "HSET"
	HEXISTS   = "HEXISTS"
	SADD      = "SADD"
	ZADD      = "ZADD"
	GET       = "GET"
	EXEC      = "EXEC"
	ZRANGE    = "ZRANGE"
	ZREVRANGE = "ZREVRANGE"
	MGET      = "MGET"
)
