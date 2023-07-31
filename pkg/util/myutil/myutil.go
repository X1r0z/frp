package myutil

import (
	"fmt"
	"github.com/fatedier/frp/pkg/msg"
)

func GetMessage(remoteIp string, loginMsg *msg.Login) string {

	if remoteIp == "" {
		return fmt.Sprintf(`客户端信息
local ip: %v
hostname: %v
os: %v
arch: %v
username: %v
version: %v`, loginMsg.Ip, loginMsg.Hostname, loginMsg.Os, loginMsg.Arch, loginMsg.Username, loginMsg.Version)
	} else {
		return fmt.Sprintf(`客户端信息
remote ip: %v
local ip: %v
hostname: %v
os: %v
arch: %v
username: %v
version: %v`, remoteIp, loginMsg.Ip, loginMsg.Hostname, loginMsg.Os, loginMsg.Arch, loginMsg.Username, loginMsg.Version)
	}
}
