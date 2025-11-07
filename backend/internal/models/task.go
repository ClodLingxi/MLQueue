package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusQueued    TaskStatus = "queued"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// JSONB type for PostgreSQL
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

type Task struct {
	ID           string     `json:"task_id" gorm:"primaryKey;type:varchar(100)"`
	Name         string     `json:"name" gorm:"type:varchar(255);not null"`
	Config       JSONB      `json:"config" gorm:"type:jsonb"`
	Priority     int        `json:"priority" gorm:"default:0;index"`
	Status       TaskStatus `json:"status" gorm:"type:varchar(20);index;default:'pending'"`
	Metadata     JSONB      `json:"metadata" gorm:"type:jsonb"`
	Result       JSONB      `json:"result" gorm:"type:jsonb"`
	ErrorMessage string     `json:"error_message" gorm:"type:text"`
	CreatedAt    time.Time  `json:"created_at" gorm:"index"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	UserID       string     `json:"user_id" gorm:"type:varchar(100);index"`
	UpdatedAt    time.Time  `json:"-"`
}

type ConfigTemplate struct {
	ID          string    `json:"template_id" gorm:"primaryKey;type:varchar(100)"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null;uniqueIndex"`
	Config      JSONB     `json:"config" gorm:"type:jsonb"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UserID      string    `json:"user_id" gorm:"type:varchar(100);index"`
}

type Test struct {
	DD string `json:"dd"`
}

type User struct {
	ID        string    `json:"user_id" gorm:"primaryKey;type:varchar(100)"`
	Email     string    `json:"email" gorm:"uniqueIndex;type:varchar(255)"`
	APIKey    string    `json:"api_key" gorm:"uniqueIndex;type:varchar(100)"`
	Tier      string    `json:"tier" gorm:"type:varchar(20);default:'standard'"` // standard, premium
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"-"`
}

type WebhookConfig struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"type:varchar(100);index"`
	URL       string    `json:"url" gorm:"type:varchar(500)"`
	Events    JSONB     `json:"events" gorm:"type:jsonb"` // Array of event types
	Active    bool      `json:"active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
}

// AutoMigrate creates tables
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Task{},
		&ConfigTemplate{},
		&User{},
		&WebhookConfig{},
	)
}
