package models

type User struct {
	ID   string `gorm:"primaryKey" json:"id"`
	Name string `json:"name,omitempty"`
	Role string `json:"role,omitempty"`
}
