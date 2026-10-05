package request

import "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"

type CreateUser struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin doorman"`
}

func (u *CreateUser) FromDTO() (*entities.User, string) {
	return &entities.User{
		Email: u.Email,
		Name:  u.Name,
		Role:  entities.Role(u.Role),
	}, u.Password
}
