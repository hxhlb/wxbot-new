package service

import (
	"encoding/json"
	"fmt"
	"log"

	"wxbot-new/internal/message"
)

// SendCDNText 发送普通文本消息
func (s *WeChatService) SendCDNText(wxid, content string) error {
	msg := message.Message{
		Type: message.MTSendCDNText,
		Data: map[string]interface{}{
			"to_wxid": wxid,
			"content": content,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	log.Printf("发送文本消息(CDN): %s", string(data))
	return s.SendMessage(string(data))
}

// SendAtCDNText 发送群@消息
func (s *WeChatService) SendAtCDNText(toWxid, content string, atList []string, isAtAll bool) error {
	msgData := map[string]interface{}{
		"to_wxid": toWxid,
		"content": content,
	}

	if isAtAll {
		msgData["at_all"] = 1
	} else {
		msgData["at_list"] = atList
	}

	msg := message.Message{
		Type: message.MTSendAtCDNText,
		Data: msgData,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败(CDN): %v", err)
	}

	log.Printf("发送@消息(CDN): %s", string(data))
	return s.SendMessage(string(data))
}
