package loader

import (
	_ "embed"
	"fmt"
	"os"
)

// 嵌入DLL文件
// 这些变量会在编译时自动填充
// 如果文件不存在,编译会失败,所以需要条件编译

//go:embed embedded/NoveLoader.dll
var embeddedNoveLoader []byte

//go:embed embedded/NoveHelper.dll
var embeddedNoveHelper []byte

// GetEmbeddedDLL 获取嵌入的DLL数据
func GetEmbeddedDLL(name string) ([]byte, error) {
	switch name {
	case "NoveLoader.dll":
		if len(embeddedNoveLoader) == 0 {
			return nil, fmt.Errorf("NoveLoader.dll 未嵌入")
		}
		return embeddedNoveLoader, nil
	case "NoveHelper.dll":
		if len(embeddedNoveHelper) == 0 {
			return nil, fmt.Errorf("NoveHelper.dll 未嵌入")
		}
		return embeddedNoveHelper, nil
	default:
		return nil, fmt.Errorf("未知的DLL: %s", name)
	}
}

// ExtractEmbeddedDLL 提取嵌入的DLL到文件
func ExtractEmbeddedDLL(name, outputPath string) (bool, error) {
	data, err := GetEmbeddedDLL(name)
	if err != nil {
		return false, err
	}

	// 写入文件
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return false, fmt.Errorf("写入文件失败: %v", err)
	}

	return true, nil
}

// HasEmbeddedDLL 检查是否有嵌入的DLL
func HasEmbeddedDLL(name string) bool {
	_, err := GetEmbeddedDLL(name)
	return err == nil
}
