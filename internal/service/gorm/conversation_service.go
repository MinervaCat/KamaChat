package gorm

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"kama_chat_server/internal/dao"
	"kama_chat_server/internal/dto/respond"
	"kama_chat_server/internal/model"
	myredis "kama_chat_server/internal/service/redis"
	"kama_chat_server/pkg/constants"
	"kama_chat_server/pkg/enum/contact/contact_status_enum"
	"kama_chat_server/pkg/enum/user_info/user_status_enum"
	"kama_chat_server/pkg/zlog"
	"time"
)

type conversationService struct{}

var ConversationService = conversationService{}

func getConversation(conversationId string) (*model.Conversation, error) {
	var conversation model.Conversation
	if res := dao.GormDB.Where("conversation_id=?", conversationId).First(&conversation); res.Error != nil {
		return nil, res.Error
	}
	return &conversation, nil
}

// GetUserSessionList 获取用户会话列表
func (s *conversationService) GetConversationList(ownerId int64) (string, []respond.ConversationListRespond, int) {
	rspString, err := myredis.GetKeyNilIsErr(fmt.Sprintf("conversation_list_%d", ownerId))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			var conversationList []model.UserConversationList
			if res := dao.GormDB.Order("last_read_seq DESC").Where("user_id = ?", ownerId).Find(&conversationList); res.Error != nil {
				if errors.Is(res.Error, gorm.ErrRecordNotFound) {
					zlog.Info("未创建用户会话")
					return "未创建用户会话", nil, 0
				} else {
					zlog.Error(res.Error.Error())
					return constants.SYSTEM_ERROR, nil, -1
				}
			}
			var conversationListRsp []respond.ConversationListRespond
			for _, conversation := range conversationList {
				if con, err := getConversation(conversation.ConversationId); err != nil {
					zlog.Error(err.Error())
				} else {
					conversationListRsp = append(conversationListRsp, respond.ConversationListRespond{
						ConversationId: conversation.ConversationId,
						Avatar:         con.Avatar,
						Type:           con.Type,
						Member:         con.Member,
						RecentMsgTime:  con.RecentMsgTime,
						LastReadSeq:    conversation.LastReadSeq,
						NotifyType:     conversation.NotifyType,
						IsTop:          conversation.IsTop,
					})
				}
			}

			rspString, err := json.Marshal(conversationListRsp)
			if err != nil {
				zlog.Error(err.Error())
			}
			if err := myredis.SetKeyEx(fmt.Sprintf("conversation_list_%d", ownerId), string(rspString), time.Minute*constants.REDIS_TIMEOUT); err != nil {
				zlog.Error(err.Error())
			}
			return "获取成功", conversationListRsp, 0
		} else {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, nil, -1
		}
	}
	var rsp []respond.ConversationListRespond
	if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
		zlog.Error(err.Error())
	}
	return "获取成功", rsp, 0
}

func (s *conversationService) CheckOpenConversationAllowed(sendId, receiveId int64) (string, bool, int) {
	var relation model.Relation
	if res := dao.GormDB.Where("user_id = ? and friend_id = ?", sendId, receiveId).First(&relation); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, false, -1
	}
	if relation.Status == contact_status_enum.BE_BLACK {
		return "已被对方拉黑，无法发起会话", false, -2
	} else if relation.Status == contact_status_enum.BLACK {
		return "已拉黑对方，先解除拉黑状态才能发起会话", false, -2
	}

	var user model.User
	if res := dao.GormDB.Where("user_id = ?", receiveId).First(&user); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, false, -1
	}
	if user.Status == user_status_enum.DISABLE {
		zlog.Info("对方已被禁用，无法发起会话")
		return "对方已被禁用，无法发起会话", false, -2
	}
	return "可以发起会话", true, 0
}

// DeleteSession 删除会话
func (s *conversationService) DeleteConversation(userId int64, conversationId string) (string, int) {

	var userConversation model.UserConversationList
	if res := dao.GormDB.Where("conversation_id = ? AND user_id = ?", conversationId, userId).Delete(&userConversation); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, -1
	}
	//session.DeletedAt.Valid = true
	//session.DeletedAt.Time = time.Now()
	//if res := dao.GormDB.Save(&session); res.Error != nil {
	//	zlog.Error(res.Error.Error())
	//	return constants.SYSTEM_ERROR, -1
	//}
	//if err := myredis.DelKeysWithSuffix(sessionId); err != nil {
	//	zlog.Error(err.Error())
	//}
	if err := myredis.DelKeysWithPattern(fmt.Sprintf("conversation_list_%d", userId)); err != nil {
		zlog.Error(err.Error())
	}
	return "删除成功", 0
}
