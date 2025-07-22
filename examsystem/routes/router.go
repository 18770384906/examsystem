// routes/router.go
package routes

import (
	"examsystem/controllers"
	"examsystem/middleware"

	"github.com/gin-gonic/gin"
)

// AppDependencies 定义应用依赖接口
type AppDependencies interface {
	GetUserController() *controllers.UserController
	GetAuthController() *controllers.AuthController
	GetQuestionController() *controllers.QuestionController
	// GetPaperController() *controllers.PaperController
}

// SetupRouter 配置所有路由
func SetupRouter(deps AppDependencies) *gin.Engine {
	r := gin.Default()

	// API 路由组
	api := r.Group("/api")
	{
		// 创建控制器实例
		authController := deps.GetAuthController()
		userController := deps.GetUserController()
		questionController := deps.GetQuestionController()
		// paperController := deps.GetPaperController()

		// 认证相关路由（无需认证）
		auth := api.Group("/auth")
		{
			auth.POST("/login", authController.Login)
			auth.POST("/logout", authController.Logout)
			auth.POST("/register", userController.Create)
		}

		// 需要认证的路由
		authorized := api.Group("/")
		authorized.Use(middleware.JWTAuth())
		{
			// 认证相关
			authorized.GET("/auth/me", authController.Me)

			// 用户路由
			userGroup := authorized.Group("/users")
			{
				userGroup.GET("/:id", userController.Get)
				userGroup.GET("", userController.List)
			}

			// 管理员权限路由
			admin := authorized.Group("/admin")
			admin.Use(middleware.AdminAuth())
			{
				admin.POST("/users", userController.Create)
				admin.PUT("/users/:id", userController.Update)
				admin.DELETE("/users/:id", userController.Delete)
			}

			// 题目管理路由（普通用户可以管理自己的题目）
			questionGroup := authorized.Group("/questions")
			{
				questionGroup.POST("/generate", questionController.GenerateQuestionsHandler)
				questionGroup.POST("/confirm", questionController.SaveSelectedQuestionsHandler)
				questionGroup.GET("", questionController.GetQuestionsByUserIDHandler)
				// questionGroup.GET("/:id", questionController.GetQuestionByIDHandler)
				// 普通用户可以编辑和删除自己的题目
				questionGroup.PUT("/:id", questionController.UpdateQuestionHandler)
				questionGroup.DELETE("/:id", questionController.DeleteQuestionHandler)
			}

			// 试卷管理路由
			paperGroup := authorized.Group("/papers")
			{
				paperGroup.GET("", paperController.GetPapersHandler)          // 获取试卷列表
				paperGroup.POST("", paperController.CreatePaperHandler)       // 创建试卷
				paperGroup.GET("/:id", paperController.GetPaperHandler)       // 获取试卷详情
				paperGroup.PUT("/:id", paperController.UpdatePaperHandler)    // 更新试卷信息
				paperGroup.DELETE("/:id", paperController.DeletePaperHandler) // 删除试卷

				// 试卷题目管理
				paperQuestionGroup := paperGroup.Group("/:id/questions")
				{
					paperQuestionGroup.POST("", paperController.AddQuestionToPaperHandler)                    // 添加题目到试卷
					paperQuestionGroup.DELETE("/:questionId", paperController.RemoveQuestionFromPaperHandler) // 从试卷中移除题目
					paperQuestionGroup.PUT("/order", paperController.UpdateQuestionOrderHandler)              // 更新试卷题目顺序
				}
			}

			// 统计路由
			statsGroup := authorized.Group("/statistics")
			{
				statsGroup.GET("/user", paperController.GetUserStatisticsHandler) // 获取用户统计信息
			}
		}
	}

	return r
}
