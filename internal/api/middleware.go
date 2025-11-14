package api

import (
	"log"
	"net/http"
	"time"
)

// Middleware 中间件函数类型
type Middleware func(http.Handler) http.Handler

// Chain 链式组合多个中间件
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 创建响应包装器以捕获状态码
			wrapper := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

			// 处理请求
			next.ServeHTTP(wrapper, r)

			// 记录日志
			log.Printf("[HTTP] %s %s - Status: %d - 耗时: %v",
				r.Method,
				r.URL.Path,
				wrapper.statusCode,
				time.Since(start),
			)
		})
	}
}

// CORSMiddleware CORS 中间件
func CORSMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 设置 CORS 头
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "3600")

			// 处理预检请求
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryMiddleware 恢复中间件(捕获 panic)
func RecoveryMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("[PANIC] %s %s - Error: %v", r.Method, r.URL.Path, err)
					ErrorResponse(w, http.StatusInternalServerError, "服务器内部错误")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// ContentTypeMiddleware 内容类型中间件
func ContentTypeMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 对于 POST/PUT 请求,检查 Content-Type
			if r.Method == http.MethodPost || r.Method == http.MethodPut {
				contentType := r.Header.Get("Content-Type")
				if contentType != "" && contentType != "application/json" {
					ErrorResponse(w, http.StatusUnsupportedMediaType, "仅支持 application/json")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// responseWrapper 响应包装器,用于捕获状态码
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
