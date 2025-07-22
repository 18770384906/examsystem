package dao

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"examsystem/dao/model"
)

// UserDAO 用户数据访问对象
type UserDAO struct {
	DB *gorm.DB
}

// NewUserDAO 创建用户DAO实例
func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{DB: db}
}

// Create 创建用户
func (dao *UserDAO) Create(user *model.User) error {
	// 密码加密
	log.Println("创建用户密码：", user.PasswordHash)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)

	return dao.DB.Create(user).Error
}

// GetByID 根据ID获取用户
func (dao *UserDAO) GetByID(id int64) (*model.User, error) {
	var user model.User
	err := dao.DB.First(&user, id).Error
	return &user, err
}

// GetByUsername 根据用户名称获取用户
func (dao *UserDAO) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := dao.DB.Where("username = ?", username).First(&user).Error
	return &user, err
}

// Update 更新用户
func (dao *UserDAO) Update(user *model.User) error {
	// 准备待更新的字段
	updates := map[string]interface{}{
		"role":     user.Role,
		"username": user.Username,
	}

	// 如果密码字段不为空，则更新密码
	if len(user.PasswordHash) > 0 {
		updates["password_hash"] = user.PasswordHash
	}

	// 使用Updates而不是Save，只更新指定字段
	return dao.DB.Model(user).Updates(updates).Error
}

// Delete 删除用户
func (dao *UserDAO) Delete(id int64) error {
	return dao.DB.Delete(&model.User{}, id).Error
}

// GetList 获取用户列表（支持分页）
func (dao *UserDAO) GetList(page, pageSize int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	// 查询总数
	err := dao.DB.Model(&model.User{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取数据列表
	offset := (page - 1) * pageSize
	err = dao.DB.Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

// ValidateLogin 验证用户登录
func (dao *UserDAO) ValidateLogin(username, password_hash string) (*model.User, error) {
	// 查找用户
	user, err := dao.GetByUsername(username)
	if err != nil {
		log.Println("查找用户错误：", err)
		return nil, err
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password_hash))
	if err != nil {
		log.Println("密码错误：", err)
		return nil, err
	}

	return user, nil
}

// ResetPassword 重置用户密码
func (dao *UserDAO) ResetPassword(username, newPassword string) error {
	user, err := dao.GetByUsername(username)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	return dao.DB.Save(user).Error
}
