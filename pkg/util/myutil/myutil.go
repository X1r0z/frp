package myutil

import (
	"fmt"
	"github.com/fatedier/frp/pkg/msg"
	"math/rand"
	"time"
)

func GetMessage(serverIp string, remoteIp string, loginMsg *msg.Login) string {

	msg := fmt.Sprintf(`客户端信息
remote ip: %v
local ip: %v
hostname: %v
os: %v
arch: %v
username: %v
version: %v`, remoteIp, loginMsg.Ip, loginMsg.HostName, loginMsg.Os, loginMsg.Arch, loginMsg.UserName, loginMsg.Version)
	if loginMsg.SocksPort != 0 {
		msg += "\n\n"
		msg += fmt.Sprintf(`socks 代理信息
server ip: %v
port: %v`, serverIp, loginMsg.SocksPort)
		if loginMsg.SocksUser != "" && loginMsg.SocksPass != "" {
			msg += "\n"
			msg += fmt.Sprintf(`user: %v
pass: %v`, loginMsg.SocksUser, loginMsg.SocksPass)
		}
	}
	return msg
}

func RandStr(length int) string {
	dicts := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	result := make([]byte, length)

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	for i := 0; i < length; i++ {
		result[i] = dicts[r.Intn(len(dicts))]
	}

	return string(result)
}
