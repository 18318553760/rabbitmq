# 概述

## 1、安装rabbitmq

具体安装方法请移步本人之前写的文章：<http://cblog.1024.company/3-14.html> 

go+rabbitmq资料：<https://github.com/rabbitmq/>  ,<https://github.com/rabbitmq/rabbitmq-tutorials/tree/master/go> 

## 2、rabbitmq的介绍

AMQP，即Advanced Message Queuing Protocol，高级消息队列协议，是应用层协议的一个开放标准，为面向消息的[中间件](http://www.diggerplus.org/archives/tag/%e4%b8%ad%e9%97%b4%e4%bb%b6)设计。消息中间件主要用于组件之间的解耦，消息的发送者无需知道消息使用者的存在，反之亦然。
AMQP的主要特征是面向消息、队列、路由（包括点对点和发布/订阅）、可靠性、安全。
RabbitMQ是一个开源的AMQP实现，服务器端用Erlang语言编写，支持多种客户端，如：Python、Ruby、.NET、Java、JMS、C、PHP、ActionScript、XMPP、STOMP等，支持AJAX。用于在分布式系统中存储转发消息，在易用性、扩展性、高可用性等方面表现不俗。
下面将重点介绍RabbitMQ中的一些基础概念，了解了这些概念，是使用好RabbitMQ的基础。

### ConnectionFactory、Connection、Channel

ConnectionFactory、Connection、Channel都是RabbitMQ对外提供的API中最基本的对象。Connection是RabbitMQ的socket链接，它封装了socket协议相关部分逻辑。ConnectionFactory为Connection的制造工厂。
Channel是我们与RabbitMQ打交道的最重要的一个接口，我们大部分的业务操作是在Channel这个接口中完成的，包括定义Queue、定义Exchange、绑定Queue与Exchange、发布消息等。

### Queue

Queue（队列）是RabbitMQ的内部对象，用于存储消息，用下图表示。
![img](https://img2018.cnblogs.com/blog/774371/201908/774371-20190819103814085-804287529.png)

RabbitMQ中的消息都只能存储在Queue中，生产者（下图中的P）生产消息并最终投递到Queue中，消费者（下图中的C）可以从Queue中获取消息并消费。

![img](https://img2018.cnblogs.com/blog/774371/201908/774371-20190819103830954-867723738.png)

多个消费者可以订阅同一个Queue，这时Queue中的消息会被平均分摊给多个消费者进行处理，而不是每个消费者都收到所有的消息并处理。

### Message acknowledgment

在实际应用中，可能会发生消费者收到Queue中的消息，但没有处理完成就宕机（或出现其他意外）的情况，这种情况下就可能会导致消息丢失。为了避免这种情况发生，我们可以要求消费者在消费完消息后发送一个回执给RabbitMQ，RabbitMQ收到消息回执（Message acknowledgment）后才将该消息从Queue中移除；如果RabbitMQ没有收到回执并检测到消费者的RabbitMQ连接断开，则RabbitMQ会将该消息发送给其他消费者（如果存在多个消费者）进行处理。这里不存在timeout概念，一个消费者处理消息时间再长也不会导致该消息被发送给其他消费者，除非它的RabbitMQ连接断开。
这里会产生另外一个问题，如果我们的开发人员在处理完业务逻辑后，忘记发送回执给RabbitMQ，这将会导致严重的bug——Queue中堆积的消息会越来越多；消费者重启后会重复消费这些消息并重复执行业务逻辑…

### Message durability

如果我们希望即使在RabbitMQ服务重启的情况下，也不会丢失消息，我们可以将Queue与Message都设置为可持久化的（durable），这样可以保证绝大部分情况下我们的RabbitMQ消息不会丢失。但依然解决不了小概率丢失事件的发生（比如RabbitMQ服务器已经接收到生产者的消息，但还没来得及持久化该消息时RabbitMQ服务器就断电了），如果我们需要对这种小概率事件也要管理起来，那么我们要用到事务。由于这里仅为RabbitMQ的简单介绍，所以这里将不讲解RabbitMQ相关的事务。

### Prefetch count

前面我们讲到如果有多个消费者同时订阅同一个Queue中的消息，Queue中的消息会被平摊给多个消费者。这时如果每个消息的处理时间不同，就有可能会导致某些消费者一直在忙，而另外一些消费者很快就处理完手头工作并一直空闲的情况。我们可以通过设置prefetchCount来限制Queue每次发送给每个消费者的消息数，比如我们设置prefetchCount=1，则Queue每次给每个消费者发送一条消息；消费者处理完这条消息后Queue会再给该消费者发送一条消息。

![img](https://img2018.cnblogs.com/blog/774371/201908/774371-20190819104007647-1027286628.png)

### Exchange

在上一节我们看到生产者将消息投递到Queue中，实际上这在RabbitMQ中这种事情永远都不会发生。实际的情况是，生产者将消息发送到Exchange（交换器，下图中的X），由Exchange将消息路由到一个或多个Queue中（或者丢弃）。

![img](https://img2018.cnblogs.com/blog/774371/201908/774371-20190819104028789-412276700.png)

Exchange是按照什么逻辑将消息路由到Queue的？这个将在Binding一节介绍。
RabbitMQ中的Exchange有四种类型，不同的类型有着不同的路由策略，这将在Exchange Types一节介绍。

### routing key

生产者在将消息发送给Exchange的时候，一般会指定一个routing key，来指定这个消息的路由规则，而这个routing key需要与Exchange Type及binding key联合使用才能最终生效。
在Exchange Type与binding key固定的情况下（在正常使用时一般这些内容都是固定配置好的），我们的生产者就可以在发送消息给Exchange时，通过指定routing key来决定消息流向哪里。
RabbitMQ为routing key设定的长度限制为255 bytes。

### Binding

RabbitMQ中通过Binding将Exchange与Queue关联起来，这样RabbitMQ就知道如何正确地将消息路由到指定的Queue了。
![img](https://img2018.cnblogs.com/blog/774371/201908/774371-20190819104128931-1338459538.png)

 

### Binding key

在绑定（Binding）Exchange与Queue的同时，一般会指定一个binding key；消费者将消息发送给Exchange时，一般会指定一个routing key；当binding key与routing key相匹配时，消息将会被路由到对应的Queue中。这个将在Exchange Types章节会列举实际的例子加以说明。
在绑定多个Queue到同一个Exchange的时候，这些Binding允许使用相同的binding key。
binding key 并不是在所有情况下都生效，它依赖于Exchange Type，比如fanout类型的Exchange就会无视binding key，而是将消息路由到所有绑定到该Exchange的Queue。

### Exchange Types

RabbitMQ常用的Exchange Type有fanout、direct、topic、headers这四种（AMQP规范里还提到两种Exchange Type，分别为system与自定义，这里不予以描述），下面分别进行介绍。

### fanout

fanout类型的Exchange路由规则非常简单，它会把所有发送到该Exchange的消息路由到所有与它绑定的Queue中。
![img](https://img2018.cnblogs.com/blog/774371/201908/774371-20190819104152177-2053988251.png)

 

上图中，生产者（P）发送到Exchange（X）的所有消息都会路由到图中的两个Queue，并最终被两个消费者（C1与C2）消费。

### direct

direct类型的Exchange路由规则也很简单，它会把消息路由到那些binding key与routing key完全匹配的Queue中。

![img](https://img2018.cnblogs.com/blog/774371/201908/774371-20190819104210818-1771762193.png)

 

以上图的配置为例，我们以routingKey=”error”发送消息到Exchange，则消息会路由到Queue1（amqp.gen-S9b…，这是由RabbitMQ自动生成的Queue名称）和Queue2（amqp.gen-Agl…）；如果我们以routingKey=”info”或routingKey=”warning”来发送消息，则消息只会路由到Queue2。如果我们以其他routingKey发送消息，则消息不会路由到这两个Queue中。

### topic

前面讲到direct类型的Exchange路由规则是完全匹配binding key与routing key，但这种严格的匹配方式在很多情况下不能满足实际业务需求。topic类型的Exchange在匹配规则上进行了扩展，它与direct类型的Exchage相似，也是将消息路由到binding key与routing key相匹配的Queue中，但这里的匹配规则有些不同，它约定：

- routing key为一个句点号“. ”分隔的字符串（我们将被句点号“. ”分隔开的每一段独立的字符串称为一个单词），如“stock.usd.nyse”、“nyse.vmw”、“quick.orange.rabbit”
- binding key与routing key一样也是句点号“. ”分隔的字符串
- binding key中可以存在两种特殊字符“*”与“#”，用于做模糊匹配，其中“*”用于匹配一个单词，“#”用于匹配多个单词（可以是零个）

 

![img](https://img2018.cnblogs.com/blog/774371/201908/774371-20190819104233644-253637000.png)

 

以上图中的配置为例，routingKey=”quick.orange.rabbit”的消息会同时路由到Q1与Q2，routingKey=”lazy.orange.fox”的消息会路由到Q1，routingKey=”lazy.brown.fox”的消息会路由到Q2，routingKey=”lazy.pink.rabbit”的消息会路由到Q2（只会投递给Q2一次，虽然这个routingKey与Q2的两个bindingKey都匹配）；routingKey=”quick.brown.fox”、routingKey=”orange”、routingKey=”quick.orange.male.rabbit”的消息将会被丢弃，因为它们没有匹配任何bindingKey。

### headers

headers类型的Exchange不依赖于routing key与binding key的匹配规则来路由消息，而是根据发送的消息内容中的headers属性进行匹配。
在绑定Queue与Exchange时指定一组键值对；当消息发送到Exchange时，RabbitMQ会取到该消息的headers（也是一个键值对的形式），对比其中的键值对是否完全匹配Queue与Exchange绑定时指定的键值对；如果完全匹配则消息会路由到该Queue，否则不会路由到该Queue。
该类型的Exchange没有用到过（不过也应该很有用武之地），所以不做介绍。

### RPC

MQ本身是基于异步的消息处理，前面的示例中所有的生产者（P）将消息发送到RabbitMQ后不会知道消费者（C）处理成功或者失败（甚至连有没有消费者来处理这条消息都不知道）。
但实际的应用场景中，我们很可能需要一些同步处理，需要同步等待服务端将我的消息处理完成后再进行下一步处理。这相当于RPC（Remote Procedure Call，远程过程调用）。在RabbitMQ中也支持RPC。

![img](https://img2018.cnblogs.com/blog/774371/201908/774371-20190819104253450-1490098886.png)

 

RabbitMQ中实现RPC的机制是：

- 客户端发送请求（消息）时，在消息的属性（MessageProperties，在AMQP协议中定义了14中properties，这些属性会随着消息一起发送）中设置两个值replyTo（一个Queue名称，用于告诉服务器处理完成后将通知我的消息发送到这个Queue中）和correlationId（此次请求的标识号，服务器处理完成后需要将此属性返还，客户端将根据这个id了解哪条请求被成功执行了或执行失败）
- 服务器端收到消息并处理
- 服务器端处理完消息后，将生成一条应答消息到replyTo指定的Queue，同时带上correlationId属性
- 客户端之前已订阅replyTo指定的Queue，从中收到服务器的应答消息后，根据其中的correlationId属性分析哪条请求被执行了，根据执行结果进行后续业务处理



## 3、RabbitMQ在golang中实践

### 目录结构：

![1578303210200](C:\Users\Administrator\Desktop\go语言\go\基础语法\image\f32f96a81ed37feff4bb6fedaba49e5.png)

新建项目rabbitmq

进入rabbitmq目录运行go mod init ，创建go.mod文件，运行go mod tidy添加和剔除相关依赖，go mod vendor将依赖添加到本项目中

### main.php



```
/**
* @program: Go
*
* @description:主程序
*
* @author: Mr.chen
*
* @create: 2020-01-06 16:12
**/
package main
import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
	"math/rand"
	cfg "rabbitmq/config"
	"rabbitmq/mq"
	"rabbitmq/service"
	"strconv"
	"time"
)
var (
	appConfig *WebConfig // 定义一个全局的变量，专门存放配置信息，引用某个结构体的地址
)
type WebConfig struct {
	log_level string
	log_path string
}
// 加载配置文件,并且赋值全局变量appConfig
func loadConfig(confType, filename string) (err error) {
	conf, err := config.NewConfig(confType, filename)
	if err != nil {
		fmt.Println("load config failed, err:%v", err)
		return
	}
	appConfig = &WebConfig{} // 初始化一个结构体
	// 初始话日记参数
	appConfig.log_level = conf.String("log::log_level")
	if len(appConfig.log_level) == 0{
		appConfig.log_level = "debug"
	}
	appConfig.log_path = conf.String("log::log_path")
	if len(appConfig.log_path) == 0 {
		appConfig.log_level = "./logs/logagent.log"
	}

	logs.Info(appConfig)
	logs.Debug("load conf succsss %v\n",appConfig)
	return

}
// 初始化日记
func initLogger() (err error)  {
	config := make(map[string]interface{})
	config["filename"] = appConfig.log_path
	config["level"] = convertLogLevel(appConfig.log_level)
	configStr, err := json.Marshal(config)
	if err != nil {
		fmt.Println("marshal failed, err:", err)
		return
	}
	logs.SetLogger(logs.AdapterFile, string(configStr))
	logs.Debug("this is a test, my name is %s", "Mr.chen")
	return
}
// 返回错误级别 1~8
func convertLogLevel(logLevel string) int {
	switch (logLevel) {
	case "debug":
		return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace
	}
	return  logs.LevelDebug
}
func main()  {

	err := loadConfig(cfg.ConfType,cfg.Filename)
	if err != nil {
		fmt.Println("load config file fail,err: %v\n",err)
		panic(err) // 配置信息加载失败，立刻退出，不进行下面的操作
	}
	// 初始化日记
	err =  initLogger()
	if err != nil {
		fmt.Println("load config file fail,err: %v\n",err)
		panic(err) // 配置信息加载失败，立刻退出，不进行下面的操作
	}
	// 测试rabbitmq
	go func() {
		for i:=1;i<50;i++ {
			data := mq.RegiserUser{
				Id:     i,
				Name:  "user_"+ strconv.Itoa(i),
				UserNum:rand.Intn(99999999),
			}
			pubData, _ := json.Marshal(data)
			ok := mq.Publish("testserver.trans","register",pubData)
			if !ok {
				fmt.Println("send rabbitmq faile")
			}
			time.Sleep(3*time.Second)
			fmt.Println(fmt.Sprintf("输入：%s",string(pubData)))
			logs.Info("输入%s",string(pubData))
		}
	}()

	//4. 运行业务逻辑，主要是处理Goroutine携程
	go func() { // 防止阻塞
		err = service.Run()
		if err != nil {
			fmt.Println(fmt.Sprintf("service run return, err:%v", err))
			return
		}
	}()
	// 查看效果
	time.Sleep(200*time.Second) //   如果这个不加sleep，main程序结束之后所有goroutines都会退出

}


```



### conf配置文件：

```
[log]
log_level = debug
log_path=E:\\Go\\src\\rabbitmq\\logs\\logs.log
[rabbitmq]
;rabbitmq相关配置
# AsyncTransferEnable : 是否开启文件异步转移(默认同步)
AsyncTransferEnable = true
#  RabbitURL : rabbitmq服务的入口url
RabbitURL = "amqp://guest:guest@127.0.0.1:5672/"
#  TransExchangeName : 用于transfer的交换机
TransExchangeName = "testserver.trans"
#  TransRegisterQueueName : Register转移队列名
TransRegisterQueueName = "testserver.trans.register"
# TransRegisterErrQueueName : Register转移失败后写入另一个队列的队列名
TransRegisterErrQueueName = "testserver.trans.register.err"
#  TransRegisterRoutingKey : routingkey
TransRegisterRoutingKey = "register"
```

conf/config

```
/**
* @program: Go
*
* @description:配置文件参数
*
* @author: Mr.chen
*
* @create: 2020-01-06 16:45
**/
package config

const (
	ConfType = "ini"
	Filename = "E:\\Go\\src\\rabbitmq\\conf\\go.conf.ini"
)

```

### mq文件：

#### config.go

```
/**
* @program: Go
*
* @description:定义需要操作的数据
*
* @author: Mr.chen
*
* @create: 2020-01-06 16:12
**/
package mq

type RegiserUser struct {
	Id int
	Name string
	UserNum int
}

```

consumer.go

```
/**
* @program: Go
*
* @description:
*
* @author: Mr.chen
*
* @create: 2020-01-06 17:02
**/
package mq

import "log"

var done chan bool

// StartConsume : 接收消息
func StartConsume(qName, cName string, callback func(msg []byte) bool) {
	msgs, err := channel.Consume(
		qName,
		cName,
		true,  //自动应答
		false, // 非唯一的消费者
		false, // rabbitMQ只能设置为false
		false, // noWait, false表示会阻塞直到有消息过来
		nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	done = make(chan bool)

	go func() {
		// 循环读取channel的数据
		for d := range msgs {
			processErr := callback(d.Body)
			if processErr {
				// TODO: 将任务写入错误队列，待后续处理
			}
		}
	}()

	// 接收done的信号, 没有信息过来则会一直阻塞，避免该函数退出
	<-done

	// 关闭通道
	channel.Close()
}

// StopConsume : 停止监听队列
func StopConsume() {
	done <- true
}


```

#### producer.go

```

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

```

### service文件

#### runtask.go

```
/**
* @program: Go
*
* @description:rabbitmq主要业务逻辑
*
* @author: Mr.chen
*
* @create: 2020-01-06 14:01
**/
package service

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"rabbitmq/mq"
	"sync"
	"time"
)
var waitGroup  sync.WaitGroup
var (
	webConfig  *mq.RabbitmqConfig
)

// 主要的业务逻辑
func RunTask() (err error) {
	for i := 0; i < 1; i++ {
		waitGroup.Add(1)
		go HandleTransferMq() // 处理接收数据
	}
	logs.Debug("all task process goroutine started")
	waitGroup.Wait()
	logs.Debug("wait all goroutine exited")
	return

}

func HandleTransferMq()  {
	logs.Debug("transfer queue goroutine is running")
	for  {
		err := MqTask()
		if err != nil {
			logs.Error("receive msg failed,%v",err)
			return
		}
		time.Sleep(3*time.Second) // 延时3秒
		return
	}
}


// ProcessTransfer : 处理文件转移
func ProcessTransfer(msg []byte) bool {
	logs.Debug(string(msg))
	pubData := mq.RegiserUser{}
	err := json.Unmarshal(msg, &pubData)
	if err != nil {
		logs.Error(err.Error())
		return false
	}
	fmt.Println(fmt.Sprintf("输出：%s",string(msg)))
	logs.Info(pubData)

	return true
}

func MqTask() (err error) {
	err = mq.InitMqConfig()
	if err != nil {
		logs.Info("init Mqconfig failed,%v",err.Error())
		return
	}
	webConfig = mq.GetConfig()
	// 是否开启异步转移功能，开启时才初始化rabbitMQ连接
	if !webConfig.AsyncTransferEnable {
		return
	}
	// 是否开启异步转移功能，开启时才初始化rabbitMQ连接
	if !webConfig.AsyncTransferEnable {
		logs.Info("异步转移文件功能目前被禁用，请检查相关配置")
		return
	}

	logs.Debug("文件转移服务启动中，开始监听转移任务队列...")
	mq.StartConsume(
		webConfig.TransRegisterQueueName,
		"transfer_oss",
		ProcessTransfer)
	return nil
}
```

#### service.go

```
/**
* @program: Go
*
* @description:
*
* @author: Mr.chen
*
* @create: 2020-01-06 17:05
**/
package service

func Run() (err error) {

	//起处理线程
	//err = RunProcess()
	err = RunTask()
	return
}
```

项目源码文件

<https://github.com/18318553760/rabbitmq> 