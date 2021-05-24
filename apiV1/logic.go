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
	user := User {
		Name: string(msg),
		Uid: uid,
		CreateTime: time.Now().Unix(),
		Conn: ws,
	}
	// 将该用户添加至 AllUser
	AllUser = append(AllUser, user)

	// 初始化
	_ = user.Conn.WriteJSON(bson.M{"data": user, "type": "init"})
	for _, u := range AllUser {
		_ = u.Conn.WriteJSON(bson.M{"data": user, "type": "add"})
	}
	// 等待该用户发送消息
	go WaitForSend(user)
}

// WaitForSend 等待该用户发送消息
func WaitForSend(user User) {
	for {
		_, msg, err := user.Conn.ReadMessage()
		if err != nil {
			break
		}
		myMsg := Msg{
			SendUser: &user,
			ReceiveUser: nil,
			Time: time.Now().Unix(),
			Data: string(msg),
		}
		Send(myMsg)
	}
	// 函数退出时执行
	defer Logout(user)
}

// WaitForReceive 用于接收消息
func WaitForReceive(user User) {
	for {
		_, msg, err := user.Conn.ReadMessage()
		if err != nil {
			break
		}
		myMsg := Msg{
			SendUser: &user,
			Time: time.Now().Unix(),
			Data: string(msg),
		}
		Send(myMsg)
	}
	defer Logout(user)
}

// Send 发送消息
func Send(msg Msg) {
	// 若msg 的接收方为空，则为广播消息
	if msg.ReceiveUser == nil {
		for _, user := range AllUser {
			err := user.Conn.WriteJSON(bson.M{"data": msg, "type":"receive"})
			if err != nil {
				log.Println("用户退出失败，", err)
			}
		}
		// 不为空表示私聊
	} else {

	}
}

// Logout 退出登录
func Logout(user User) {
	for _, u := range AllUser {
		err := u.Conn.WriteJSON(bson.M{"data": user, "type":"logout"})
		if err != nil {
			log.Println("用户退出失败，", err)
		}
	}

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
}
