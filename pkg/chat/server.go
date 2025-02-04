package chat

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	pb "khchat/api/chat/pb"
)

type ChatServer struct {
	pb.UnimplementedChatServiceServer
	Db      *pgxpool.Pool
	Redis   *redis.Client
	mu      sync.Mutex
	Streams map[int64]pb.ChatService_ReceiveMessagesServer // 存储用户的 gRPC Stream
}

// **1. 发送消息（判断在线状态，直接推送或存入 Redis）**
func (s *ChatServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	_, err := s.Db.Exec(ctx, "INSERT INTO messages (sender_id, receiver_id, content, created_at) VALUES ($1, $2, $3, $4)",
		req.SenderId, req.ReceiverId, req.Message, time.Now())
	if err != nil {
		return nil, err
	}

	// **检查用户是否在线**
	onlineStatus, err := s.CheckOnlineStatus(ctx, &pb.CheckOnlineStatusRequest{UserId: req.ReceiverId})
	if err != nil {
		log.Println("检查在线状态失败:", err)
	}
	if onlineStatus.IsOnline {
		// **如果在线，直接推送消息**
		go s.pushMessageToClient(req.ReceiverId, req)
		return &pb.SendMessageResponse{Success: true}, nil
	} else {
		// **用户不在线，存入 Redis，包含 SenderId**
		key := fmt.Sprintf("offline:%d", req.ReceiverId)
		offlineMsg := fmt.Sprintf("%d:%s", req.SenderId, req.Message) // 存储格式 "SenderId:Message"
		s.Redis.RPush(ctx, key, offlineMsg)
		log.Printf("用户 %d 不在线，消息存入 Redis", req.ReceiverId)

		return &pb.SendMessageResponse{Success: true}, nil
	}
}

// **2. 监听 gRPC 消息流**
func (s *ChatServer) ReceiveMessages(req *pb.ListenMessagesRequest, stream pb.ChatService_ReceiveMessagesServer) error {
	ctx := stream.Context() // 确保使用流的上下文
	userID := req.UserId

	// **记录用户的消息流**
	s.mu.Lock()
	s.Streams[userID] = stream
	s.mu.Unlock()

	log.Printf("用户 %d 连接 gRPC 监听消息流", userID)

	// **使用 defer 确保删除 Stream**
	defer func() {
		s.mu.Lock()
		delete(s.Streams, userID)
		s.mu.Unlock()
		log.Printf("用户 %d 断开 gRPC 连接，已删除其 Stream", userID)
	}()

	// **获取 Redis 中的离线消息**
	key := fmt.Sprintf("offline:%d", userID)
	messages, err := s.Redis.LRange(ctx, key, 0, -1).Result()
	if err == nil && len(messages) > 0 {
		for _, msg := range messages {
			parts := strings.SplitN(msg, ":", 2)
			if len(parts) == 2 {
				senderID, _ := strconv.ParseInt(parts[0], 10, 64)
				message := parts[1]

				err := stream.Send(&pb.Message{
					SenderId:   senderID,
					ReceiverId: userID,
					Message:    message,
					Timestamp:  time.Now().Format(time.DateTime),
				})
				if err != nil {
					log.Printf("用户 %d WebSocket 断开，无法发送离线消息", userID)
					return err
				}
			}
		}
		s.Redis.Del(ctx, key) // **删除 Redis 里的离线消息**
		log.Printf("用户 %d 收到了 %d 条离线消息", userID, len(messages))
	}

	// **监听 gRPC 连接状态**
	<-ctx.Done() // 阻塞等待用户断开
	log.Printf("检测到用户 %d 断开 gRPC 连接", userID)
	return nil
}

// **3. 通过 gRPC Stream 直接推送消息**
func (s *ChatServer) pushMessageToClient(receiverID int64, msg *pb.SendMessageRequest) {
	s.mu.Lock()
	stream, exists := s.Streams[receiverID]
	s.mu.Unlock()

	if exists {
		err := stream.Send(&pb.Message{
			SenderId:   msg.SenderId,
			ReceiverId: msg.ReceiverId,
			Message:    msg.Message,
			Timestamp:  time.Now().Format(time.DateTime),
		})
		if err != nil {
			log.Println("gRPC 消息推送失败:", err)
		} else {
			log.Printf("消息已实时推送给用户 %d: %s", receiverID, msg.Message)
		}
	} else {
		log.Printf("用户 %d 不在线，无法推送", receiverID)
	}
}

// **4. 查询用户是否在线**
func (s *ChatServer) CheckOnlineStatus(ctx context.Context, req *pb.CheckOnlineStatusRequest) (*pb.CheckOnlineStatusResponse, error) {
	// **检查用户是否有活跃的 gRPC 连接**
	s.mu.Lock()
	_, isOnline := s.Streams[req.UserId]
	s.mu.Unlock()

	return &pb.CheckOnlineStatusResponse{IsOnline: isOnline}, nil
}
