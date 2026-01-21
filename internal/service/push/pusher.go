package push

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"kama_chat_server/internal/config"
	myKafka "kama_chat_server/internal/service/kafka"
	pb "kama_chat_server/pb"
	"kama_chat_server/pkg/constants"
	"kama_chat_server/pkg/zlog"
	"log"
	"net"
	"net/http"
)

type pusher struct {
	pb.UnimplementedPushServer
	messageChan chan *MessagePush
	Clients     map[int64]*Client
	Login       chan *Client // 登录通道
	Logout      chan *Client // 退出登录通道
}

type MessagePush struct {
	UserId  int64  `json:"user_id"`
	Message []byte `json:"message"`
}

var Pusher *pusher

func init() {
	if Pusher == nil {
		Pusher = &pusher{
			messageChan: make(chan *MessagePush),
			Clients:     make(map[int64]*Client),
			Login:       make(chan *Client),
			Logout:      make(chan *Client),
		}

	}
}

func (p *pusher) Start() {
	zlog.Info("Pusher开始启动")
	listen, err := net.Listen("tcp", ":9090")
	if err != nil {
		zlog.Error(err.Error())
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPushServer(grpcServer, &pusher{})

	go func() {
		err = grpcServer.Serve(listen)
		if err != nil {
			zlog.Error(err.Error())
		}
	}()
	zlog.Info("Pusher开始服务")
	for {
		select {
		case message := <-p.messageChan:
			{
				userId, msg := message.UserId, message.Message
				p.Clients[userId].SendBack <- msg
			}
		case client := <-p.Login:
			{
				p.Clients[client.UserId] = client
				zlog.Info(fmt.Sprintf("欢迎来到kama聊天服务器，亲爱的用户%s\n", client.UserId))
				err := client.Conn.WriteMessage(websocket.TextMessage, []byte("欢迎来到kama聊天服务器"))
				if err != nil {
					log.Fatal(err.Error())
				}
			}
		case client := <-p.Logout:
			{
				delete(p.Clients, client.UserId)
				zlog.Info(fmt.Sprintf("用户%s退出登录\n", client.UserId))
				if err := client.Conn.WriteMessage(websocket.TextMessage, []byte("已退出登录")); err != nil {
					log.Fatal(err.Error())
				}
			}

		}

	}
}

func (p *pusher) Push(ctx context.Context, req *pb.PushRequest) (*pb.ResponseForPush, error) {
	p.messageChan <- &MessagePush{
		UserId:  req.UserId,
		Message: req.Message,
	}
	return &pb.ResponseForPush{}, nil
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	// 检查连接的Origin头
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var ctx = context.Background()

var messageMode = config.GetConfig().KafkaConfig.MessageMode

// 读取websocket消息并发送给send通道
func (c *Client) Read() {
	zlog.Info("ws read goroutine start")
	for {
		// 阻塞有一定隐患，因为下面要处理缓冲的逻辑，但是可以先不做优化，问题不大
		_, jsonMessage, err := c.Conn.ReadMessage() // 阻塞状态
		if err != nil {
			zlog.Error(err.Error())
			return // 直接断开websocket
		} else {
			// 方法1：大端序（网络字节序）
			bufBig := make([]byte, 8)
			binary.BigEndian.PutUint64(bufBig, uint64(c.UserId))
			if err := myKafka.KafkaService.ConversationWriter.WriteMessages(ctx, kafka.Message{
				Key:   bufBig,
				Value: jsonMessage,
			}); err != nil {
				zlog.Error(err.Error())
			}
			zlog.Info("已发送消息：" + string(jsonMessage))
		}
	}
}

// 从send通道读取消息发送给websocket
func (c *Client) Write() {
	zlog.Info("ws write goroutine start")
	for messageBack := range c.SendBack { // 阻塞状态
		// 通过 WebSocket 发送消息
		err := c.Conn.WriteMessage(websocket.TextMessage, messageBack)
		if err != nil {
			zlog.Error(err.Error())
			return // 直接断开websocket
		}
		// log.Println("已发送消息：", messageBack.Message)
		// 说明顺利发送，修改状态为已发送
		//if res := dao.GormDB.Model(&model.Message{}).Where("uuid = ?", messageBack.Uuid).Update("status", message_status_enum.Sent); res.Error != nil {
		//	zlog.Error(res.Error.Error())
		//}
	}
}

type Client struct {
	UserId   int64
	Conn     *websocket.Conn
	SendBack chan []byte
}

// NewClientInit 当接受到前端有登录消息时，会调用该函数
func NewClientInit(c *gin.Context, userId int64) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Error(err.Error())
	}
	client := &Client{
		UserId:   userId,
		Conn:     conn,
		SendBack: make(chan []byte, constants.CHANNEL_SIZE),
	}
	zlog.Info("开始登陆client")
	Pusher.Login <- client
	zlog.Info("准备启动read&write")
	go client.Read()
	go client.Write()
	zlog.Info("ws连接成功")
}

// ClientLogout 当接受到前端有登出消息时，会调用该函数
func ClientLogout(userId int64) (string, int) {

	client := Pusher.Clients[userId]
	if client != nil {
		Pusher.Logout <- client
		if err := client.Conn.Close(); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		//close(client.SendTo)
		close(client.SendBack)
	}
	return "退出成功", 0
}

func (p *pusher) Close() {
	close(p.Login)
	close(p.Logout)
	close(p.messageChan)
}
