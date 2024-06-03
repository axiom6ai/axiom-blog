package globalInit

import (
	"github.com/gin-gonic/gin"
	"io"
	"os"
)

// SetLog log设置
func (a *app) SetLog() {
	// 禁用控制台颜色，将日志写入文件时不需要控制台颜色。
	gin.DisableConsoleColor()

	// 记录到文件。
	f, _ := os.Create("axiom-log.log")
	gin.DefaultWriter = io.MultiWriter(f)

	// 需要同时将日志写入文件和控制台
	if conf.Stdout {
		gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	}
}
