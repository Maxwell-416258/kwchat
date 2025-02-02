// pkg/db/postgres.go
package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"khchat/pkg/global"
	"log"
)

func InitPostgres() (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		global.PostgresConf.User, global.PostgresConf.Password, global.PostgresConf.Host, global.PostgresConf.Port, global.PostgresConf.DbName))
	if err != nil {
		log.Fatalf("无法连接 PostgreSQL: %v", err)
		return nil, err
	}
	return dbpool, nil
}
