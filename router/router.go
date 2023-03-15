package router

import "github.com/gin-gonic/gin"

func Router(router *gin.Engine) {
	api := router.Group("/api")
	WebRouter(api.Group("/web"))

}
