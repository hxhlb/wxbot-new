//go:build !windows

package wxsvc

import (
    "errors"
    "log"
    "time"
)

// 非 Windows 平台的 stub，实现接口以保证跨平台编译与基本可运行
type stubLoader struct{
    connected bool
    clientID  uint32
}

func newLoader(loaderPath string) Loader { // 参数占位，便于后续替换
    return &stubLoader{}
}

func (s *stubLoader) InitWeChatSocket(connectCb func(uint32), recvCb func(uint32, []byte), closeCb func(uint32)) bool {
    // 简单模拟一次连接回调
    go func() {
        time.Sleep(300 * time.Millisecond)
        s.clientID = 1
        s.connected = true
        if connectCb != nil {
            connectCb(s.clientID)
        }
    }()
    return true
}

func (s *stubLoader) InjectWeChat(_ string) (uint32, error) {
    s.clientID = 1
    s.connected = true
    return s.clientID, nil
}

func (s *stubLoader) SendWeChatData(clientID uint32, message string) (bool, error) {
    if !s.connected || clientID == 0 {
        return false, errors.New("service not connected")
    }
    log.Printf("[stub] 发送消息: %s", message)
    return true, nil
}

func (s *stubLoader) DestroyWeChat() error {
    if s.connected {
        log.Printf("[stub] 连接已断开")
    }
    s.connected = false
    s.clientID = 0
    return nil
}

func (s *stubLoader) UseUtf8() bool { return true }

func (s *stubLoader) InjectWeChat2(_ string, _ string) (uint32, error) {
    s.clientID = 1
    s.connected = true
    return s.clientID, nil
}

func (s *stubLoader) InjectWeChatPid(pid uint32, _ string) (uint32, error) {
    if pid == 0 {
        return 0, errors.New("invalid pid")
    }
    s.clientID = pid
    s.connected = true
    return s.clientID, nil
}

func (s *stubLoader) InjectWeChatMultiOpen(_ string, _ string) (uint32, error) {
    s.clientID = 2
    s.connected = true
    return s.clientID, nil
}

func (s *stubLoader) GetUserWeChatVersion() (string, error) {
    return "0.0.0-stub", nil
}

func (s *stubLoader) GetInstallWeixinVersion() (string, error) {
    return "", nil
}
