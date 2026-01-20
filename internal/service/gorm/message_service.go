package gorm

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"kama_chat_server/internal/config"
	"kama_chat_server/internal/dao"
	"kama_chat_server/internal/dto/respond"
	"kama_chat_server/internal/model"
	"kama_chat_server/pkg/constants"
	"kama_chat_server/pkg/zlog"
	"os"
	"path/filepath"
)

type messageService struct {
}

var MessageService = new(messageService)

func (m *messageService) GetMessageAfterSeq(userId int64, seq int64) (string, []respond.MessageRespond, int) {
	var messageList []model.UserMsgList
	if res := dao.GormDB.Where("user_id = ? AND seq >= ?", userId, seq).Find(&messageList); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, nil, -1
	}

	var rspList []respond.MessageRespond
	for _, message := range messageList {
		if msg, err := getMessage(message.MsgId); err != nil {
			zlog.Error(err.Error())
			continue
		} else {
			rspList = append(rspList, respond.MessageRespond{
				MsgId:          message.MsgId,
				ConversationId: message.ConversationId,
				Seq:            message.Seq,
				SendId:         msg.UserId,
				Type:           msg.Type,
				Content:        msg.Content,
				Status:         msg.Status,
				SendTime:       msg.SendTime,
			})
		}
	}
	return "获取聊天记录成功", rspList, 0
}

func (m *messageService) GetMessageBetween(userId, firstSeq, lastSeq int64) (string, []respond.MessageRespond, int) {
	var messageList []model.UserMsgList
	if res := dao.GormDB.Where("user_id = ? AND seq >= ? AND seq <= ?", userId, firstSeq, lastSeq).Find(&messageList); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, nil, -1
	}
	var rspList []respond.MessageRespond
	for _, message := range messageList {
		if msg, err := getMessage(message.MsgId); err != nil {
			zlog.Error(err.Error())
			continue
		} else {
			rspList = append(rspList, respond.MessageRespond{
				MsgId:          message.MsgId,
				ConversationId: message.ConversationId,
				Seq:            message.Seq,
				SendId:         msg.UserId,
				Type:           msg.Type,
				Content:        msg.Content,
				Status:         msg.Status,
				SendTime:       msg.SendTime,
			})
		}
	}
	return "获取聊天记录成功", rspList, 0
}

func (m *messageService) GetMessageBySeq(userId, seq int64) (string, respond.MessageRespond, int) {
	var message model.UserMsgList
	if res := dao.GormDB.Where("user_id = ? AND seq = ?", userId, seq).Find(&message); res.Error != nil {
		zlog.Error(res.Error.Error())
		return constants.SYSTEM_ERROR, respond.MessageRespond{}, -1
	}
	var rsp respond.MessageRespond

	if msg, err := getMessage(message.MsgId); err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, rsp, -1
	} else {
		rsp = respond.MessageRespond{
			MsgId:          message.MsgId,
			ConversationId: message.ConversationId,
			Seq:            message.Seq,
			SendId:         msg.UserId,
			Type:           msg.Type,
			Content:        msg.Content,
			Status:         msg.Status,
			SendTime:       msg.SendTime,
		}
	}

	return "获取聊天记录成功", rsp, 0
}

func getMessage(msgId int64) (*model.Msg, error) {
	var message model.Msg
	if res := dao.GormDB.Where("msg_id = ?", msgId).Find(&message); res.Error != nil {
		zlog.Error(res.Error.Error())
		return nil, res.Error
	}
	return &message, nil
}

// UploadAvatar 上传头像
func (m *messageService) UploadAvatar(c *gin.Context) (string, int) {
	if err := c.Request.ParseMultipartForm(constants.FILE_MAX_SIZE); err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}
	mForm := c.Request.MultipartForm
	for key, _ := range mForm.File {
		file, fileHeader, err := c.Request.FormFile(key)
		if err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		defer file.Close()
		zlog.Info(fmt.Sprintf("文件名：%s，文件大小：%d", fileHeader.Filename, fileHeader.Size))
		// 原来Filename应该是213451545.xxx，将Filename修改为avatar_ownerId.xxx
		ext := filepath.Ext(fileHeader.Filename)
		zlog.Info(ext)
		localFileName := config.GetConfig().StaticAvatarPath + "/" + fileHeader.Filename
		out, err := os.Create(localFileName)
		if err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		zlog.Info("完成文件上传")
	}
	return "上传成功", 0
}

// UploadFile 上传文件
func (m *messageService) UploadFile(c *gin.Context) (string, int) {
	if err := c.Request.ParseMultipartForm(constants.FILE_MAX_SIZE); err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}
	mForm := c.Request.MultipartForm
	for key, _ := range mForm.File {
		file, fileHeader, err := c.Request.FormFile(key)
		if err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		defer file.Close()
		zlog.Info(fmt.Sprintf("文件名：%s，文件大小：%d", fileHeader.Filename, fileHeader.Size))
		// 原来Filename应该是213451545.xxx，将Filename修改为avatar_ownerId.xxx
		ext := filepath.Ext(fileHeader.Filename)
		zlog.Info(ext)
		localFileName := config.GetConfig().StaticFilePath + "/" + fileHeader.Filename
		out, err := os.Create(localFileName)
		if err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		zlog.Info("完成文件上传")
	}
	return "上传成功", 0
}
