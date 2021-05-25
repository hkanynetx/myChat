package apiV1

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"time"
)

// 协议升级
var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Connect 用户连接
// 创建新用户并添加至 AllUser
func Connect(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("升级协议失败，", err)
	}
	// 读取用户名
	_, msg, err := ws.ReadMessage()
	// 初始化用户
	uid, err := uuid.NewUUID()
	user := User{
		Name:       string(msg),
		Uid:        uid,
		CreateTime: time.Now().Unix(),
		Conn:       ws,
	}
	// 将该用户添加至 AllUser
	AllUser = append(AllUser, user)

	//初始化
	Broadcast(bson.M{"data": user, "type": "add"})

	// 等待该用户发送消息
	go WaitForSend(user)
}

// WaitForSend 等待该用户发送消息
func WaitForSend(user User) {
	var msg bson.M
	var myMsg Msg
	for {
		err := user.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}
		if msg["uid"] == "" {
			myMsg = Msg{
				SendUser:    &user,
				ReceiveUser: nil,
				Time:        time.Now().Unix(),
				Data:        msg["msg"].(string),
			}
		} else {
			for _, u := range AllUser {
				if u.Uid.String() == msg["uid"].(string) {
					myMsg = Msg{
						SendUser:    &user,
						ReceiveUser: &u,
						Time:        time.Now().Unix(),
						Data:        msg["msg"].(string),
					}
					break
				}
			}
		}
		Send(myMsg)
	}
	// 函数退出时执行
	defer Logout(user)
}

// Send 发送消息
func Send(msg Msg) {
	// 接收者为空，广播消息
	if msg.ReceiveUser == nil {
		Broadcast(bson.M{"data": msg, "type": "receive"})

		// 私聊消息
	} else {
		// 发送消息
		err := msg.ReceiveUser.Conn.WriteJSON(bson.M{"data": msg, "type": "private"})
		if err != nil {
			log.Println("写入失败，", err)
		}
		// 发送消息
		err = msg.SendUser.Conn.WriteJSON(bson.M{"data": msg, "type": "private"})
		if err != nil {
			log.Println("写入失败，", err)
		}
	}
}

// Logout 退出登录
func Logout(user User) {
	// 释放内存
	for i, each := range AllUser {
		if user.Uid == each.Uid {
			// 关闭连接
			err := user.Conn.Close()
			if err != nil {
				log.Println("用户退出失败，", err)
			}
			// 移除该用户
			AllUser = append(AllUser[:i], AllUser[i+1:]...)
			break
		}
	}
	// 通知所有在线用户
	Broadcast(bson.M{"data": user, "type": "logout"})
}

// Broadcast 全体广播
func Broadcast(msg bson.M) {
	for _, user := range AllUser {
		err := user.Conn.WriteJSON(msg)
		if err != nil {
			log.Println("写入失败，", err)
		}
	}
}

// GetAllMembers 当前聊天室总人数
func GetAllMembers(c *gin.Context) {
	length := len(AllUser)
	c.JSON(200, gin.H{
		"length": length,
		"status": true,
	})
}
