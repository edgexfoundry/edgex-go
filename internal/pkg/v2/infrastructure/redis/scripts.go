package redis

import "github.com/gomodule/redigo/redis"

/***********************
 * Lua scripts notes:
 * magic number 4096 is less < 8000 (redis:/deps/lua/lapi.c:LUAI_MAXCSTACK -> unpack error)
 * assumes a single instance
 * `get*` scripts are implementations for range operations. Can be used when the server is
 * remote in order to reduce latency.
 */
const (
	scriptUnlinkZSETMembers = `
	local magic = 4096
	local ids = redis.call('ZRANGE', KEYS[1], 0, -1)
	if #ids > 0 then
		for i = 1, #ids, magic do
			redis.call('UNLINK', unpack(ids, i, i+magic < #ids and i+magic or #ids))
		end
	end
	`
	scriptUnlinkCollection = `
	local magic = 4096
	redis.replicate_commands()
	local c = 0
	repeat
		local s = redis.call('SCAN', c, 'MATCH', ARGV[1] .. '*')
		c = tonumber(s[1])
		if #s[2] > 0 then
			redis.call('UNLINK', unpack(s[2]))
		end
	until c == 0
	`
)

const (
	unlinkZSETMembers = "unlinkZSETMembers"
	unlinkCollection  = "unlinkCollection"
)

var redisScripts = map[string]redis.Script{
	unlinkZSETMembers: *redis.NewScript(1, scriptUnlinkZSETMembers),
	unlinkCollection:  *redis.NewScript(0, scriptUnlinkCollection),
}
