package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"kama_chat_server/internal/dao"
	"kama_chat_server/internal/model"
	myKafka "kama_chat_server/internal/service/kafka"
	pb "kama_chat_server/pb"
	"kama_chat_server/pkg/zlog"
	"time"
)

type userServer struct {
	//Conn    *websocket.Conn
	userSeq    map[int64]int64
	grpcClient pb.PushClient
}

var UserServer *userServer

func init() {
	if UserServer == nil {
		conn, err := grpc.NewClient("101.43.155.144:9090", grpc.WithInsecure())
		if err != nil {
			zlog.Error(err.Error())
		}
		client := pb.NewPushClient(conn)
		UserServer = &userServer{
			userSeq:    make(map[int64]int64),
			grpcClient: client,
		}
	}
}

type UserMsg struct {
	UserId         int64     `json:"user_id"`
	MsgId          int64     `json:"msg_id"`
	ConversationId string    `json:"conversation_id"`
	Seq            int64     `json:"seq"`
	SendId         int64     `json:"send_id"`
	Type           int8      `json:"type"`
	Content        string    `json:"content"`
	Status         int8      `json:"status"`
	SendTime       time.Time `json:"send_time"`
}

func (u *userServer) Start() {
	zlog.Info("UserServer开始启动")
	defer func() {
		zlog.Warn("进入defer函数")
		if r := recover(); r != nil {
			zlog.Error(fmt.Sprintf("userServer panic recovered: %v", r))
			// 打印堆栈跟踪
		}
	}()
	for {
		kafkaMessage, err := myKafka.KafkaService.UserReader.ReadMessage(context.Background())
		if err != nil {
			zlog.Error(err.Error())
		}

		zlog.Info(fmt.Sprintf("topic=%s, partition=%d, offset=%d, key=%s, value=%s", kafkaMessage.Topic, kafkaMessage.Partition, kafkaMessage.Offset, kafkaMessage.Key, kafkaMessage.Value))
		data := kafkaMessage.Value
		var userMsg UserMsg
		if err := json.Unmarshal(data, &userMsg); err != nil {
			zlog.Error(err.Error())
		}
		zlog.Info("user_server原消息为：" + string(data))
		//log.Println("user_server原消息为：", data, "反序列化后为：", userMsg)
		userMsg.Seq = u.userSeq[userMsg.UserId]
		userMsgList := &model.UserMsgList{
			UserId:         userMsg.UserId,
			MsgId:          userMsg.MsgId,
			ConversationId: userMsg.ConversationId,
			Seq:            userMsg.Seq,
		}
		//
		if res := dao.GormDB.Create(&userMsgList); res.Error != nil {
			zlog.Error(res.Error.Error())
		}
		zlog.Info("已存入:" + string(data))
		//
		message, err := json.Marshal(userMsg)
		if err != nil {
			zlog.Error(err.Error())
		}
		//
		zlog.Info("已序列化:" + string(data))
		_, err = u.grpcClient.Push(context.Background(), &pb.PushRequest{
			UserId:  userMsg.UserId,
			Message: message,
		})
		if err != nil {
			zlog.Error(err.Error())
		}
		zlog.Info("已push:" + string(message))
	}
}

func (u *userServer) getSeq(userId int64) int64 {
	if u.userSeq[userId] == 0 {
		var seq int64
		// 只使用Pluck，更简洁
		_ = dao.GormDB.Model(&model.UserMsgList{}).
			Where("user_id = ?", userId).
			Order("seq DESC").
			Limit(1).
			Pluck("seq", &seq).Error
		u.userSeq[userId] = seq
	}
	u.userSeq[userId]++
	return u.userSeq[userId]
}
