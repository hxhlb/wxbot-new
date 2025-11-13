package wxsvc

// Loader 抽象出微信底层能力，便于跨平台与替换实现
type Loader interface {
    // InitWeChatSocket 注册回调（可选）
    InitWeChatSocket(connectCb func(clientID uint32), recvCb func(clientID uint32, data []byte), closeCb func(clientID uint32)) bool
    // InjectWeChat 注入并返回 clientID
    InjectWeChat(dllPath string) (uint32, error)
    // SendWeChatData 发送原始 JSON 字符串
    SendWeChatData(clientID uint32, message string) (bool, error)
    // DestroyWeChat 清理资源
    DestroyWeChat() error

    // 额外接口：与 Python 版保持一致
    UseUtf8() bool
    InjectWeChat2(dllPath, exePath string) (uint32, error)
    InjectWeChatPid(pid uint32, dllPath string) (uint32, error)
    InjectWeChatMultiOpen(dllPath, exePath string) (uint32, error)
    GetUserWeChatVersion() (string, error)
    GetInstallWeixinVersion() (string, error)
}
