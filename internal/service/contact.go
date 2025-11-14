package service

import (
	"encoding/json"
	"fmt"
	"log"

	"wxbot-new/internal/message"
)

// HelperGetFriendList 获取好友列表
func (s *WeChatService) HelperGetFriendList() error {
	msg := message.Message{
		Type: message.MTFriendList,
		Data: make(map[string]interface{}),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	log.Printf("获取好友列表请求: %s", string(data))
	return s.SendMessage(string(data))
}
