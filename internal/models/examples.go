package models

import (
	"time"
)

// ExampleUser 演示struct tags的使用
type ExampleUser struct {
	// db:"id" 表示这个字段对应数据库中的 id 列
	// json:"id" 表示JSON序列化时使用 "id" 作为字段名
	ID int `json:"id" db:"id"`

	// binding:"required,email" 表示这个字段在Gin验证时必须存在且必须是邮箱格式
	// db:"email" 表示对应数据库中的 email 列
	Email string `json:"email" db:"email" binding:"required,email"`

	// binding:"required,min=2,max=50" 表示必填，长度在2-50之间
	Name string `json:"name" db:"name" binding:"required,min=2,max=50"`

	// binding:"min=18,max=120" 表示年龄在18-120之间
	Age int `json:"age" db:"age" binding:"min=18,max=120"`

	// json:"-" 表示这个字段不会出现在JSON响应中（敏感信息）
	// binding:"required,min=6" 表示密码必填且至少6位
	Password string `json:"-" db:"password_hash" binding:"required,min=6"`

	// json:"created_at" 时间字段的JSON序列化
	// db:"created_at" 对应数据库时间戳字段
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// json:"profile,omitempty" 如果Profile为nil，则JSON中不包含此字段
	Profile *UserProfile `json:"profile,omitempty" db:"-"`
}

// UserProfile 用户资料（嵌套结构体）
type UserProfile struct {
	Bio     string `json:"bio" db:"bio"`
	Website string `json:"website" db:"website" binding:"omitempty,url"` // 可选，但如果提供必须是URL格式
}

// LoginRequest 登录请求示例
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`    // 必填邮箱
	Password string `json:"password" binding:"required,min=6"` // 必填密码，至少6位
}

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	Name    *string `json:"name" binding:"omitempty,min=2,max=50"`  // 可选，但如果提供则长度2-50
	Age     *int    `json:"age" binding:"omitempty,min=18,max=120"` // 可选，但如果提供则18-120
	Bio     *string `json:"bio" binding:"omitempty,max=500"`        // 可选，最多500字符
	Website *string `json:"website" binding:"omitempty,url"`        // 可选，但必须是URL格式
}
