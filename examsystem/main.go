package main

import (
	"examsystem/config"
	"examsystem/controllers"
	"examsystem/dao"
	"examsystem/routes"
	"examsystem/service"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 应用依赖
type AppDependencies struct {
	DB                 *gorm.DB
	UserDAO            *dao.UserDAO
	QuestionDAO        *dao.QuestionDAO
	UserService        *service.UserService
	QuestionService    *service.QuestionService
	userController     *controllers.UserController
	authController     *controllers.AuthController
	questionController *controllers.QuestionController
}

// GetUserController 获取用户控制器
func (d *AppDependencies) GetUserController() *controllers.UserController {
	if d.userController == nil {
		d.userController = controllers.NewUserController(d.UserService)
	}
	return d.userController
}

// GetAuthController 获取认证控制器
func (d *AppDependencies) GetAuthController() *controllers.AuthController {
	if d.authController == nil {
		d.authController = controllers.NewAuthController(d.UserService)
	}
	return d.authController
}

// GetQuestionController 获取题目控制器
func (d *AppDependencies) GetQuestionController() *controllers.QuestionController {
	if d.questionController == nil {
		d.questionController = controllers.NewQuestionController(d.QuestionService)
	}
	return d.questionController
}

func main() {
	// 获取配置
	appConfig := config.GetConfig()

	// 初始化数据库连接（根据是否 reset）
	db, err := dao.InitDB(appConfig.ResetDB)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	log.Println("数据库初始化完成")

	// 初始化依赖
	deps := initDependencies(db)

	// 设置Gin模式
	gin.SetMode(appConfig.Mode)

	// 设置路由
	r := routes.SetupRouter(deps)

	// 日志输出
	log.Printf("服务器启动于 %s 端口，运行模式: %s\n", appConfig.ServerPort, appConfig.Mode)

	// 启动服务器
	err = r.Run(appConfig.ServerPort)
	if err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// 初始化依赖
func initDependencies(db *gorm.DB) *AppDependencies {
	// 初始化DAO
	userDAO := dao.NewUserDAO(db)
	questionDAO := dao.NewQuestionDAO(db)

	// 初始化服务
	userService := service.NewUserService(userDAO)
	questionService := service.NewQuestionService(questionDAO, config.LoadAIConfig())

	return &AppDependencies{
		DB:              db,
		UserDAO:         userDAO,
		QuestionDAO:     questionDAO,
		UserService:     userService,
		QuestionService: questionService,
	}
}
