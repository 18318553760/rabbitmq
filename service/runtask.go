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