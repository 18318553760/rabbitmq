
/**
* @program: Go
*
* @description:mq生产者,包与包之间不要相互引用，否则会报错，所以包的调用是独立的
*
* @author: Mr.chen
*
* @create: 2019-12-25 10:25
**/
package mq
import (
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
	"github.com/streadway/amqp"
	cfg "rabbitmq/config"
)
type RabbitmqConfig struct {
	AsyncTransferEnable        bool
	RabbitURL     string
	TransExchangeName   string
	TransRegisterQueueName string
	TransRegisterErrQueueName string
	TransRegisterRoutingKey string
}

var (
	conn *amqp.Connection
	Rabbitmq  *RabbitmqConfig
	channel *amqp.Channel
)
// 如果异常关闭，会接收通知
var notifyClose chan *amqp.Error
func InitMqConfig() (err error) {
	// 初始化结构体
	Rabbitmq = &RabbitmqConfig{}
	// 获取rabbitmq参数

	conf, err := config.NewConfig(cfg.ConfType, cfg.Filename)
	if err != nil {
		fmt.Errorf("load config failed, err:%v", err)
		return
	}
	AsyncTransferEnable,err := conf.Bool("rabbitmq::AsyncTransferEnable")
	if err !=nil {
		err =  fmt.Errorf("init config failed, RabbitmqConfig[%s]  config is null", AsyncTransferEnable)
		return
	}
	Rabbitmq.AsyncTransferEnable = AsyncTransferEnable
	Rabbitmq.RabbitURL = conf.String("rabbitmq::RabbitURL")
	if len(Rabbitmq.RabbitURL) == 0 {
		err = fmt.Errorf("init config failed, RabbitmqConfig[%s]  config is null", Rabbitmq.RabbitURL)
		return
	}

	Rabbitmq.TransExchangeName= conf.String("rabbitmq::TransExchangeName")
	if len(Rabbitmq.TransExchangeName) == 0 {
		err = fmt.Errorf("init config failed, RabbitmqConfig[%s]  config is null", Rabbitmq.TransExchangeName)
		return
	}
	Rabbitmq.TransRegisterQueueName =conf.String("rabbitmq::TransRegisterQueueName")
	if len(Rabbitmq.TransRegisterQueueName) == 0 {
		err = fmt.Errorf("init config failed, RabbitmqConfig[%s]  config is null", Rabbitmq.TransRegisterQueueName)
		return
	}

	Rabbitmq.TransRegisterErrQueueName = conf.String("rabbitmq::TransRegisterErrQueueName")
	if len(Rabbitmq.TransRegisterErrQueueName) == 0 {
		err = fmt.Errorf("init config failed, RabbitmqConfig[%s]  config is null", Rabbitmq.TransRegisterErrQueueName)
		return
	}
	Rabbitmq.TransRegisterRoutingKey = conf.String("rabbitmq::TransRegisterRoutingKey")
	if len(Rabbitmq.TransRegisterRoutingKey) == 0 {
		err = fmt.Errorf("init config failed, RabbitmqConfig[%s]  config is null", Rabbitmq.TransRegisterRoutingKey)
		return
	}
	fmt.Println(Rabbitmq)
	return
}
func init() {
	err := InitMqConfig()  // 此配置必须运行在同一个main,否则无法获取配置信息
	if err != nil {
		logs.Error(err)
		return
	}
	//webConfig := base.GetConfig()
	//redisConf = webConfig.RedisConf
	//pool = newRedisPool()

	// 是否开启异步转移功能，开启时才初始化rabbitMQ连接
	if !Rabbitmq.AsyncTransferEnable {
		return
	}
	if initChannel() {
		channel.NotifyClose(notifyClose)
	}
	// 断线自动重连
	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				channel = nil
				logs.Debug("onNotifyChannelClosed: %+v\n", msg)

				initChannel()
			}
		}
	}()
}

func initChannel() bool {
	if channel != nil {
		return true
	}
	fmt.Println(Rabbitmq.RabbitURL)
	conn, err := amqp.Dial(Rabbitmq.RabbitURL)
	if err != nil {
		logs.Error(err.Error())
		return false
	}

	channel, err = conn.Channel()
	if err != nil {
		logs.Error(err.Error())
		return false
	}
	return true
}
// Publish : 发布消息
func Publish(exchange, routingKey string, msg []byte) bool {
	if !initChannel() {
		return false
	}
	if nil == channel.Publish(
		exchange,
		routingKey,
		false, // 如果没有对应的queue, 就会丢弃这条小心
		false, //
		amqp.Publishing{
			ContentType: "text/plain",
			Body: msg}) {
		return true
	}
	return false
}
func GetConfig() *RabbitmqConfig {
	return Rabbitmq
}
