/*******************************************************************************
 * Copyright 2018 Redis Labs Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
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
	scriptGetObjectsByRange = `
	local magic = 4096
	local ids = redis.call('ZRANGE', KEYS[1], ARGV[1], ARGV[2])
	local rep = {}
	if #ids > 0 then
		for i = 1, #ids, magic do
			local temp = redis.call('MGET', unpack(ids, i, i+magic < #ids and i+magic or #ids))
			for _, o in ipairs(temp) do
				table.insert(rep, o)
			end
		end
		return rep
	else
		return nil
	end
	`
	scriptGetObjectsByRangeFilter = `
	local magic = 4096
	local ids = redis.call('ZRANGE', KEYS[1], ARGV[1], ARGV[2])
	local rep = {}
	if #ids > 0 then
		for i, id in ipairs(ids) do
			local v = redis.call('ZSCORE', KEYS[2], id)
			if v == nil then
				ids[i] = nil
			end
		end
		for i = 1, #ids, magic do
			local temp = redis.call('MGET', unpack(ids, i, i+magic < #ids and i+magic or #ids))
			for _, o in ipairs(temp) do
				table.insert(rep, o)
			end
		end
	else
		return nil
	end
	return rep
	`
	scriptGetObjectsByScore = `
	local magic = 4096
	local cmd = {
		'ZRANGEBYSCORE', KEYS[1], ARGV[1],
		tonumber(ARGV[2]) < 0 and '+inf' or ARGV[2],
	}
	if tonumber(ARGV[3]) ~= 0 then
		table.insert(cmd, 'LIMIT')
		table.insert(cmd, 0)
		table.insert(cmd, ARGV[3])
	end
	local ids = redis.call(unpack(cmd))
	local rep = {}
	if #ids > 0 then
		for i = 1, #ids, magic do
			local temp = redis.call('MGET', unpack(ids, i, i+magic < #ids and i+magic or #ids))
			for _, o in ipairs(temp) do
				table.insert(rep, o)
			end
		end
	else
		return nil
	end
	return rep
	`
	scriptUnlinkZsetMembers = `
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

var scripts = map[string]redis.Script{
	"getObjectsByRange":       *redis.NewScript(1, scriptGetObjectsByRange),
	"getObjectsByRangeFilter": *redis.NewScript(2, scriptGetObjectsByRangeFilter),
	"getObjectsByScore":       *redis.NewScript(1, scriptGetObjectsByScore),
	"unlinkZsetMembers":       *redis.NewScript(1, scriptUnlinkZsetMembers),
	"unlinkCollection":        *redis.NewScript(0, scriptUnlinkCollection),
}

func getObjectsByRangeLua(conn redis.Conn, key string, start, end int) (objects [][]byte, err error) {
	s := scripts["getObjectsByRange"]
	objects, err = redis.ByteSlices(s.Do(conn, key, start, end))
	if err != nil {
		return nil, err
	}

	return objects, nil
}

func getObjectsByRangeFilterLua(conn redis.Conn, key string, filter string, start, end int) (objects [][]byte, err error) {
	s := scripts["getObjectsByRangeFilter"]
	objects, err = redis.ByteSlices(s.Do(conn, key, filter, start, end))
	if err != nil {
		return nil, err
	}

	return objects, nil
}

// Return objects by a score from a zset
// if limit is 0, all are returned
// if end is negative, it is considered as positive infinity
func getObjectsByScoreLua(conn redis.Conn, key string, start, end int64, limit int) (objects [][]byte, err error) {
	s := scripts["getObjectsByScore"]
	objects, err = redis.ByteSlices(s.Do(conn, key, start, end, limit))
	if err != nil {
		return nil, err
	}

	return objects, nil
}
