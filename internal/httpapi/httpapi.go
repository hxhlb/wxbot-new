package httpapi

import (
    "encoding/json"
    "log"
    "net/http"

    "wxbot-new/internal/wxsvc"
)

// RegisterRoutes 注册 HTTP 路由
func RegisterRoutes(mux *http.ServeMux, svc *wxsvc.Service) {
    mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var body map[string]any
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            http.Error(w, "invalid json", http.StatusBadRequest)
            return
        }

        // 支持自定义 payload 或文本直发
        var payload map[string]any

        if t, ok := body["type"]; ok {
            if data, ok2 := body["data"].(map[string]any); ok2 {
                payload = map[string]any{
                    "type": t,
                    "data": data,
                }
            }
        }

        if payload == nil {
            text, _ := body["text"].(string)
            if text == "" {
                // 兼容 message 字段
                text, _ = body["message"].(string)
            }
            if text == "" {
                http.Error(w, "text is required", http.StatusBadRequest)
                return
            }
            room, _ := body["room_wxid"].(string)
            if room == "" {
                room = "47945916190@chatroom"
            }
            payload = map[string]any{
                "type": wxsvc.MT_SEND_TEXTMSG,
                "data": map[string]any{
                    "room_wxid": room,
                    "content":   text,
                },
            }
        }

        ok := svc.SendPayload(payload)
        resp := map[string]any{"success": ok, "payload": payload}
        b, _ := json.Marshal(resp)
        if !ok {
            w.WriteHeader(http.StatusInternalServerError)
        }
        if _, err := w.Write(b); err != nil {
            log.Printf("响应写入失败: %v", err)
        }
    })
}

