package notify

import (
	"axiom-blog/internal/notify/qo"
	"axiom-blog/internal/notify/service/impl"
	"github.com/gin-gonic/gin"
)

/**
  @author: xichencx@163.com
  @date:2021/12/13
  @description: 通知服务路由
**/

type Controller struct{}

var notify = &impl.Notify{}

func (c Controller) RegisterRoute(g *gin.RouterGroup) {
	notifyGroup := g.Group("/notify")

	//新增通知
	notifyGroup.POST("/add", func(context *gin.Context) {
		query := new(qo.AddNotificationQO)
		notify.AddNotification(context, query)
	})

	//更新通知
	notifyGroup.POST("/update", func(context *gin.Context) {
		query := new(qo.UpdateNotificationQO)
		notify.UpdateNotification(context, query)
	})
}

func (c Controller) RegisterSpecialRoute(g *gin.RouterGroup) {
	notifyGroup := g.Group("/notify")

	//查询系统通知
	notifyGroup.POST("/query", notify.SystemNotify)
}
