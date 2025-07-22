package dao

import (
	"examsystem/dao/model"

	"gorm.io/gorm"
)

type QuestionDAO struct {
	DB *gorm.DB
}

func NewQuestionDAO(db *gorm.DB) *QuestionDAO {
	return &QuestionDAO{DB: db}
}

// CreateQuestion 创建题目
func (dao *QuestionDAO) CreateQuestion(question *model.Question) error {
	return dao.DB.Create(question).Error
}

// GetQuestionByID 获取题目（包含已删除的）
func (dao *QuestionDAO) GetQuestionByID(id int64) (*model.Question, error) {
	var question model.Question
	err := dao.DB.Unscoped().First(&question, id).Error
	return &question, err
}

// GetUndeletedQuestionByID 获取未删除的题目
func (dao *QuestionDAO) GetUndeletedQuestionByID(id int64) (*model.Question, error) {
	var question model.Question
	err := dao.DB.First(&question, id).Error
	return &question, err
}

// UpdateQuestion 更新题目
func (dao *QuestionDAO) UpdateQuestion(question *model.Question) error {
	return dao.DB.Save(question).Error
}

// DeleteQuestion 软删除题目
func (dao *QuestionDAO) DeleteQuestion(id int64) error {
	return dao.DB.Delete(&model.Question{}, id).Error
}

// PermanentDeleteQuestion 永久删除题目
func (dao *QuestionDAO) PermanentDeleteQuestion(id int64) error {
	return dao.DB.Unscoped().Delete(&model.Question{}, id).Error
}

// GetQuestionsByUserID 获取用户题目列表（未删除的）
func (dao *QuestionDAO) GetQuestionsByUserID(userID int64, language, questionType, keyword string) ([]*model.Question, error) {
	var questions []*model.Question
	query := dao.DB.Where("user_id = ?", userID)

	if language != "" {
		query = query.Where("language = ?", language)
	}

	if questionType != "" {
		query = query.Where("question_type = ?", questionType)
	}

	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}

	err := query.Order("created_at DESC").Find(&questions).Error
	return questions, err
}

// GetGeneratedQuestionsByUserID 获取用户逻辑删除状态的题目（未确认）
func (dao *QuestionDAO) GetGeneratedQuestionsByUserID(userID int64) ([]*model.Question, error) {
	var questions []*model.Question
	err := dao.DB.Unscoped().
		Where("user_id = ? AND deleted_at IS NOT NULL", userID).
		Find(&questions).Error
	return questions, err
}

// RestoreQuestionsByID 取消逻辑删除（恢复题目）
func (dao *QuestionDAO) RestoreQuestionsByID(ids []int64) error {
	return dao.DB.Unscoped().
		Model(&model.Question{}).
		Where("id IN ?", ids).
		Update("deleted_at", nil).Error
}

// DeleteQuestionsPermanently 物理删除
func (dao *QuestionDAO) DeleteQuestionsPermanently(ids []int64) error {
	return dao.DB.Unscoped().
		Where("id IN ?", ids).
		Delete(&model.Question{}).Error
}

func (dao *QuestionDAO) BatchCreateQuestions(questions []*model.Question) error {
	return dao.DB.Create(&questions).Error
}
