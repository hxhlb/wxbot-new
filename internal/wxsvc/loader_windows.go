//go:build windows

package wxsvc

import (
    "errors"
    "log"
    "runtime"
    "syscall"
    "unsafe"
)

// 与 Python 版一致的偏移地址（通常为 32 位 DLL）
const (
    offInitWeChatSocket      = 0xB080
    offGetUserWeChatVersion  = 0xCB80
    offInjectWeChat          = 0xCC10
    offSendWeChatData        = 0xAF90
    offDestroyWeChat         = 0xC540
    offUseUtf8               = 0xC680
    offInjectWeChat2         = 0xCC30
    offInjectWeChatPid       = 0xB750
    offInjectWeChatMultiOpen = 0xC780
    offGetInstallWeixinVersion = 0x0 // Python 版本未提供有效偏移
)

type winLoader struct {
    module syscall.Handle
    base   uintptr

    // 计算后的函数地址
    fnInitWeChatSocket uintptr
    fnGetVersion       uintptr
    fnInjectWeChat     uintptr
    fnSendWeChatData   uintptr
    fnDestroyWeChat    uintptr
    fnUseUtf8          uintptr
    fnInjectWeChat2    uintptr
    fnInjectWeChatPid  uintptr
    fnInjectWeChatMO   uintptr
    fnGetInstallVer    uintptr

    // 回调与透传
    connectCb func(clientID uint32)
    recvCb    func(clientID uint32, data []byte)
    closeCb   func(clientID uint32)

    // Windows 回调指针
    cConnect uintptr
    cRecv    uintptr
    cClose   uintptr

    connected bool
    clientID  uint32
}

func newLoader(loaderPath string) Loader {
    w := &winLoader{}

    if runtime.GOARCH != "386" { // Python 版同样限制 32 位
        log.Printf("检测到 %s 架构，可能与 32 位 DLL 不兼容", runtime.GOARCH)
    }

    // 加载 DLL 并计算函数地址
    h, err := syscall.LoadLibrary(loaderPath)
    if err != nil {
        log.Printf("加载 DLL 失败: %v", err)
        return w // 保持对象可用，但后续调用会失败
    }
    w.module = h
    w.base = uintptr(h)
    w.fnInitWeChatSocket = w.base + offInitWeChatSocket
    w.fnGetVersion = w.base + offGetUserWeChatVersion
    w.fnInjectWeChat = w.base + offInjectWeChat
    w.fnSendWeChatData = w.base + offSendWeChatData
    w.fnDestroyWeChat = w.base + offDestroyWeChat
    w.fnUseUtf8 = w.base + offUseUtf8
    w.fnInjectWeChat2 = w.base + offInjectWeChat2
    w.fnInjectWeChatPid = w.base + offInjectWeChatPid
    w.fnInjectWeChatMO = w.base + offInjectWeChatMultiOpen
    if offGetInstallWeixinVersion != 0 {
        w.fnGetInstallVer = w.base + offGetInstallWeixinVersion
    }

    // 尝试设置 UTF-8
    w.callUseUtf8()
    return w
}

func (w *winLoader) InitWeChatSocket(connectCb func(uint32), recvCb func(uint32, []byte), closeCb func(uint32)) bool {
    w.connectCb = connectCb
    w.recvCb = recvCb
    w.closeCb = closeCb

    // 准备 Windows 回调
    w.cConnect = syscall.NewCallback(func(clientID uintptr) uintptr {
        if w.connectCb != nil {
            w.connectCb(uint32(clientID))
        }
        return 0
    })
    w.cRecv = syscall.NewCallback(func(clientID, dataPtr, length uintptr) uintptr {
        if w.recvCb != nil && dataPtr != 0 && length > 0 {
            // 将 dataPtr/length 转为字节切片，并复制一份
            var b = unsafe.Slice((*byte)(unsafe.Pointer(dataPtr)), int(length))
            dup := make([]byte, len(b))
            copy(dup, b)
            w.recvCb(uint32(clientID), dup)
        }
        return 0
    })
    w.cClose = syscall.NewCallback(func(clientID uintptr) uintptr {
        if w.closeCb != nil {
            w.closeCb(uint32(clientID))
        }
        return 0
    })

    if w.fnInitWeChatSocket == 0 {
        log.Printf("InitWeChatSocket 地址无效")
        return false
    }

    r1, _, _ := syscall.Syscall(w.fnInitWeChatSocket, 3, w.cConnect, w.cRecv, w.cClose)
    return r1 != 0
}

func (w *winLoader) InjectWeChat(dllPath string) (uint32, error) {
    if w.fnInjectWeChat == 0 {
        return 0, errors.New("InjectWeChat 地址无效")
    }

    // 以 UTF-8 传入，末尾 0 结尾
    b := append([]byte(dllPath), 0)
    var p uintptr
    if len(b) > 0 {
        p = uintptr(unsafe.Pointer(&b[0]))
    }
    r1, _, e1 := syscall.Syscall(w.fnInjectWeChat, 1, p, 0, 0)
    if r1 == 0 && e1 != 0 {
        log.Printf("InjectWeChat 调用错误: %v", e1)
    }
    w.clientID = uint32(r1)
    w.connected = w.clientID != 0
    if !w.connected {
        return 0, errors.New("InjectWeChat 返回 0")
    }
    return w.clientID, nil
}

func (w *winLoader) SendWeChatData(clientID uint32, message string) (bool, error) {
    if w.fnSendWeChatData == 0 {
        return false, errors.New("SendWeChatData 地址无效")
    }
    if clientID == 0 {
        return false, errors.New("无效的 clientID")
    }
    b := append([]byte(message), 0)
    var p uintptr
    if len(b) > 0 {
        p = uintptr(unsafe.Pointer(&b[0]))
    }
    r1, _, e1 := syscall.Syscall(w.fnSendWeChatData, 2, uintptr(clientID), p, 0)
    if r1 == 0 && e1 != 0 {
        log.Printf("SendWeChatData 调用错误: %v", e1)
    }
    return r1 != 0, nil
}

func (w *winLoader) DestroyWeChat() error {
    if w.fnDestroyWeChat == 0 {
        return nil
    }
    r1, _, e1 := syscall.Syscall(w.fnDestroyWeChat, 0, 0, 0, 0)
    if r1 == 0 && e1 != 0 {
        log.Printf("DestroyWeChat 调用错误: %v", e1)
    }
    if w.module != 0 {
        _ = syscall.FreeLibrary(w.module)
        w.module = 0
    }
    w.connected = false
    w.clientID = 0
    return nil
}

func (w *winLoader) callUseUtf8() {
    if w.fnUseUtf8 != 0 {
        r1, _, e1 := syscall.Syscall(w.fnUseUtf8, 0, 0, 0, 0)
        if r1 == 0 && e1 != 0 {
            log.Printf("UseUtf8 调用错误: %v", e1)
        }
    }
}

// 公开方法：与 Loader 接口一致
func (w *winLoader) UseUtf8() bool {
    if w.fnUseUtf8 == 0 { return false }
    r1, _, _ := syscall.Syscall(w.fnUseUtf8, 0, 0, 0, 0)
    return r1 != 0
}

func (w *winLoader) InjectWeChat2(dllPath, exePath string) (uint32, error) {
    if w.fnInjectWeChat2 == 0 {
        return 0, errors.New("InjectWeChat2 地址无效")
    }
    db := append([]byte(dllPath), 0)
    eb := append([]byte(exePath), 0)
    var dp, ep uintptr
    if len(db) > 0 { dp = uintptr(unsafe.Pointer(&db[0])) }
    if len(eb) > 0 { ep = uintptr(unsafe.Pointer(&eb[0])) }
    r1, _, e1 := syscall.Syscall(w.fnInjectWeChat2, 2, dp, ep, 0)
    if r1 == 0 && e1 != 0 { log.Printf("InjectWeChat2 调用错误: %v", e1) }
    return uint32(r1), nil
}

func (w *winLoader) InjectWeChatPid(pid uint32, dllPath string) (uint32, error) {
    if w.fnInjectWeChatPid == 0 {
        return 0, errors.New("InjectWeChatPid 地址无效")
    }
    db := append([]byte(dllPath), 0)
    var dp uintptr
    if len(db) > 0 { dp = uintptr(unsafe.Pointer(&db[0])) }
    r1, _, e1 := syscall.Syscall(w.fnInjectWeChatPid, 2, uintptr(pid), dp, 0)
    if r1 == 0 && e1 != 0 { log.Printf("InjectWeChatPid 调用错误: %v", e1) }
    return uint32(r1), nil
}

func (w *winLoader) InjectWeChatMultiOpen(dllPath, exePath string) (uint32, error) {
    if w.fnInjectWeChatMO == 0 {
        return 0, errors.New("InjectWeChatMultiOpen 地址无效")
    }
    db := append([]byte(dllPath), 0)
    eb := append([]byte(exePath), 0)
    var dp, ep uintptr
    if len(db) > 0 { dp = uintptr(unsafe.Pointer(&db[0])) }
    if len(eb) > 0 { ep = uintptr(unsafe.Pointer(&eb[0])) }
    r1, _, e1 := syscall.Syscall(w.fnInjectWeChatMO, 2, dp, ep, 0)
    if r1 == 0 && e1 != 0 { log.Printf("InjectWeChatMultiOpen 调用错误: %v", e1) }
    return uint32(r1), nil
}

func (w *winLoader) GetUserWeChatVersion() (string, error) {
    if w.fnGetVersion == 0 {
        return "", errors.New("GetUserWeChatVersion 地址无效")
    }
    buf := make([]byte, 64)
    p := uintptr(unsafe.Pointer(&buf[0]))
    r1, _, e1 := syscall.Syscall(w.fnGetVersion, 1, p, 0, 0)
    if r1 == 0 && e1 != 0 { log.Printf("GetUserWeChatVersion 调用错误: %v", e1) }
    if r1 == 0 { return "", nil }
    // NUL 终止
    n := 0
    for ; n < len(buf); n++ { if buf[n] == 0 { break } }
    return string(buf[:n]), nil
}

func (w *winLoader) GetInstallWeixinVersion() (string, error) {
    if w.fnGetInstallVer == 0 {
        return "", nil // 与 Python 保持一致：未实现则返回空字符串
    }
    buf := make([]byte, 64)
    p := uintptr(unsafe.Pointer(&buf[0]))
    r1, _, _ := syscall.Syscall(w.fnGetInstallVer, 1, p, 0, 0)
    if r1 == 0 { return "", nil }
    n := 0
    for ; n < len(buf); n++ { if buf[n] == 0 { break } }
    return string(buf[:n]), nil
}
