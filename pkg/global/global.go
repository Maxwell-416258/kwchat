package global

type AppConfig struct {
	Port int
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DbName   string
}

type RedisConfig struct {
	Host string
	Port int
}

var AppConf AppConfig
var PostgresConf PostgresConfig
var RedisConf RedisConfig
