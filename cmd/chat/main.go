package main

import (
	"fmt"
	"google.golang.org/grpc"
	pb "khchat/api/chat/pb"
	"khchat/pkg/cache"
	"khchat/pkg/cfg"
	"khchat/pkg/chat"
	"khchat/pkg/db"
	"khchat/pkg/global"
	"log"
	"net"
)

func main() {
	// 初始化配置
	cfg.InitConfig()

	// 初始化数据库连接
	dbpool, err := db.InitPostgres()
	if err != nil {
		log.Fatal("无法连接 PostgreSQL:", err)
	}
	defer dbpool.Close()

	// 初始化 Redis 客户端
	redisClient := cache.InitRedis()

	// 创建并启动 gRPC 服务
	streams := make(map[int64]pb.ChatService_ReceiveMessagesServer)
	server := &chat.ChatServer{Streams: streams, Db: dbpool, Redis: redisClient}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", global.AppConf.Port))
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChatServiceServer(grpcServer, server)

	log.Println(fmt.Sprintf("gRPC 服务器启动，监听端口 :%d", global.AppConf.Port))
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
