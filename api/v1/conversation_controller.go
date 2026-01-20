package v1

import (
	"github.com/gin-gonic/gin"
	"kama_chat_server/internal/dto/request"
	"kama_chat_server/internal/service/gorm"
	"kama_chat_server/pkg/constants"
	"kama_chat_server/pkg/zlog"
	"net/http"
)

// OpenConversation 打开会话
//func OpenConversation(c *gin.Context) {
//	var openSessionReq request.UserFriendRequest
//	if err := c.BindJSON(&openSessionReq); err != nil {
//		zlog.Error(err.Error())
//		c.JSON(http.StatusOK, gin.H{
//			"code":    500,
//			"message": constants.SYSTEM_ERROR,
//		})
//		return
//	}
//	message, sessionId, ret := gorm.SessionService.OpenSession(openSessionReq)
//	JsonBack(c, message, ret, sessionId)
//}

func GetConversationList(c *gin.Context) {
	var req request.UserRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, conversationList, ret := gorm.ConversationService.GetConversationList(req.UserId)
	JsonBack(c, message, ret, conversationList)
}

// DeleteSession 删除会话
func DeleteConversation(c *gin.Context) {
	var req request.UserConversationRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, ret := gorm.ConversationService.DeleteConversation(req.UserId, req.ConversationId)
	JsonBack(c, message, ret, nil)
}

// CheckOpenSessionAllowed 检查是否可以打开会话
func CheckOpenConversationAllowed(c *gin.Context) {
	var req request.UserFriendRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, res, ret := gorm.ConversationService.CheckOpenConversationAllowed(req.UserId, req.FriendId)
	JsonBack(c, message, ret, res)
}
