package loader

import (
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// InjectionMethod 注入方式
type InjectionMethod int

const (
	// MethodClassic 经典LoadLibrary注入 (兼容模式)
	MethodClassic InjectionMethod = iota
	// MethodManualMap 手动映射注入 (隐蔽模式)
	MethodManualMap
)

// Injector 增强型注入器
type Injector struct {
	loaderPath string
	dllPath    string
	method     InjectionMethod
	loader     *NoveLoader
}

// NewInjector 创建注入器
func NewInjector(loaderPath, dllPath string, method InjectionMethod) (*Injector, error) {
	return &Injector{
		loaderPath: loaderPath,
		dllPath:    dllPath,
		method:     method,
	}, nil
}

// Initialize 初始化
func (inj *Injector) Initialize() error {
	// 创建NoveLoader用于回调管理
	loader, err := NewNoveLoader(inj.loaderPath)
	if err != nil {
		return fmt.Errorf("创建NoveLoader失败: %v", err)
	}
	inj.loader = loader
	return nil
}

// InjectWeChat 注入微信
func (inj *Injector) InjectWeChat() (uint32, error) {
	switch inj.method {
	case MethodManualMap:
		return inj.injectManualMap()
	case MethodClassic:
		return inj.injectClassic()
	default:
		return 0, fmt.Errorf("未知的注入方式: %d", inj.method)
	}
}

// injectManualMap 使用Manual Mapping注入
func (inj *Injector) injectManualMap() (uint32, error) {
	log.Println("[ManualMap] 开始手动映射注入...")

	// 1. 读取DLL文件
	var dllData []byte
	var err error

	// 优先从嵌入资源读取
	dllData, err = GetEmbeddedDLL("NoveHelper.dll")
	if err != nil {
		// 如果嵌入资源不存在,从磁盘读取
		log.Printf("[ManualMap] 嵌入资源不可用,从磁盘读取: %s", inj.dllPath)
		dllData, err = os.ReadFile(inj.dllPath)
		if err != nil {
			return 0, fmt.Errorf("读取DLL失败: %v", err)
		}
	} else {
		log.Println("[ManualMap] 使用嵌入的DLL资源")
	}

	// 2. 获取微信进程PID
	wechatPID, err := inj.findWeChatProcess()
	if err != nil {
		return 0, fmt.Errorf("查找微信进程失败: %v", err)
	}

	log.Printf("[ManualMap] 找到微信进程: PID=%d", wechatPID)

	// 3. 延迟注入 (5-10秒随机,增加隐蔽性)
	delay := 5 + (time.Now().UnixNano() % 5)
	log.Printf("[ManualMap] 延迟 %d 秒后执行注入...", delay)
	time.Sleep(time.Duration(delay) * time.Second)

	// 4. 创建Manual Mapper
	mapper, err := NewManualMapper(wechatPID, dllData)
	if err != nil {
		return 0, fmt.Errorf("创建ManualMapper失败: %v", err)
	}

	// 5. 执行注入
	remoteBase, err := mapper.Inject()
	if err != nil {
		return 0, fmt.Errorf("Manual Mapping注入失败: %v", err)
	}

	log.Printf("[ManualMap] 注入成功! 远程基址: 0x%X", remoteBase)

	// 返回模拟的clientID (使用PID作为标识)
	return wechatPID, nil
}

// injectClassic 使用经典方式注入 (兼容模式)
func (inj *Injector) injectClassic() (uint32, error) {
	log.Println("[Classic] 使用经典LoadLibrary注入...")

	if inj.loader == nil {
		return 0, fmt.Errorf("Loader未初始化")
	}

	// 检查DLL是否需要从嵌入资源提取
	dllPath := inj.dllPath

	// 尝试从嵌入资源提取
	if extracted, err := ExtractEmbeddedDLL("NoveHelper.dll", inj.dllPath); err == nil && extracted {
		log.Println("[Classic] 使用嵌入的DLL资源")
		// 确保退出时清理
		defer os.Remove(inj.dllPath)
	}

	clientID, err := inj.loader.InjectWeChat(dllPath)
	if err != nil {
		return 0, err
	}

	log.Printf("[Classic] 注入成功! ClientID=%d", clientID)
	return clientID, nil
}

// findWeChatProcess 查找微信进程
func (inj *Injector) findWeChatProcess() (uint32, error) {
	// 使用 Windows API 枚举进程查找 WeChat.exe
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, fmt.Errorf("创建进程快照失败: %v", err)
	}
	defer syscall.CloseHandle(snapshot)

	var procEntry syscall.ProcessEntry32
	procEntry.Size = uint32(unsafe.Sizeof(procEntry))

	// 获取第一个进程
	err = syscall.Process32First(snapshot, &procEntry)
	if err != nil {
		return 0, fmt.Errorf("枚举进程失败: %v", err)
	}

	// 遍历所有进程
	for {
		// 将进程名转换为字符串
		exeFile := syscall.UTF16ToString(procEntry.ExeFile[:])

		// 查找 WeChat.exe (不区分大小写)
		if strings.EqualFold(exeFile, "WeChat.exe") {
			log.Printf("[Injector] 找到微信进程: %s (PID=%d)", exeFile, procEntry.ProcessID)
			return procEntry.ProcessID, nil
		}

		// 获取下一个进程
		err = syscall.Process32Next(snapshot, &procEntry)
		if err != nil {
			break
		}
	}

	return 0, fmt.Errorf("未找到微信进程 (WeChat.exe)")
}

// GetLoader 获取NoveLoader实例
func (inj *Injector) GetLoader() *NoveLoader {
	return inj.loader
}

// Release 释放资源
func (inj *Injector) Release() error {
	if inj.loader != nil {
		return inj.loader.Release()
	}
	return nil
}
