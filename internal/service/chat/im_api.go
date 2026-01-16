package chat

//import (
//	"encoding/json"
//	"fmt"
//	"github.com/segmentio/kafka-go"
//	"kama_chat_server/internal/dao"
//	"kama_chat_server/internal/dto/request"
//	"kama_chat_server/internal/model"
//	myKafka "kama_chat_server/internal/service/kafka"
//	"kama_chat_server/pkg/enum/message/message_type_enum"
//	"kama_chat_server/pkg/util/random"
//	"kama_chat_server/pkg/zlog"
//	"time"
//)
//
//type imApi struct{}
//
//var ImApi = new(imApi)
//
//func (i *imApi) SendMessage(chatReq request.ChatMessageRequest) {
//	if chatReq.Type == message_type_enum.Text {
//		msgTime := time.Now()
//		var conversationId string
//		if len(chatReq.ConversationId) == 0 {
//			conversationId = i.generateConversationId(chatReq.SendId, chatReq.ReceiveId)
//			conversation := model.Conversation{
//				ConversationId: conversationId,
//				Type:           int8(0),
//				Member:         int32(2),
//				RecentMsgTime:  msgTime,
//			}
//			if res := dao.GormDB.Create(&conversation); res.Error != nil {
//				zlog.Error(res.Error.Error())
//			}
//		} else {
//			conversationId = chatReq.ConversationId
//			if res := dao.GormDB.Model(&model.Conversation{}).
//				Where("conversation_id = ?", conversationId).
//				Update("recent_msg_time", msgTime); res.Error != nil {
//				zlog.Error(res.Error.Error())
//			}
//		}
//		//todo获取msg全局唯一id
//		msgId := int64(random.GetRandomInt(12))
//		conversationMsgRequest := &request.ConversationMsgRequest{
//			ConversationId: conversationId,
//			Type:           chatReq.Type,
//			Content:        chatReq.Content,
//			UserId:         chatReq.SendId,
//			MsgId:          msgId,
//			SendTime:       msgTime,
//		}
//
//		jsonMessage, err := json.Marshal(conversationMsgRequest)
//		if err != nil {
//			zlog.Error(err.Error())
//		}
//
//		if err := myKafka.KafkaService.ConversationWriter.WriteMessages(ctx, kafka.Message{
//			Key:   []byte(conversationId),
//			Value: jsonMessage,
//		}); err != nil {
//			zlog.Error(err.Error())
//		}
//		zlog.Info("已发送消息：" + string(jsonMessage))
//	}
//}
//
//func (i *imApi) generateConversationId(userId1 int64, userId2 int64) string {
//	if userId1 < userId2 {
//		return fmt.Sprintf("%d-%d", userId1, userId2)
//	} else {
//		return fmt.Sprintf("%d-%d", userId2, userId1)
//	}
//}
