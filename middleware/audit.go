package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"policy-backend/user"
	"time"

	"github.com/labstack/echo/v4"
)

type AuditLog struct {
	Timestamp    string `json:"timestamp"`
	Level        string `json:"level"`
	RemoteIP     string `json:"remote_ip"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	Query        string `json:"query"`
	UserID       uint   `json:"user_id,omitempty"`
	Username     string `json:"username,omitempty"`
	StatusCode   int    `json:"status_code"`
	ResponseTime int64  `json:"response_time_ms"`
	UserAgent    string `json:"user_agent,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

func (l *AuditLog) String() string {
	logStr := fmt.Sprintf("[%s] [%s] %s %s%s",
		l.Timestamp,
		l.Level,
		l.Method,
		l.Path,
		l.Query,
	)

	if l.UserID > 0 {
		logStr += fmt.Sprintf(" [user:%d@%s]", l.UserID, l.Username)
	}

	logStr += fmt.Sprintf(" => %d (%dms)", l.StatusCode, l.ResponseTime)

	if l.ErrorMessage != "" {
		logStr += fmt.Sprintf(" Error: %s", l.ErrorMessage)
	}

	return logStr
}

func (l *AuditLog) JSON() string {
	data, _ := json.Marshal(l)
	return string(data)
}

type AuditConfig struct {
	EnableJSON   bool
	LogLevel     string
	LogFile      string
	ExcludePaths []string
	IncludePaths []string
}

var defaultAuditConfig = AuditConfig{
	EnableJSON: false,
	LogLevel:   "info",
	LogFile:    "",
}

func AuditMiddleware(config *AuditConfig) echo.MiddlewareFunc {
	if config == nil {
		config = &defaultAuditConfig
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			path := c.Request().URL.Path
			method := c.Request().Method
			query := c.Request().URL.RawQuery

			if query != "" {
				query = "?" + query
			}

			if shouldExclude(path, config.ExcludePaths) {
				return next(c)
			}

			if len(config.IncludePaths) > 0 && !shouldInclude(path, config.IncludePaths) {
				return next(c)
			}

			err := next(c)

			responseTime := time.Since(start).Milliseconds()

			log := &AuditLog{
				Timestamp:    time.Now().Format(time.RFC3339),
				Level:        config.LogLevel,
				RemoteIP:     c.RealIP(),
				Method:       method,
				Path:         path,
				Query:        query,
				StatusCode:   c.Response().Status,
				ResponseTime: responseTime,
				UserAgent:    c.Request().UserAgent(),
			}

			if currentUser, ok := c.Get("user").(*user.User); ok {
				log.UserID = currentUser.ID
				log.Username = currentUser.Username
			}

			if err != nil {
				log.ErrorMessage = err.Error()
				if log.StatusCode == 0 {
					log.StatusCode = http.StatusInternalServerError
				}
			}

			if config.LogFile != "" {
				file, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err == nil {
					defer file.Close()
					if config.EnableJSON {
						file.WriteString(log.JSON() + "\n")
					} else {
						file.WriteString(log.String() + "\n")
					}
				}
			}

			if config.EnableJSON {
				fmt.Println(log.JSON())
			} else {
				fmt.Println(log.String())
			}

			return err
		}
	}
}

func shouldExclude(path string, excludePaths []string) bool {
	for _, p := range excludePaths {
		if matchPath(path, p) {
			return true
		}
	}
	return false
}

func shouldInclude(path string, includePaths []string) bool {
	for _, p := range includePaths {
		if matchPath(path, p) {
			return true
		}
	}
	return false
}

func matchPath(path, pattern string) bool {
	if path == pattern {
		return true
	}

	if len(pattern) > 2 && pattern[len(pattern)-2:] == "/*" {
		prefix := pattern[:len(pattern)-2]
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}

	if len(pattern) > 1 && pattern[len(pattern)-1:] == "/" {
		if path == pattern[:len(pattern)-1] {
			return true
		}
	}

	return false
}

func GetCurrentUser(c echo.Context) *user.User {
	if user, ok := c.Get("user").(*user.User); ok {
		return user
	}
	return nil
}

func GetCurrentUserID(c echo.Context) uint {
	if user := GetCurrentUser(c); user != nil {
		return user.ID
	}
	return 0
}

func GetCurrentUsername(c echo.Context) string {
	if user := GetCurrentUser(c); user != nil {
		return user.Username
	}
	return ""
}

func LogInfo(message string, fields ...interface{}) {
	log := &AuditLog{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "info",
	}
	fmt.Printf("[%s] [INFO] %s\n", log.Timestamp, fmt.Sprintf(message, fields...))
}

func LogWarning(message string, fields ...interface{}) {
	log := &AuditLog{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "warning",
	}
	fmt.Printf("[%s] [WARNING] %s\n", log.Timestamp, fmt.Sprintf(message, fields...))
}

func LogError(message string, fields ...interface{}) {
	log := &AuditLog{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "error",
	}
	fmt.Printf("[%s] [ERROR] %s\n", log.Timestamp, fmt.Sprintf(message, fields...))
}

func LogDebug(message string, fields ...interface{}) {
	log := &AuditLog{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "debug",
	}
	fmt.Printf("[%s] [DEBUG] %s\n", log.Timestamp, fmt.Sprintf(message, fields...))
}
