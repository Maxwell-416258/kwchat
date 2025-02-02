package cfg

import (
	"github.com/spf13/viper"
	"khchat/pkg/global"
	"log"
)

// 初始化数据库、app配置
func InitConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("pkg/cfg/")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	//获取配置
	global.AppConf.Port = viper.GetInt("app.port")
	global.PostgresConf.Host = viper.GetString("postgres.host")
	global.PostgresConf.Port = viper.GetInt("postgres.port")
	global.PostgresConf.User = viper.GetString("postgres.user")
	global.PostgresConf.Password = viper.GetString("postgres.password")
	global.PostgresConf.DbName = viper.GetString("postgres.dbname")
	global.RedisConf.Host = viper.GetString("redis.host")
	global.RedisConf.Port = viper.GetInt("redis.port")

}
