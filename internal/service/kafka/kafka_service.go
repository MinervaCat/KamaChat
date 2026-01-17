package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	myconfig "kama_chat_server/internal/config"
	"kama_chat_server/pkg/zlog"
	"time"
)

var ctx = context.Background()

type kafkaService struct {
	ConversationWriter *kafka.Writer
	ConversationReader *kafka.Reader
	KafkaConn          *kafka.Conn
	UserWriter         *kafka.Writer
	UserReader         *kafka.Reader
}

var KafkaService = new(kafkaService)

func (k *kafkaService) KafkaInit2() {
	//k.CreateTopic()
	kafkaConfig := myconfig.GetConfig().KafkaConfig
	k.ConversationWriter = &kafka.Writer{
		Addr:                   kafka.TCP(kafkaConfig.HostPort),
		Topic:                  kafkaConfig.ConversationImTopic,
		Balancer:               &kafka.Hash{},
		WriteTimeout:           kafkaConfig.Timeout * time.Second,
		RequiredAcks:           kafka.RequireNone,
		AllowAutoTopicCreation: false,
	}
}

// KafkaInit 初始化kafka
func (k *kafkaService) KafkaInit() {
	//k.CreateTopic()
	kafkaConfig := myconfig.GetConfig().KafkaConfig
	k.ConversationWriter = &kafka.Writer{
		Addr:                   kafka.TCP(kafkaConfig.HostPort),
		Topic:                  kafkaConfig.ConversationImTopic,
		Balancer:               &kafka.Hash{},
		WriteTimeout:           kafkaConfig.Timeout * time.Second,
		RequiredAcks:           kafka.RequireNone,
		AllowAutoTopicCreation: false,
	}
	k.ConversationReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaConfig.HostPort},
		Topic:          kafkaConfig.ConversationImTopic,
		CommitInterval: kafkaConfig.Timeout * time.Second,
		GroupID:        "chat",
		StartOffset:    kafka.LastOffset,
	})
	k.UserWriter = &kafka.Writer{
		Addr:                   kafka.TCP(kafkaConfig.HostPort),
		Topic:                  kafkaConfig.UserImTopic,
		Balancer:               &kafka.Hash{},
		WriteTimeout:           kafkaConfig.Timeout * time.Second,
		RequiredAcks:           kafka.RequireNone,
		AllowAutoTopicCreation: false,
	}
	k.UserReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaConfig.HostPort},
		Topic:          kafkaConfig.UserImTopic,
		CommitInterval: kafkaConfig.Timeout * time.Second,
		GroupID:        "chat",
		StartOffset:    kafka.LastOffset,
	})
}

func (k *kafkaService) KafkaClose() {
	if err := k.ConversationWriter.Close(); err != nil {
		zlog.Error(err.Error())
	}
	if err := k.ConversationReader.Close(); err != nil {
		zlog.Error(err.Error())
	}
	if err := k.UserWriter.Close(); err != nil {
		zlog.Error(err.Error())
	}
	if err := k.UserReader.Close(); err != nil {
		zlog.Error(err.Error())
	}
}

//// CreateTopic 创建topic
//func (k *kafkaService) CreateTopic() {
//	// 如果已经有topic了，就不创建了
//	kafkaConfig := myconfig.GetConfig().KafkaConfig
//
//	chatTopic := kafkaConfig.ChatTopic
//
//	// 连接至任意kafka节点
//	var err error
//	k.KafkaConn, err = kafka.Dial("tcp", kafkaConfig.HostPort)
//	if err != nil {
//		zlog.Error(err.Error())
//	}
//
//	topicConfigs := []kafka.TopicConfig{
//		{
//			Topic:             chatTopic,
//			NumPartitions:     kafkaConfig.Partition,
//			ReplicationFactor: 1,
//		},
//	}
//
//	// 创建topic
//	if err = k.KafkaConn.CreateTopics(topicConfigs...); err != nil {
//		zlog.Error(err.Error())
//	}
//
//}
