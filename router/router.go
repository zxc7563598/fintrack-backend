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
		authGroup.POST("/bills/calendar", middleware.DecryptMiddleware[controller.GetBillCalendarRequest](), controller.GetBillCalendarHandler)
		authGroup.POST("/bills/info", middleware.DecryptMiddleware[controller.GetBillInfoRequest](), controller.GetBillInfoHandler)
		authGroup.POST("/bills/save", middleware.DecryptMiddleware[controller.StoreBillRecordRequest](), controller.StoreBillRecordHandler)
		authGroup.POST("/bills/delete", middleware.DecryptMiddleware[controller.DeleteBillRecordRequest](), controller.DeleteBillRecordHandler)

		authGroup.POST("/file/alipay/upload/csv", controller.UploadAlipayCSVHandler)
		authGroup.POST("/file/alipay/upload/zip", controller.UploadAlipayZIPHandler)
		authGroup.POST("/file/wechat/upload/xlsx", controller.UploadWeChatXLSXHandler)
		authGroup.POST("/file/wechat/upload/zip", controller.UploadWeChatZIPHandler)

		authGroup.POST("/file/alipay/overview", middleware.DecryptMiddleware[controller.GetAlipayCSVOverviewRequest](), controller.GetAlipayCSVOverviewHandler)
		authGroup.POST("/file/wechat/overview", middleware.DecryptMiddleware[controller.GetWeChatXLSXOverviewRequest](), controller.GetWeChatXLSXOverviewHandler)
		authGroup.POST("/file/alipay/store", middleware.DecryptMiddleware[controller.StoreAlipayCSVInfoRequest](), controller.StoreAlipayCSVInfoHandler)
		authGroup.POST("/file/wechat/store", middleware.DecryptMiddleware[controller.StoreWechatXLSXInfoRequest](), controller.StoreWechatXLSXInfoHandler)

		authGroup.POST("/file/alipay/email", middleware.DecryptMiddleware[controller.GetAlipayBillMailRequest](), controller.GetAlipayBillMailHandler)

		authGroup.POST("/statistics/account/category", middleware.DecryptMiddleware[controller.GetStatisticsRequest](), controller.AccountBalanceCategoryHandler)
		authGroup.POST("/statistics/income/category", middleware.DecryptMiddleware[controller.GetStatisticsRequest](), controller.IncomeCategoryHandler)
		authGroup.POST("/statistics/expense/category", middleware.DecryptMiddleware[controller.GetStatisticsRequest](), controller.ExpenseCategoryHandler)
		authGroup.POST("/statistics/income/account/category", middleware.DecryptMiddleware[controller.GetStatisticsRequest](), controller.IncomeAccountCategoryHandler)
		authGroup.POST("/statistics/expense/account/category", middleware.DecryptMiddleware[controller.GetStatisticsRequest](), controller.ExpenseAccountCategoryHandler)
		authGroup.POST("/statistics/account-balance/trend", middleware.DecryptMiddleware[controller.GetStatisticsRequest](), controller.AccountBalanceTrendHandler)
		authGroup.POST("/statistics/income/trend", middleware.DecryptMiddleware[controller.GetStatisticsRequest](), controller.IncomeCategoryTrendHandler)
		authGroup.POST("/statistics/expense/trend", middleware.DecryptMiddleware[controller.GetStatisticsRequest](), controller.ExpenseCategoryTrendHandler)
		authGroup.POST("/statistics/income/account/trend", middleware.DecryptMiddleware[controller.GetStatisticsRequest](), controller.IncomeAccountTrendHandler)
		authGroup.POST("/statistics/expense/account/trend", middleware.DecryptMiddleware[controller.GetStatisticsRequest](), controller.ExpenseAccountTrendHandler)
	}
	return r
}
