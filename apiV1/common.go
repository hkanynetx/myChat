package apiV1

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// 这里定义一些结构体

// AllUser 存放所有用户
var AllUser []User

type User struct {
	Name       string          `json:"name"`        //用户名
	Uid        uuid.UUID       `json:"uid"`         //随机uid
	CreateTime int64           `json:"create_time"` //加入时间
	Conn       *websocket.Conn //连接
}

type Msg struct {
	SendUser    *User  `json:"send_user"`    //发送该消息的用户
	ReceiveUser *User  `json:"receive_user"` //接受该消息的用户 为nil表示全体广播
	Time        int64  `json:"time"`         //发送时间
	Data        string `json:"data"`         //消息内容
}
