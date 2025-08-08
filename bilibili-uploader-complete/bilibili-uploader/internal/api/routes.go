package api

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
 )

// SetupRoutes 设置路由
func SetupRoutes(handler *Handler) *gin.Engine {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.New()
	
	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// 配置CORS
	config := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(config))
	
	// API路由组 (优先注册API路由)
	api := router.Group("/api")
	{
		// 健康检查
		api.GET("/health", handler.HealthCheck)
		
		// 上传相关接口
		upload := api.Group("/upload")
		{
			upload.POST("/init", handler.InitUpload)           // 初始化上传
			upload.POST("/chunk", handler.UploadChunk)         // 上传分片
			upload.POST("/complete", handler.CompleteUpload)   // 完成上传
			upload.POST("/cover", handler.UploadCover)         // 上传封面
			upload.GET("/status/:upload_id", handler.GetUploadStatus) // 获取上传状态
		}
		
		// 提交相关接口
		submit := api.Group("/submit")
		{
			submit.POST("/video", handler.SubmitVideo) // 提交视频稿件
		}
	}

    // **修改这里**：移除StaticFS，使用NoRoute处理所有未匹配的路由
    // 对于所有未匹配的路由（非API路由），都返回index.html，用于单页应用路由
    router.NoRoute(func(c *gin.Context) {
        // 尝试从web/static目录提供文件
        filepath := c.Request.URL.Path
        if filepath == "/" {
            filepath = "/index.html"
        }
        // 检查文件是否存在于web/static中
        if _, err := http.Dir("./web/static" ).Open(filepath); err == nil {
            c.File("./web/static" + filepath)
        } else {
            // 如果不是静态文件，则返回index.html (SPA回退)
            c.File("./web/static/index.html")
        }
    })
    	
    	return router
}
