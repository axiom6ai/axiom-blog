package service

import (
	"axiom-blog/internal/notify/qo"
	"github.com/gin-gonic/gin"
)

/**
  @author: xichencx@163.com
  @date:2021/12/6
  @description: 通知服务接口定义
**/

type INotify interface {
	// AddNotification 添加通知内容
	AddNotification(ctx *gin.Context, query *qo.AddNotificationQO)

	// UpdateNotification 更新通知
	UpdateNotification(ctx *gin.Context, query *qo.UpdateNotificationQO)

	// SystemNotify 查询通知内容
	SystemNotify(ctx *gin.Context)
}
