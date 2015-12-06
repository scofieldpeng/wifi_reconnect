// 该文件用来
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"strconv"
)

// 配置结构体
type Config struct {
	PingUrl      string // ping的url
	PingTryCount int    // ping的次数
	SleepMinute  int    // ping的间隔
	LogPath      string // 日志路径
	Debug        bool   // 是否是debug模式
	DeviceName   string // 设备名称
}

const (
	DEFAULT_PING_URL        = "baidu.com" // 默认ping的url
	DEFAULT_PING_TRY_COUNT  = 2                  // 默认ping的次数
	DEFAULT_PING_INTERVAL   = 1                  // ping的时间间隔
	DEFAULT_PING_DEVICENAME = "eth0"             // 默认网卡名称
)

var (
	config Config // 配置信息
)

func init() {
	pingUrl := flag.String("u", DEFAULT_PING_URL, "ping的url")
	pingCount := flag.Int("c", DEFAULT_PING_TRY_COUNT, "ping的次数")
	pingSleepMinute := flag.Int("i", 1, "每次ping的时间间隔")
	debug := flag.Bool("debug", false, "是否开启默认模式")
	logPath := flag.String("o", "", "log文件存放路径")
	deviceName := flag.String("d", DEFAULT_PING_DEVICENAME, "重启的网卡名称")

	if len(os.Args)>1 && os.Args[1] == "-h"{
		fmt.Println("该脚本能够监控某个网卡是否连上外网, 如果没有则自动连接, 源码见http://github.com/scofieldpeng/wifi_reconnect. 以下是配置参数:")
		flag.PrintDefaults()
		os.Exit(0)
	}
	flag.Parse()

	// 检查当前是否运行于root环境下
	if os.Getuid() != 0 {
		log.Println("程序需要运行于root环境下")
		os.Exit(1)
	}

	if !support() {
		os.Exit(1)
	}

	if strings.Index(*pingUrl, "http") != -1 {
		log.Println("ping的url不能为http://或者https://开头")
		os.Exit(1)
	}
	if *pingCount < 1 {
		*pingCount = 2
	}
	if *pingSleepMinute < 1 {
		*pingSleepMinute = 1
	}
	config.PingUrl = *pingUrl
	config.PingTryCount = *pingCount
	config.SleepMinute = *pingSleepMinute
	config.Debug = *debug
	config.LogPath = *logPath
	config.DeviceName = *deviceName

	if err := initLog(*debug, *logPath); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// 检查网卡设备是否有效
	if !isDeviceValid() {
		log.Println("网卡设备",config.DeviceName,"无效")
		os.Exit(1)
	}
}

// initLog 初始化log信息
// 使用官方log包,当为debug模式时所有log信息输出到标准输出设备,否则输出到设置的log文件中. 如果初始化失败,返回error
func initLog(debug bool, logPath ...string) error {
	logWriter := os.Stdout

	if !debug {
		if len(logPath) == 0 || logPath[0] == "" {
			absolutePath, err := filepath.Abs(os.Args[0])
			if err != nil {
				log.Println("获取应用绝对路径失败")
				os.Exit(1)
			}
			appDir := filepath.Dir(absolutePath)
			config.LogPath = appDir + string(os.PathSeparator) + "wifi_reconnect.log"
			if err := os.MkdirAll(appDir, os.FileMode(0755)); err != nil {
				errors.New("建立log文件夹失败!error:" + err.Error())
			}
		}

		if tmpLogWriter, err := os.OpenFile(config.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0655)); err != nil {
			errors.New("初始化log文件失败!error:" + err.Error())
		} else {
			logWriter = tmpLogWriter
		}
	}

	log.SetPrefix("[wifi_reconnect]")
	log.SetFlags(log.Ldate | log.Ltime)
	log.SetOutput(logWriter)

	return nil
}

// support 检测系统是否支持相关命令
func support() bool {
	cmdList := make([]string,4)
	cmdList[0] = "ping"
	cmdList[1] = "ifconfig"
	cmdList[2] = "ifup"
	cmdList[3] = "ifdown"

	for _,cmd := range cmdList {
		if _,err := exec.LookPath(cmd);err != nil{
			log.Println("系统没有找到",cmd,"命令")
			return false
		}
	}

	return true
}

// isDeviceValid 检测网卡设备是否可用
func isDeviceValid() bool {
	ifconfigCmd := exec.Command("ifconfig", config.DeviceName)
	if _, err := ifconfigCmd.Output(); err != nil {
		return false
	}
	return true
}

// isConnect 检测机器是否是否连接,返回bool值
func isConnect() bool {
	pingCmd := exec.Command("ping", "-c", strconv.Itoa(config.PingTryCount), config.PingUrl)
	if pingRes, err := pingCmd.CombinedOutput(); err != nil {
		log.Println(string(pingRes))
		return false
	}

	return true
}

// restartWifi 重启网卡设备
func restartDevice() bool {
	shutdownCmd := exec.Command("ifdown", config.DeviceName)
	if shutdownRes, err := shutdownCmd.CombinedOutput(); err != nil {
		log.Println(string(shutdownRes))
		return false
	}

	shutupCmd := exec.Command("ifup", config.DeviceName)
	if shutupRes, err := shutupCmd.CombinedOutput(); err != nil {
		log.Println(string(shutupRes))
		return false
	}

	return true
}

func main() {
	fmt.Println("running")

	for {
		if !isConnect() {
			log.Println("网卡",config.DeviceName,"断开")
			if !restartDevice() {
				log.Println("重连",config.DeviceName,"失败")
			} else {
				log.Println("重连",config.DeviceName,"成功")
			}
		}
		time.Sleep(time.Minute * time.Duration(config.SleepMinute))
	}
}
