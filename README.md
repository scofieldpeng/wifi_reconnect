# wifi_reconnect

## 由来

买了一个某米路由器后想起家里有一闲置很久的上网本,于是便用来当做了服务器.因为路由器是挂在墙上,于是上网本用wifi来进行交换,悲伤地发现这厮在掉线或者路由器重启后不能自动重连,于是用go写了这个小东西专门用来判断网卡是否连接.

## 原理

启动后间隔ping下外网的一个域名,通过能够ping通来查看是否连接上外网,如果没有连上,则调用系统命令`ifup`命令重新启动wifi

## 安装

```golang
go get http://github.com/scofieldpeng/wifi_reconnect
```

## 使用

涉及到网卡操作,必须在root权限下运行:)

## 配置

```
-h
    查看帮助信息
-c int
    ping的次数 (default 2)
-d string
    重启的网卡名称 (default "eth0")
-debug
    是否开启默认模式
-i int
    每次ping的时间间隔 (default 1)
-o string
    log文件存放路径
-u string
    ping的url (default "http://baidu.com")
```

## License

The MIT License(http://opensource.org/licenses/MIT) , 请随意修改

## 贡献

如果你有好的意见或建议，请自行fork修改或者提issue或pull request