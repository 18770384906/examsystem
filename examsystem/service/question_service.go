package service

import (
	"bytes"
	"encoding/json"
	"examsystem/config"
	"examsystem/dao"
	"examsystem/dao/model"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type QuestionService struct {
	questionDAO *dao.QuestionDAO
	aiConfig    config.AIConfig
}

func NewQuestionService(questionDAO *dao.QuestionDAO, aiConfig config.AIConfig) *QuestionService {
	return &QuestionService{
		questionDAO: questionDAO,
		aiConfig:    aiConfig,
	}
}

// GenerateQuestions 生成题目
func (s *QuestionService) GenerateQuestions(userID int64, aiModel, language string, questionType model.QuestionType, keywords string, numQuestions int) ([]*model.Question, error) {
	// 验证题目类型
	if questionType != model.QuestionTypeSingle && questionType != model.QuestionTypeMultiple {
		return nil, fmt.Errorf("无效的题目类型: %s", questionType)
	}

	// 构造提示语
	prompt := s.constructPrompt(aiModel, language, string(questionType), keywords, numQuestions)

	// 获取API密钥和URL
	apiKey := s.getAPIKey(aiModel)
	url := s.getAPIURL(aiModel)

	log.Printf("AI模型: %s, API Key: %s, URL: %s", aiModel, apiKey, url)

	// 调用AI API
	questions, err := s.callAIAPI(url, apiKey, prompt)
	if err != nil {
		return nil, err
	}

	// 设置元信息并准备逻辑删除状态
	// now := time.Now()
	for _, question := range questions {
		question.UserID = userID
		question.AIModel = aiModel
		question.Language = language
		question.Keywords = keywords
		question.DeletedAt.Time = time.Now()
		question.DeletedAt.Valid = true
	}

	// 保存题目到数据库（逻辑保存）
	if err := s.questionDAO.BatchCreateQuestions(questions); err != nil {
		return nil, fmt.Errorf("保存题目失败: %v", err)
	}

	return questions, nil
}

// constructPrompt 构造AI提示语
func (s *QuestionService) constructPrompt(aiModel, language, questionType, keywords string, numQuestions int) string {
	typeDesc := "单选题"
	if questionType == string(model.QuestionTypeMultiple) {
		typeDesc = "多选题"
	}

	return fmt.Sprintf(`
    请严格按照以下JSON格式生成%d道关于"%s"的%s编程%s，每题必须有4个选项，答案使用选项索引（如"A", "B", "C", "D"）：
    {
        "questions": [
            {
                "title": "题目内容",
                "options": ["选项A", "选项B", "选项C", "选项D"],
                "answer": "正确选项索引",
                "explanation": "答案解析"
            }
        ]
    }
    `, numQuestions, keywords, language, typeDesc)
}

// getAPIKey 获取API密钥
func (s *QuestionService) getAPIKey(aiModel string) string {
	switch aiModel {
	case "通义千问":
		return s.aiConfig.TongyiAPIKey
	case "deepseek":
		return s.aiConfig.DeepSeekAPIKey
	default:
		return ""
	}
}

// getAPIURL 获取API URL
func (s *QuestionService) getAPIURL(aiModel string) string {
	switch aiModel {
	case "通义千问":
		return s.aiConfig.TongyiAPIURL
	case "deepseek":
		return s.aiConfig.DeepSeekAPIURL
	default:
		log.Printf("错误：未知的AI模型: %s", aiModel)
		return ""
	}
}

// callAIAPI 调用AI API
func (s *QuestionService) callAIAPI(url, apiKey, prompt string) ([]*model.Question, error) {
	// 构建符合DeepSeek API格式的请求体
	payload := map[string]interface{}{
		"model": "deepseek-chat", // 指定模型，根据实际情况修改
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,  // 控制生成的随机性，范围0-1
		"max_tokens":  2000, // 最大生成token数
	}

	// 转换为JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 定义可重试的错误类型
	retryableStatusCodes := map[int]bool{
		http.StatusRequestTimeout:      true,
		http.StatusTooManyRequests:     true,
		http.StatusInternalServerError: true,
		http.StatusBadGateway:          true,
		http.StatusServiceUnavailable:  true,
		http.StatusGatewayTimeout:      true,
	}

	// 指数退避重试策略
	var (
		resp          *http.Response
		body          []byte
		maxRetries    = 3
		retryDelay    = 1 * time.Second
		maxRetryDelay = 4 * time.Second
	)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("AI API 请求重试中 (%d/%d)...", attempt, maxRetries)
			time.Sleep(retryDelay)
			retryDelay = min(retryDelay*2, maxRetryDelay) // 指数退避
		}

		// 发送请求
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err = client.Do(req)
		if err != nil {
			log.Printf("AI API 请求失败 (尝试 %d/%d): %v", attempt+1, maxRetries, err)
			if attempt == maxRetries {
				return nil, fmt.Errorf("AI API 请求失败，已达到最大重试次数: %v", err)
			}
			continue
		}

		// 读取响应
		body, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close() // 立即关闭响应体

		if err != nil {
			log.Printf("AI API 读取响应失败 (尝试 %d/%d): %v", attempt+1, maxRetries, err)
			if attempt == maxRetries {
				return nil, fmt.Errorf("AI API 读取响应失败，已达到最大重试次数: %v", err)
			}
			continue
		}

		// 检查响应状态码
		if resp.StatusCode != http.StatusOK {
			log.Printf("AI API 返回非成功状态码 (尝试 %d/%d): %d, 响应: %s",
				attempt+1, maxRetries, resp.StatusCode, string(body))

			// 判断是否为可重试的状态码
			if retryableStatusCodes[resp.StatusCode] && attempt < maxRetries {
				continue
			}

			// 非重试错误，直接返回
			return nil, fmt.Errorf("AI API 错误: %d", resp.StatusCode)
		}

		// 请求成功，跳出重试循环
		break
	}

	// 解析DeepSeek API的标准响应格式
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("JSON解析失败: %v, 响应内容: %s", err, string(body))
		return nil, err
	}

	// 提取AI返回的内容
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("AI API 返回空结果")
	}

	aiResponse := response.Choices[0].Message.Content

	// 解析AI返回的内容为题目列表
	return s.parseAIResponse(aiResponse)
}

// 辅助函数：返回较小值
func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// 解析AI返回的内容为题目列表
func (s *QuestionService) parseAIResponse(content string) ([]*model.Question, error) {
	// 预处理：去除Markdown标记
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// 尝试解析JSON
	var questionsData struct {
		Questions []struct {
			Title       string   `json:"title"`
			Options     []string `json:"options"`
			Answer      string   `json:"answer"`
			Explanation string   `json:"explanation"`
		} `json:"questions"`
	}

	if err := json.Unmarshal([]byte(content), &questionsData); err != nil {
		log.Printf("JSON解析失败: %v, 处理后的内容: %s", err, content)
		return nil, err
	}

	// 验证并转换题目
	var questions []*model.Question
	for _, q := range questionsData.Questions {
		// 验证选项数量
		if len(q.Options) != 4 {
			log.Printf("警告: 题目 '%s' 的选项数量不足4个，实际: %d", q.Title, len(q.Options))
			continue
		}

		// 验证答案格式
		validAnswers := map[string]bool{"A": true, "B": true, "C": true, "D": true}
		if !validAnswers[q.Answer] {
			log.Printf("警告: 题目 '%s' 的答案格式不正确，应为A-D，实际: %s", q.Title, q.Answer)
			// 创建一个字符串数组来存储选项索引
			answerIndex := []string{"A", "B", "C", "D"}

			// 尝试从选项中查找匹配的答案
			for i, option := range q.Options {
				if option == q.Answer {
					q.Answer = answerIndex[i] // 使用字符串数组来获取选项索引
					break
				}
			}

			// 如果仍未找到匹配，跳过该题目
			if !validAnswers[q.Answer] {
				log.Printf("错误: 无法转换题目 '%s' 的答案，跳过", q.Title)
				continue
			}
		}

		// 转换选项为JSON字符串
		optionsJSON, err := json.Marshal(q.Options)
		if err != nil {
			return nil, err
		}

		questions = append(questions, &model.Question{
			Title:        q.Title,
			QuestionType: model.QuestionTypeSingle,
			Options:      string(optionsJSON),
			Answer:       q.Answer,
			Explanation:  q.Explanation,
		})
	}

	if len(questions) == 0 {
		return nil, fmt.Errorf("没有有效的题目被解析")
	}

	return questions, nil
}

// SaveSelectedQuestions 保存选中的题目，入库选中题目并物理删除未选中的题目
func (s *QuestionService) SaveSelectedQuestions(userID int64, selectedIDs []int64) error {
	log.Printf("[DEBUG] SaveSelectedQuestions start: userID=%d selectedIDs=%v", userID, selectedIDs)

	allGeneratedQuestions, err := s.questionDAO.GetGeneratedQuestionsByUserID(userID)
	if err != nil {
		log.Printf("[ERROR] 查询未确认题目失败: %v", err)
		return err
	}

	log.Printf("[DEBUG] 查询到未确认题目数量: %d", len(allGeneratedQuestions))
	for _, q := range allGeneratedQuestions {
		log.Printf("[DEBUG] Question ID: %d, DeletedAt: %v, Valid: %v", q.ID, q.DeletedAt.Time, q.DeletedAt.Valid)
	}

	selectedMap := make(map[int64]bool)
	for _, id := range selectedIDs {
		selectedMap[id] = true
	}

	var toRestoreIDs []int64
	var toDeleteIDs []int64

	for _, q := range allGeneratedQuestions {
		if selectedMap[q.ID] {
			toRestoreIDs = append(toRestoreIDs, q.ID)
		} else {
			toDeleteIDs = append(toDeleteIDs, q.ID)
		}
	}

	log.Printf("[DEBUG] 需要恢复的题目ID: %v", toRestoreIDs)
	log.Printf("[DEBUG] 需要物理删除的题目ID: %v", toDeleteIDs)

	if len(toRestoreIDs) > 0 {
		if err := s.questionDAO.RestoreQuestionsByID(toRestoreIDs); err != nil {
			log.Printf("[ERROR] 恢复题目失败: %v", err)
			return fmt.Errorf("确认题目失败: %v", err)
		}
	}

	if len(toDeleteIDs) > 0 {
		if err := s.questionDAO.DeleteQuestionsPermanently(toDeleteIDs); err != nil {
			log.Printf("[ERROR] 物理删除题目失败: %v", err)
			return fmt.Errorf("清除未选题目失败: %v", err)
		}
	}

	log.Printf("[DEBUG] SaveSelectedQuestions 完成")
	return nil
}

// GetQuestionsByUserID 获取用户题目列表
func (s *QuestionService) GetQuestionsByUserID(userID int64, language, questionType, keyword string) ([]*model.Question, error) {
	questions, err := s.questionDAO.GetQuestionsByUserID(userID, language, questionType, keyword)
	if err != nil {
		return nil, err
	}

	// 将Options从JSON字符串转换为slice（可选，根据需要）
	for _, q := range questions {
		fmt.Println(q.ID)
		// 注意：这里没有转换，因为控制器可能需要原始JSON格式
		// 如需转换，可在此处添加json.Unmarshal
	}

	return questions, nil
}

// UpdateQuestion 更新题目
func (s *QuestionService) UpdateQuestion(question *model.Question) error {
	// 验证题目存在且属于当前用户
	existingQuestion, err := s.questionDAO.GetUndeletedQuestionByID(question.ID)
	if err != nil {
		return err
	}

	if existingQuestion.UserID != question.UserID {
		return fmt.Errorf("无权更新该题目")
	}

	// 验证题目类型
	if question.QuestionType != model.QuestionTypeSingle &&
		question.QuestionType != model.QuestionTypeMultiple {
		return fmt.Errorf("无效的题目类型: %s", question.QuestionType)
	}

	// 更新题目
	return s.questionDAO.UpdateQuestion(question)
}

// DeleteQuestion 软删除题目
func (s *QuestionService) DeleteQuestion(userID, questionID int64) error {
	// 验证题目存在且属于当前用户
	question, err := s.questionDAO.GetUndeletedQuestionByID(questionID)
	if err != nil {
		return err
	}

	if question.UserID != userID {
		return fmt.Errorf("无权删除该题目")
	}

	// // 软删除题目
	return s.questionDAO.DeleteQuestion(questionID)
	//物理永久删除
	// return s.questionDAO.PermanentDeleteQuestion(questionID)
}
