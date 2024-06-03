package controller

import (
	"axiom-blog/internal/controller/article"
	"axiom-blog/internal/controller/auth"
	"axiom-blog/internal/controller/comment"
	"axiom-blog/internal/controller/like"
	"axiom-blog/internal/controller/notify"
	"axiom-blog/internal/controller/object-processing"
	"axiom-blog/internal/controller/review"
	"axiom-blog/internal/controller/user"
	"github.com/gin-gonic/gin"
)

type IRegisterRoute interface {
	RegisterRoute(g *gin.RouterGroup)
}

func RegisterSpecialRoutes(g *gin.RouterGroup) {
	new(user.Controller).RegisterSpecialRoute(g)
	new(article.Controller).RegisterSpecialRoute(g)
	new(notify.Controller).RegisterSpecialRoute(g)
	new(comment.Controller).RegisterSpecialRoute(g)
}

// RegisterPortalRoutes 统一注册portal路由
func RegisterPortalRoutes(g *gin.RouterGroup) {
	IRegisterRoute.RegisterRoute(new(like.Controller), g)
	IRegisterRoute.RegisterRoute(new(comment.Controller), g)
}

// RegisterRoutes 统一注册admin路由
func RegisterRoutes(g *gin.RouterGroup) {
	IRegisterRoute.RegisterRoute(new(user.Controller), g)
	IRegisterRoute.RegisterRoute(new(article.Controller), g)
	IRegisterRoute.RegisterRoute(new(auth.Controller), g)
	IRegisterRoute.RegisterRoute(new(review.Controller), g)
	IRegisterRoute.RegisterRoute(new(notify.Controller), g)
	IRegisterRoute.RegisterRoute(new(object_processing.Controller), g)
}
