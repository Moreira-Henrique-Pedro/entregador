package response

import "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"

type User struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
}

func NewUser(user *entities.User) User {
	return User{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Role:   string(user.Role),
	}
}
