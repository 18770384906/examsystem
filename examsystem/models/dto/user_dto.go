package dto

import "time"

// 登录请求
type LoginRequest struct {
	Username     string `json:"username" binding:"required"`
	PasswordHash string `json:"password_hash" binding:"required,min=6"`
}

// 注册请求
type RegisterRequest struct {
	PasswordHash string `json:"password_hash" binding:"required,min=6"`
	Role         string `json:"role,omitempty"`
	Username     string `json:"username" binding:"required"`
}

// 用户更新请求
type UserUpdateRequest struct {
	PasswordHash string `json:"password_hash,omitempty"` // 密码可选
	Role         string `json:"role,omitempty"`
	Username     string `json:"username" binding:"required"`
	Status       int32  `json:"status,omitempty"`
}

// 登录响应
type LoginResponse struct {
	TokenType string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"` // 过期时间，单位：秒
}

// 用户响应
type UserResponse struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username"`
	Role      string     `json:"role"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// 用户列表响应
type UserListResponse struct {
	List  []*UserResponse `json:"list"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}
