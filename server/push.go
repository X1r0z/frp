package server

import (
	"bytes"
	"fmt"
	"net/http"
)

func FeishuBotPush(token string, message string) {
	url := "https://open.feishu.cn/open-apis/bot/v2/hook/" + token

	data := fmt.Sprintf(`{
    "msg_type": "post",
    "content": {
        "post": {
            "zh_cn": {
                "title": "frp 客户端上线通知",
                "content": [
                    [{
                        "tag": "text",
                        "text": %#v
                    }]
                ]
            }
        }
    }
}`, message)
	http.Post(url, "application/json", bytes.NewBufferString(data))
}
