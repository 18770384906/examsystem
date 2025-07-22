package controllers

import (
	"encoding/json"
	"examsystem/dao/model"
	"examsystem/service"
	"examsystem/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type QuestionController struct {
	questionService *service.QuestionService
}

func NewQuestionController(questionService *service.QuestionService) *QuestionController {
	return &QuestionController{
		questionService: questionService,
	}
}

// GenerateQuestionsHandler 生成题目
func (c *QuestionController) GenerateQuestionsHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.Unauthorized(ctx, "未登录")
		return
	}

	aiModel := ctx.Query("ai_model")
	language := ctx.Query("language")
	questionType := ctx.Query("question_type")
	keywords := ctx.Query("keywords")
	numQuestions, _ := strconv.Atoi(ctx.Query("num_questions"))

	if questionType != string(model.QuestionTypeSingle) && questionType != string(model.QuestionTypeMultiple) {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的题目类型", "data": nil})
		return
	}

	questions, err := c.questionService.GenerateQuestions(int64(userID.(uint)), aiModel, language, model.QuestionType(questionType), keywords, numQuestions)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成题目失败", "data": nil})
		return
	}

	var response []map[string]interface{}
	for _, q := range questions {
		var opts []string
		json.Unmarshal([]byte(q.Options), &opts)

		response = append(response, map[string]interface{}{
			"id":           q.ID,
			"title":        q.Title,
			"questionType": q.QuestionType,
			"options":      opts,
			"answer":       q.Answer,
			"explanation":  q.Explanation,
			"keywords":     q.Keywords,
			"language":     q.Language,
			"aiModel":      q.AIModel,
			"userID":       q.UserID,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "生成成功", "data": response})
}

// SaveSelectedQuestionsHandler 保存选中的题目
func (c *QuestionController) SaveSelectedQuestionsHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.Unauthorized(ctx, "未登录")
		return
	}

	var req struct {
		SelectedIDs []int64 `json:"selected_ids"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "data": nil})
		return
	}

	err := c.questionService.SaveSelectedQuestions(int64(userID.(uint)), req.SelectedIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "确认失败", "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "确认成功", "data": nil})
}

// GetQuestionsByUserIDHandler 获取用户的题目,DeletedAt IS NULL 筛选逻辑
func (c *QuestionController) GetQuestionsByUserIDHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.Unauthorized(ctx, "未登录")
		return
	}

	// language := ctx.Query("language")
	// questionType := ctx.Query("question_type")
	// keyword := ctx.Query("keyword")

	questions, err := c.questionService.GetQuestionsByUserID(int64(userID.(uint)), "", "", "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取失败", "data": nil})
		return
	}

	var result []map[string]interface{}
	for _, q := range questions {
		var opts []string
		json.Unmarshal([]byte(q.Options), &opts)

		result = append(result, map[string]interface{}{
			"id":           q.ID,
			"title":        q.Title,
			"questionType": q.QuestionType,
			"options":      opts,
			"answer":       q.Answer,
			"explanation":  q.Explanation,
			"keywords":     q.Keywords,
			"language":     q.Language,
			"aiModel":      q.AIModel,
			"userID":       q.UserID,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": result})
}

// UpdateQuestionHandler 更新题目
func (c *QuestionController) UpdateQuestionHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.Unauthorized(ctx, "未登录")
		return
	}
	// userID = int64(userID.(uint))

	questionID, _ := strconv.ParseInt(ctx.Param("id"), 10, 64)

	// 解析请求体
	var request struct {
		Title        string             `json:"title"`
		QuestionType model.QuestionType `json:"questionType"`
		Options      []string           `json:"options"`
		Answer       string             `json:"answer"`
		Explanation  string             `json:"explanation"`
		Keywords     string             `json:"keywords"`
		Language     string             `json:"language"`
		AIModel      string             `json:"aiModel"`
	}

	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	// 将Options转换为JSON字符串
	optionsJSON, err := json.Marshal(request.Options)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "选项格式错误",
			"data":    nil,
		})
		return
	}

	// 创建要更新的题目对象
	question := &model.Question{
		ID:           questionID,
		UserID:       int64(userID.(uint)),
		Title:        request.Title,
		QuestionType: request.QuestionType,
		Options:      string(optionsJSON),
		Answer:       request.Answer,
		Explanation:  request.Explanation,
		Keywords:     request.Keywords,
		Language:     request.Language,
		AIModel:      request.AIModel,
	}

	// 调用服务层更新题目
	err = c.questionService.UpdateQuestion(question)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新题目失败",
			"data":    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新题目成功",
		"data":    nil,
	})
}

// DeleteQuestionHandler 删除题目
func (c *QuestionController) DeleteQuestionHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.Unauthorized(ctx, "未登录")
		return
	}
	questionID, _ := strconv.ParseInt(ctx.Param("id"), 10, 64)

	err := c.questionService.DeleteQuestion(int64(userID.(uint)), questionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败", "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功", "data": nil})
}
