package wxsvc

// 常量与类型定义，贴合 pythondemo.py
const (
    MT_DEBUG_LOG    = 11024
    MT_USER_LOGIN   = 11025
    MT_USER_LOGOUT  = 11026 // 注意: 源文件重复定义了 11026/11027，这里保留常用
    MT_SEND_TEXTMSG = 11036
)

