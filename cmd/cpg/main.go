package main

import (
	"axiom-blog/config"
	"axiom-blog/global/globalInit"
	"axiom-blog/internal/controller"
	"axiom-blog/middleware"
	"context"
	"errors"
	"flag"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	//编译信息，默认unknown
	gitCommitLog = "unknown commit"
	buildTime    = "build time "
	gitRelease   = "unknown"
	localDebug   = flag.Bool("local", false, "本地启动默认不启动https服务")
)

func init() {
	config.ConfInit()
	globalInit.DbInit()
	//globalInit.RedisInit()
	globalInit.App.SetFrameMode(gin.ReleaseMode)
	globalInit.App.FillBuildInfo(gitCommitLog, buildTime, gitRelease)
	globalInit.App.SetLog()
}

func main() {
	if *localDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Printf("Added %v route %v -> %v (%d handlers)\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}
	// header
	r.Use(middleware.NoCache)
	r.Use(middleware.Secure)
	r.Use(middleware.Options)

	//TODO 服务访问权限，通过OAuth服务实现

	// 后端路由组
	//登录等特殊页
	special := r.Group("/special")
	controller.RegisterSpecialRoutes(special)

	//门户页
	portal := r.Group("")
	portal.Use(middleware.JwtAuth)
	controller.RegisterPortalRoutes(portal)

	//管理页
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.JwtAuth)
	adminGroup.Use(middleware.PermissionAuth)
	controller.RegisterRoutes(adminGroup)

	//TODO 获取CA证书，并自动更新
	srv := &http.Server{
		Addr:           ":8000",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if *localDebug {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Printf("服务启动失败: %s\n", err)
			}
		} else {
			if err := srv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Printf("启动TLS服务失败: %s\n", err)
			}
		}
	}()

	log.Printf("服务启动成功：%v", globalInit.App)

	// 等待中断信号以优雅地关闭服务器，为关闭服务器操作设置一个5秒的超时
	quit := make(chan os.Signal, 1)

	// kill (无参数) 默认发送 syscall.SIGTERM
	// kill -2 发送 syscall.SIGINT
	// kill -9 发送 syscall.SIGKILL 但是不能被捕获，所以不需要添加它
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞在此，等待信号
	<-quit
	log.Println("Shutdown Server ...")

	// 创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 不阻塞的关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown:", err)
	}

	// 捕获ctx.Done()的信号
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}
