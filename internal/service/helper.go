package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

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

// HelperGetCurrentLoginInfo 获取当前登录信息(同步方式,带超时)
func (s *WeChatService) HelperGetCurrentLoginInfo() (*message.CurrentLoginInfoData, error) {
	// 构造消息(无需 trace)
	msg := message.Message{
		Type: message.MTCurrentLoginInfo,
		Data: map[string]interface{}{},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("序列化消息失败: %v", err)
	}

	// 使用消息类型和客户端ID注册等待响应
	responseChan := s.responseManager.RegisterRequest(int(message.MTCurrentLoginInfo), s.clientID, 10*time.Second)

	// 发送请求
	log.Printf("获取当前登录信息请求 [msgType=%d, clientID=%d]: %s", message.MTCurrentLoginInfo, s.clientID, string(data))
	if err := s.SendMessage(string(data)); err != nil {
		s.responseManager.CancelRequest(int(message.MTCurrentLoginInfo), s.clientID) // 发送失败时清理注册
		return nil, fmt.Errorf("发送消息失败: %v", err)
	}

	// 等待响应
	respData, err := s.responseManager.WaitForResponse(responseChan, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("等待响应超时: %v", err)
	}

	// 解析响应数据
	loginInfo := &message.CurrentLoginInfoData{}
	if account, ok := respData["account"].(string); ok {
		loginInfo.Account = account
	}
	if avatar, ok := respData["avatar"].(string); ok {
		loginInfo.Avatar = avatar
	}
	if nickname, ok := respData["nickname"].(string); ok {
		loginInfo.Nickname = nickname
	}
	if wxid, ok := respData["wxid"].(string); ok {
		loginInfo.Wxid = wxid
	}

	log.Printf("获取登录信息成功: %+v", loginInfo)
	return loginInfo, nil
}

// HelperSendText 发送普通文本消息
func (s *WeChatService) HelperSendText(toWxid, content string) error {
	msg := message.Message{
		Type: message.MTSendText,
		Data: map[string]interface{}{
			"to_wxid": toWxid,
			"content": content,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	log.Printf("发送文本消息: %s", string(data))
	return s.SendMessage(string(data))
}

// HelperSendAtText 发送@消息
func (s *WeChatService) HelperSendAtText(toWxid, content string, atList []string) error {
	msg := message.Message{
		Type: message.MTSendAtText,
		Data: map[string]interface{}{
			"to_wxid": toWxid,
			"content": content,
			"at_list": atList,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	log.Printf("发送@消息: %s", string(data))
	return s.SendMessage(string(data))
}

// HelperSendCard 发送卡片消息
func (s *WeChatService) HelperSendCard(toWxid, cardWxid string) error {
	msg := message.Message{
		Type: message.MTSendCard,
		Data: map[string]interface{}{
			"to_wxid":   toWxid,
			"card_wxid": cardWxid,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	log.Printf("发送卡片消息: %s", string(data))
	return s.SendMessage(string(data))
}

// HelperSendURL 发送链接消息
func (s *WeChatService) HelperSendURL(toWxid, title, desc, url, imageURL string) error {
	msg := message.Message{
		Type: message.MTSendURL,
		Data: map[string]interface{}{
			"to_wxid":   toWxid,
			"title":     title,
			"desc":      desc,
			"url":       url,
			"image_url": imageURL,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	log.Printf("发送链接消息: %s", string(data))
	return s.SendMessage(string(data))
}

// HelperSendImage 发送图片消息
func (s *WeChatService) HelperSendImage(toWxid, filePath string) error {
	msg := message.Message{
		Type: message.MTSendImage,
		Data: map[string]interface{}{
			"to_wxid": toWxid,
			"file":    filePath,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	log.Printf("发送图片消息: %s", string(data))
	return s.SendMessage(string(data))
}

// HelperSendFile 发送文件消息
func (s *WeChatService) HelperSendFile(toWxid, filePath string) error {
	msg := message.Message{
		Type: message.MTSendFile,
		Data: map[string]interface{}{
			"to_wxid": toWxid,
			"file":    filePath,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	log.Printf("发送文件消息: %s", string(data))
	return s.SendMessage(string(data))
}

// HelperSendVideo 发送视频消息
func (s *WeChatService) HelperSendVideo(toWxid, filePath string) error {
	msg := message.Message{
		Type: message.MTSendVideo,
		Data: map[string]interface{}{
			"to_wxid": toWxid,
			"file":    filePath,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	log.Printf("发送视频消息: %s", string(data))
	return s.SendMessage(string(data))
}

// HelperSendGif 发送GIF消息
func (s *WeChatService) HelperSendGif(toWxid, filePath string) error {
	msg := message.Message{
		Type: message.MTSendGif,
		Data: map[string]interface{}{
			"to_wxid": toWxid,
			"file":    filePath,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	log.Printf("发送GIF消息: %s", string(data))
	return s.SendMessage(string(data))
}

// HelperRefreshQRCode 刷新二维码(同步方式,带超时)
func (s *WeChatService) HelperRefreshQRCode() (*message.RefreshQRCodeData, error) {
	// 构造消息
	msg := message.Message{
		Type: message.MTRefreshQRCode,
		Data: map[string]interface{}{},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("序列化消息失败: %v", err)
	}

	// 使用消息类型和客户端ID注册等待响应
	responseChan := s.responseManager.RegisterRequest(int(message.MTRefreshQRCode), s.clientID, 10*time.Second)

	// 发送请求
	log.Printf("刷新二维码请求 [msgType=%d, clientID=%d]: %s", message.MTRefreshQRCode, s.clientID, string(data))
	if err := s.SendMessage(string(data)); err != nil {
		s.responseManager.CancelRequest(int(message.MTRefreshQRCode), s.clientID) // 发送失败时清理注册
		return nil, fmt.Errorf("发送消息失败: %v", err)
	}

	// 等待响应
	respData, err := s.responseManager.WaitForResponse(responseChan, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("等待响应超时: %v", err)
	}

	// 解析响应数据
	qrData := &message.RefreshQRCodeData{}
	if file, ok := respData["file"].(string); ok {
		qrData.File = file
	}
	if qrcode, ok := respData["qrcode"].(string); ok {
		qrData.QRCode = qrcode
	}
	if pid, ok := respData["pid"].(float64); ok {
		qrData.PID = int(pid)
	}

	log.Printf("刷新二维码成功: %+v", qrData)
	return qrData, nil
}
