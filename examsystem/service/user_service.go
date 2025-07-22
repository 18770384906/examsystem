package service

import (
	"examsystem/dao"
	"examsystem/dao/model"
)

// UserService 用户服务
type UserService struct {
	userDAO *dao.UserDAO
}

// NewUserService 创建用户服务实例
func NewUserService(userDAO *dao.UserDAO) *UserService {
	return &UserService{
		userDAO: userDAO,
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(user *model.User) error {
	return s.userDAO.Create(user)
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id int64) (*model.User, error) {
	return s.userDAO.GetByID(id)
}

// GetUserByUsername 根据名称获取用户
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	return s.userDAO.GetByUsername(username)
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(user *model.User) error {
	return s.userDAO.Update(user)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id int64) error {
	return s.userDAO.Delete(id)
}

// GetUserList 获取用户列表（支持分页）
func (s *UserService) GetUserList(page, pageSize int) ([]*model.User, int64, error) {
	return s.userDAO.GetList(page, pageSize)
}

// Login 用户登录
func (s *UserService) Login(username, password string) (*model.User, error) {
	// 验证登录信息
	user, err := s.userDAO.ValidateLogin(username, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ResetPassword 重置密码
func (s *UserService) ResetPassword(username, newPassword string) error {
	return s.userDAO.ResetPassword(username, newPassword)
}
