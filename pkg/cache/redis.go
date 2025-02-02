// pkg/cache/redis.go
package cache

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"khchat/pkg/global"
)

func InitRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", global.RedisConf.Host, global.RedisConf.Port),
	})
}
