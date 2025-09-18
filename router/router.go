package router

import (
	"github.com/zxc7563598/fintrack-backend/controller"
	"github.com/zxc7563598/fintrack-backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LanguageMiddleware("zh"))
	// 初始化 RSA 私钥
	err := middleware.InitRSAKey("./private.pem")
	if err != nil {
		panic(err)
	}
	// 不需要认证的路由
	noAuthGroup := r.Group("/api")
	{
		noAuthGroup.POST("/register", middleware.DecryptMiddleware[controller.LoginRegisterRequest](), controller.LoginRegisterHandler)
		noAuthGroup.POST("/login", middleware.DecryptMiddleware[controller.LoginRequest](), controller.LoginHandler)
		noAuthGroup.POST("/refresh-token", middleware.DecryptMiddleware[controller.RefreshTokenRequest](), controller.RefreshTokenHandler)
	}
	// 需要认证的路由
	authGroup := r.Group("/api", middleware.AuthMiddleware())
	{
		authGroup.POST("/asset-overview", controller.AssetOverviewHandler)

		authGroup.POST("/user/email", controller.GetUserEmailsHandler)
		authGroup.POST("/user/email/save", middleware.DecryptMiddleware[controller.StoreUserEmailRequest](), controller.StoreUserEmailHandler)
		authGroup.POST("/user/email/delete", middleware.DecryptMiddleware[controller.DeleteUserEmailRequest](), controller.DeleteUserEmailHandler)

		authGroup.POST("/bills", middleware.DecryptMiddleware[controller.GetBillListRequest](), controller.GetBillListHandler)
		authGroup.POST("/bills/info", middleware.DecryptMiddleware[controller.GetBillInfoRequest](), controller.GetBillInfoHandler)
		authGroup.POST("/bills/save", middleware.DecryptMiddleware[controller.StoreBillRecordRequest](), controller.StoreBillRecordHandler)
		authGroup.POST("/bills/delete", middleware.DecryptMiddleware[controller.DeleteBillRecordRequest](), controller.DeleteBillRecordHandler)

		authGroup.POST("/file/alipay/upload/csv", controller.UploadAlipayCSVHandler)
		authGroup.POST("/file/alipay/upload/zip", controller.UploadAlipayZIPHandler)
		authGroup.POST("/file/alipay/overview", middleware.DecryptMiddleware[controller.GetAlipayCSVOverviewRequest](), controller.GetAlipayCSVOverviewHandler)
		authGroup.POST("/file/alipay/store", middleware.DecryptMiddleware[controller.StoreAlipayCSVInfoRequest](), controller.StoreAlipayCSVInfoHandler)
	}
	return r
}
