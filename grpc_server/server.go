package grpc_server

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"kama_chat_server/internal/dto/request"
	"kama_chat_server/internal/service/gorm"
	pb "kama_chat_server/pb"
	"kama_chat_server/pkg/zlog"
	"net"
)

type server struct {
	pb.UnimplementedKamaChatServer
}

var Server = new(server)

func (s *server) Start() {
	zlog.Info("gRPC server开始启动")
	listen, err := net.Listen("tcp", ":9090")
	if err != nil {
		zlog.Error(err.Error())
	}
	grpcServer := grpc.NewServer()
	pb.RegisterKamaChatServer(grpcServer, &server{})
	err = grpcServer.Serve(listen)
	if err != nil {
		zlog.Error(err.Error())
	}
	zlog.Info("gRPC server停止")
}

func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.UserResponse, error) {
	zlog.Info("调用register")
	registerReq := request.RegisterRequest{
		Telephone: req.Telephone,
		Password:  req.Password,
		Nickname:  req.Nickname,
	}

	message, userInfo, ret := gorm.UserInfoService.Register(registerReq)

	if ret == 0 {
		res := &pb.UserResponse{
			UserId:    userInfo.UserId,
			Nickname:  userInfo.Nickname,
			Telephone: userInfo.Telephone,
			Avatar:    userInfo.Avatar,
			Gender:    int32(userInfo.Gender),
			Signature: userInfo.Signature,
			Birthday:  userInfo.Birthday,
			CreatedAt: userInfo.CreatedAt,
			IsAdmin:   int32(userInfo.IsAdmin),
			Status:    int32(userInfo.Status),
		}
		return res, nil
	} else {
		return nil, errToRpc(message, ret)
	}
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.UserResponse, error) {
	zlog.Info("调用login")
	loginReq := request.LoginRequest{
		Telephone: req.Telephone,
		Password:  req.Password,
	}
	message, userInfo, ret := gorm.UserInfoService.Login(loginReq)

	if ret == 0 {
		res := &pb.UserResponse{
			UserId:    userInfo.UserId,
			Nickname:  userInfo.Nickname,
			Telephone: userInfo.Telephone,
			Avatar:    userInfo.Avatar,
			Gender:    int32(userInfo.Gender),
			Signature: userInfo.Signature,
			Birthday:  userInfo.Birthday,
			CreatedAt: userInfo.CreatedAt,
			IsAdmin:   int32(userInfo.IsAdmin),
			Status:    int32(userInfo.Status),
		}
		return res, nil
	} else {
		return nil, errToRpc(message, ret)
	}
}

func errToRpc(message string, ret int) error {
	if ret == -1 {
		return status.Error(codes.Internal, message)
	} else {
		return status.Error(codes.NotFound, message)
	}
}

func (s *server) GetMessageBySeq(ctx context.Context, req *pb.GetMessageBySeqRequest) (*pb.ResponseForGetMessageBySeq, error) {
	var message string
	var ret int
	var rsp *pb.ResponseForGetMessageBySeq
	if req.EndSeq == -1 {
		message, rsp, ret = gorm.MessageService.GetMessageAfterSeq(req.UserId, req.StartSeq)
	} else {
		message, rsp, ret = gorm.MessageService.GetMessageBetween(req.UserId, req.StartSeq, req.EndSeq)
	}
	if ret == 0 {
		return rsp, nil
	} else {
		return nil, errToRpc(message, ret)
	}
}

func (s *server) GetConversationList(ctx context.Context, req *pb.UserIdRequest) (*pb.ResponseForGetConversationList, error) {
	message, rsp, ret := gorm.ConversationService.GetConversationList(req.UserId)
	if ret == 0 {
		return rsp, nil
	} else {
		return nil, errToRpc(message, ret)
	}
}

func (s *server) GetFriendList(ctx context.Context, req *pb.UserIdRequest) (*pb.RespondForGetFriendList, error) {
	message, rsp, ret := gorm.UserContactService.GetFriendList(req.UserId)
	if ret == 0 {
		return rsp, nil
	} else {
		return nil, errToRpc(message, ret)
	}
}

//rpc applyFriend (ApplyFriendRequest) returns (Response) {}
//rpc getApplyList (UserIdRequest) returns (ResponseForApplyList) {}
//rpc respondToApply (RespondToApply) returns (Response) {}
