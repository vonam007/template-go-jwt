package models

type User struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Role string `json:"role,omitempty"`
}
