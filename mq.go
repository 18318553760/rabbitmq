/**
* @program: Go
*
* @description:
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

