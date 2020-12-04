package redis

import "strings"

// Redis commmands used in this project
// Reference: https://redis.io/commands
const (
	MULTI            = "MULTI"
	SET              = "SET"
	GET              = "GET"
	EXISTS           = "EXISTS"
	DEL              = "DEL"
	HSET             = "HSET"
	HGET             = "HGET"
	HEXISTS          = "HEXISTS"
	HDEL             = "HDEL"
	SADD             = "SADD"
	SREM             = "SREM"
	ZADD             = "ZADD"
	ZREM             = "ZREM"
	EXEC             = "EXEC"
	ZRANGE           = "ZRANGE"
	ZREVRANGE        = "ZREVRANGE"
	MGET             = "MGET"
	ZCARD            = "ZCARD"
	ZCOUNT           = "ZCOUNT"
	UNLINK           = "UNLINK"
	ZRANGEBYSCORE    = "ZRANGEBYSCORE"
	ZREVRANGEBYSCORE = "ZREVRANGEBYSCORE"
	LIMIT            = "LIMIT"
)

const (
	InfiniteMin     = "-inf"
	InfiniteMax     = "+inf"
	GreaterThanZero = "(0"
	DBKeySeparator  = ":"
)

// CreateKey creates Redis key by connecting the target key with DBKeySeparator
func CreateKey(targets ...string) string {
	return strings.Join(targets, DBKeySeparator)
}
