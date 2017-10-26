package redisex

import (
	"github.com/go-redis/redis"
)

func NewEvalShaCmd(sha1 string, keys []string, args ...interface{}) *redis.Cmd {
	cmdArgs := make([]interface{}, 3+len(keys)+len(args))
	cmdArgs[0] = "evalsha"
	cmdArgs[1] = sha1
	cmdArgs[2] = len(keys)
	for i, key := range keys {
		cmdArgs[3+i] = key
	}
	pos := 3 + len(keys)
	for i, arg := range args {
		cmdArgs[pos+i] = arg
	}
	return redis.NewCmd(cmdArgs...)
}
