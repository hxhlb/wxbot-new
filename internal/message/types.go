package message

// MessageType 消息类型常量
type MessageType int

const (
	MTDebugLog         MessageType = 11024 // 调试日志
	MTUserLogin        MessageType = 11025 // 用户登录
	MTUserLogout       MessageType = 11026 // 用户登出
	MTFriendList       MessageType = 11030 // 好友列表
	MTSendText         MessageType = 11036 // 发送文本消息
	MTSendAtText       MessageType = 11037 // 发送@消息
	MTSendCard         MessageType = 11038 // 发送卡片
	MTSendURL          MessageType = 11039 // 发送链接
	MTSendImage        MessageType = 11040 // 发送图片
	MTSendFile         MessageType = 11041 // 发送文件
	MTSendVideo        MessageType = 11042 // 发送视频
	MTSendGif          MessageType = 11043 // 发送GIF
	MTChatMessage      MessageType = 11046 // 聊天消息
	MTCurrentLoginInfo MessageType = 11028 // 当前登录信息
)

// Message 消息结构
type Message struct {
	Type MessageType            `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// SendTextData 发送文本消息数据
type SendTextData struct {
	ToWxid  string `json:"to_wxid"`
	Content string `json:"content"`
}

// SendAtTextData 发送@消息数据
type SendAtTextData struct {
	ToWxid  string   `json:"to_wxid"`
	Content string   `json:"content"`
	AtList  []string `json:"at_list"`
}

// SendCardData 发送卡片数据
type SendCardData struct {
	ToWxid   string `json:"to_wxid"`
	CardWxid string `json:"card_wxid"`
}

// SendURLData 发送链接数据
type SendURLData struct {
	ToWxid   string `json:"to_wxid"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	URL      string `json:"url"`
	ImageURL string `json:"image_url"`
}

// SendFileData 发送文件数据
type SendFileData struct {
	ToWxid string `json:"to_wxid"`
	File   string `json:"file"`
}
