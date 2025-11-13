package wxsvc

// 这些方法是对底层 Loader 的简单透传，方便外部按需调用

func (s *Service) UseUtf8() bool {
    return s.loader.UseUtf8()
}

func (s *Service) InjectWeChat2(dllPath, exePath string) (uint32, error) {
    return s.loader.InjectWeChat2(dllPath, exePath)
}

func (s *Service) InjectWeChatPid(pid uint32, dllPath string) (uint32, error) {
    return s.loader.InjectWeChatPid(pid, dllPath)
}

func (s *Service) InjectWeChatMultiOpen(dllPath, exePath string) (uint32, error) {
    return s.loader.InjectWeChatMultiOpen(dllPath, exePath)
}

func (s *Service) GetUserWeChatVersion() (string, error) {
    return s.loader.GetUserWeChatVersion()
}

func (s *Service) GetInstallWeixinVersion() (string, error) {
    return s.loader.GetInstallWeixinVersion()
}

