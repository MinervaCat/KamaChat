package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"kama_chat_server/internal/dao"
	"kama_chat_server/internal/dto/request"
	"kama_chat_server/internal/model"
	myKafka "kama_chat_server/internal/service/kafka"
	"kama_chat_server/pkg/enum/message/message_status_enum"
	"kama_chat_server/pkg/util/random"
	"kama_chat_server/pkg/zlog"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"time"
)

type conversationServer struct {
	conversationSeq map[string]int64
}

var ConversationServer *conversationServer

var conversationQuit = make(chan os.Signal, 1)

func init() {
	if ConversationServer == nil {
		ConversationServer = &conversationServer{
			conversationSeq: make(map[string]int64),
		}
	}
	//signal.Notify(kafkaQuit, syscall.SIGINT, syscall.SIGTERM)
}

func (c *conversationServer) Start() {
	zlog.Info("ConversationServer开始启动")
	defer func() {
		if r := recover(); r != nil {
			zlog.Error(fmt.Sprintf("conversationServer panic recovered: %v", r))
			// 打印堆栈跟踪
			debug.PrintStack()
		}
	}()
	for {
		kafkaMessage, err := myKafka.KafkaService.ConversationReader.ReadMessage(context.Background())
		if err != nil {
			zlog.Error(err.Error())
		}
		//log.Printf("topic=%s, partition=%d, offset=%d, key=%s, value=%s", kafkaMessage.Topic, kafkaMessage.Partition, kafkaMessage.Offset, kafkaMessage.Key, kafkaMessage.Value)
		zlog.Info(fmt.Sprintf("topic=%s, partition=%d, offset=%d, key=%s, value=%s", kafkaMessage.Topic, kafkaMessage.Partition, kafkaMessage.Offset, kafkaMessage.Key, kafkaMessage.Value))
		data := kafkaMessage.Value

		var chatReq request.ChatMessageRequest
		if err := json.Unmarshal(data, &chatReq); err != nil {
			zlog.Error(err.Error())
		}
		log.Println("原消息为：", string(data), "反序列化后为：", chatReq)

		msgTime := time.Now()
		var conversationId string
		if len(chatReq.ConversationId) == 0 {
			conversationId = c.generateConversationId(chatReq.SendId, chatReq.ReceiveId)
			conversation := model.Conversation{
				ConversationId: conversationId,
				Type:           int8(0),
				Member:         int32(2),
				RecentMsgTime:  msgTime,
			}
			if res := dao.GormDB.Create(&conversation); res.Error != nil {
				zlog.Error(res.Error.Error())
			}
			//用户会话链
			userConversationList := &model.UserConversationList{
				UserId:         chatReq.SendId,
				ConversationId: conversationId,
			}
			if res := dao.GormDB.Create(&userConversationList); res.Error != nil {
				zlog.Error(res.Error.Error())
			}

			userConversationList = &model.UserConversationList{
				UserId:         chatReq.ReceiveId,
				ConversationId: conversationId,
			}
			if res := dao.GormDB.Create(&userConversationList); res.Error != nil {
				zlog.Error(res.Error.Error())
			}

		} else {
			conversationId = chatReq.ConversationId
			if res := dao.GormDB.Model(&model.Conversation{}).
				Where("conversation_id = ?", conversationId).
				Update("recent_msg_time", msgTime); res.Error != nil {
				zlog.Error(res.Error.Error())
			}
		}

		//todo获取msg全局唯一id
		msgId := int64(random.GetRandomInt(12))

		// 存message
		msg := model.Msg{
			MsgId:          msgId,
			ConversationId: conversationId,
			Type:           chatReq.Type,
			Content:        chatReq.Content,
			Url:            "",
			UserId:         chatReq.SendId,
			FileSize:       "0B",
			FileType:       "",
			FileName:       "",
			Status:         message_status_enum.Unsent,
			SendTime:       time.Now(),
			AVData:         "",
		}
		// 对SendAvatar去除前面/static之前的所有内容，防止ip前缀引入
		//message.SendAvatar = normalizePath(message.SendAvatar)
		if res := dao.GormDB.Create(&msg); res.Error != nil {
			zlog.Error(res.Error.Error())
		}

		//存会话消息链
		conversationMsgList := model.ConversationMsgList{
			ConversationId: conversationId,
			MsgId:          msgId,
			Seq:            c.getSeq(conversationId),
		}
		if res := dao.GormDB.Create(&conversationMsgList); res.Error != nil {
			zlog.Error(res.Error.Error())
		}
		//会话表更新时间
		res := dao.GormDB.Model(&model.Conversation{}).
			Where("conversation_id = ?", conversationId).
			Update("recent_msg_time", msgTime)
		if res.Error != nil {
			zlog.Error(res.Error.Error())
		}

		var userIds []int64
		if res := dao.GormDB.Model(&model.UserConversationList{}).
			Where("conversation_id = ?", conversationId).
			Pluck("user_id", &userIds); res.Error != nil {
			zlog.Error(res.Error.Error())
		}

		for _, userId := range userIds {
			userMsg := UserMsg{
				UserId:         userId,
				MsgId:          msg.MsgId,
				ConversationId: msg.ConversationId,
				SendId:         msg.UserId,
				Type:           msg.Type,
				Content:        msg.Content,
				Status:         msg.Status,
				SendTime:       msg.SendTime,
			}
			jsonMessage, err := json.Marshal(userMsg)
			if err != nil {
				zlog.Error(err.Error())
			}
			if err := myKafka.KafkaService.UserWriter.WriteMessages(context.Background(), kafka.Message{
				Key:   []byte(strconv.Itoa(int(userId))),
				Value: jsonMessage,
			}); err != nil {
				zlog.Error(err.Error())
			}
			zlog.Info("已发送消息：" + string(jsonMessage))
		}

	}
}

func (c *conversationServer) generateConversationId(userId1 int64, userId2 int64) string {
	if userId1 < userId2 {
		return fmt.Sprintf("%d-%d", userId1, userId2)
	} else {
		return fmt.Sprintf("%d-%d", userId2, userId1)
	}
}

func (c *conversationServer) getSeq(conversationId string) int64 {
	if c.conversationSeq[conversationId] == 0 {
		var seq int64
		// 只使用Pluck，更简洁
		_ = dao.GormDB.Model(&model.ConversationMsgList{}).
			Where("conversation_id = ?", conversationId).
			Order("seq DESC").
			Limit(1).
			Pluck("seq", &seq).Error
		c.conversationSeq[conversationId] = seq
	}
	c.conversationSeq[conversationId]++
	return c.conversationSeq[conversationId]
}
