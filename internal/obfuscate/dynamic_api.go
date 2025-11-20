package obfuscate

import (
	"syscall"
	"unsafe"
)

// DynamicAPI 动态 API 加载器
type DynamicAPI struct {
	dll  *syscall.DLL
	proc *syscall.Proc
}

// LoadAPI 动态加载 Windows API
func LoadAPI(dllName, procName string) (*DynamicAPI, error) {
	// 添加混淆延迟
	JitterDelay(20)

	dll, err := syscall.LoadDLL(dllName)
	if err != nil {
		return nil, err
	}

	// 添加混淆延迟
	JitterDelay(10)

	proc, err := dll.FindProc(procName)
	if err != nil {
		return nil, err
	}

	return &DynamicAPI{
		dll:  dll,
		proc: proc,
	}, nil
}

// Call 调用 API (最多 9 个参数)
func (api *DynamicAPI) Call(args ...uintptr) (uintptr, error) {
	numArgs := len(args)
	if numArgs > 9 {
		numArgs = 9
	}

	// 填充参数数组
	var a [9]uintptr
	for i := 0; i < numArgs; i++ {
		a[i] = args[i]
	}

	ret, _, err := syscall.Syscall9(
		api.proc.Addr(),
		uintptr(numArgs),
		a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8],
	)

	// Windows API 成功时也可能返回 err，需要检查返回值
	if ret == 0 && err != 0 {
		return ret, err
	}

	return ret, nil
}

// Release 释放 DLL
func (api *DynamicAPI) Release() error {
	if api.dll != nil {
		return api.dll.Release()
	}
	return nil
}

// DynamicAPIPool API 池 (缓存已加载的 API)
type DynamicAPIPool struct {
	cache map[string]*DynamicAPI
}

// NewDynamicAPIPool 创建 API 池
func NewDynamicAPIPool() *DynamicAPIPool {
	return &DynamicAPIPool{
		cache: make(map[string]*DynamicAPI),
	}
}

// Get 获取 API (缓存机制)
func (pool *DynamicAPIPool) Get(dllName, procName string) (*DynamicAPI, error) {
	key := dllName + "::" + procName

	if api, exists := pool.cache[key]; exists {
		return api, nil
	}

	api, err := LoadAPI(dllName, procName)
	if err != nil {
		return nil, err
	}

	pool.cache[key] = api
	return api, nil
}

// ReleaseAll 释放所有 API
func (pool *DynamicAPIPool) ReleaseAll() {
	for _, api := range pool.cache {
		api.Release()
	}
	pool.cache = make(map[string]*DynamicAPI)
}

// 全局 API 池
var globalAPIPool = NewDynamicAPIPool()

// GetAPI 从全局池获取 API
func GetAPI(dllName, procName string) (*DynamicAPI, error) {
	return globalAPIPool.Get(dllName, procName)
}

// CallAPI 直接调用 API (简化版)
func CallAPI(dllName, procName string, args ...uintptr) (uintptr, error) {
	api, err := GetAPI(dllName, procName)
	if err != nil {
		return 0, err
	}
	return api.Call(args...)
}

// StringToUintptr 字符串转 uintptr (用于 API 调用)
func StringToUintptr(s string) uintptr {
	b := append([]byte(s), 0)
	return uintptr(unsafe.Pointer(&b[0]))
}
