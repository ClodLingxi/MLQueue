package models

import (
	"time"
)

// Group 代表一个ML项目组
type Group struct {
	ID          string    `json:"group_id" gorm:"primaryKey;type:varchar(100)"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	UserID      string    `json:"user_id" gorm:"type:varchar(100);index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系 - 一个Group包含多个TrainingUnit
	TrainingUnits []TrainingUnit `json:"-" gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
}

// TrainingUnit 训练单元，云端和Python环境各保留一份
type TrainingUnit struct {
	ID          string `json:"unit_id" gorm:"primaryKey;type:varchar(100)"`
	GroupID     string `json:"group_id" gorm:"type:varchar(100);index"`
	Name        string `json:"name" gorm:"type:varchar(255);not null"`
	Description string `json:"description" gorm:"type:text"`

	// 基础配置
	Config JSONB `json:"config" gorm:"type:jsonb"`

	// 同步版本控制
	Version int `json:"version" gorm:"default:1"` // 每次修改递增

	// 状态
	Status string `json:"status" gorm:"type:varchar(20);default:'idle'"` // idle/running/completed

	// Python客户端连接状态
	ConnectionStatus string     `json:"connection_status" gorm:"type:varchar(20);default:'disconnected'"` // connected/disconnected
	LastHeartbeat    *time.Time `json:"last_heartbeat" gorm:"type:timestamp"`                             // 最后心跳时间

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	UserID string `json:"user_id" gorm:"type:varchar(100);index"`

	// 关联关系 - 一个TrainingUnit包含多个TrainingQueue
	TrainingQueues []TrainingQueue `json:"-" gorm:"foreignKey:UnitID;constraint:OnDelete:CASCADE"`
}

// TrainingQueue 训练队列
type TrainingQueue struct {
	ID     string `json:"queue_id" gorm:"primaryKey;type:varchar(100)"`
	UnitID string `json:"unit_id" gorm:"type:varchar(100);index"`
	Name   string `json:"name" gorm:"type:varchar(255);not null"`

	// 训练参数（由Python环境定义，前端可修改）
	Parameters JSONB `json:"parameters" gorm:"type:jsonb"`

	// 队列顺序（自动分配，可通过API调整）
	// 数字越小越靠前执行
	Order int `json:"order" gorm:"not null;index"`

	// 执行状态（由Python客户端控制）
	Status string `json:"status" gorm:"type:varchar(20);default:'pending';index"`
	// pending: 等待执行
	// running: Python正在执行
	// completed: 执行完成
	// failed: 执行失败
	// cancelled: 已取消

	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`

	// 训练结果
	Result   JSONB  `json:"result" gorm:"type:jsonb"`  // 训练结果
	Metrics  JSONB  `json:"metrics" gorm:"type:jsonb"` // 训练指标
	ErrorMsg string `json:"error_msg" gorm:"type:text"`

	// 元数据
	CreatedBy string    `json:"created_by" gorm:"type:varchar(20)"` // 'client' or 'web'
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	UserID string `json:"user_id" gorm:"type:varchar(100);index"`
}

// AutoMigrateV2 creates new tables
func AutoMigrateV2(db interface{ AutoMigrate(...interface{}) error }) error {
	return db.AutoMigrate(
		&Group{},
		&TrainingUnit{},
		&TrainingQueue{},
	)
}
