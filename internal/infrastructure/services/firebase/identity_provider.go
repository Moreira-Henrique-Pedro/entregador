package firebase

import (
	"context"
	"fmt"

	"firebase.google.com/go/v4/auth"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

const roleClaim = "role"

type authClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
	CreateUser(ctx context.Context, user *auth.UserToCreate) (*auth.UserRecord, error)
	SetCustomUserClaims(ctx context.Context, uid string, customClaims map[string]interface{}) error
	DeleteUser(ctx context.Context, uid string) error
}

type IdentityProvider struct {
	client authClient
}

func NewIdentityProvider(client authClient) *IdentityProvider {
	return &IdentityProvider{client: client}
}

func (p *IdentityProvider) VerifyToken(ctx context.Context, token string) (*entities.User, error) {
	verified, err := p.client.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", entities.ErrUnauthenticated, err)
	}
	return userFromToken(verified), nil
}

func (p *IdentityProvider) CreateUser(ctx context.Context, user *entities.User, password string) (*entities.User, error) {
	record, err := p.client.CreateUser(ctx, (&auth.UserToCreate{}).
		Email(user.Email).
		Password(password).
		DisplayName(user.Name))
	if auth.IsEmailAlreadyExists(err) {
		return nil, fmt.Errorf("email %s: %w", user.Email, entities.ErrUserAlreadyExists)
	}
	if err != nil {
		return nil, fmt.Errorf("create firebase user: %w", err)
	}

	if err := p.assignRole(ctx, record.UID, user.Role); err != nil {
		return nil, err
	}

	return &entities.User{ID: record.UID, Email: record.Email, Name: record.DisplayName, Role: user.Role}, nil
}

func (p *IdentityProvider) assignRole(ctx context.Context, uid string, role entities.Role) error {
	err := p.client.SetCustomUserClaims(ctx, uid, map[string]interface{}{roleClaim: string(role)})
	if err == nil {
		return nil
	}
	if deleteErr := p.client.DeleteUser(ctx, uid); deleteErr != nil {
		return fmt.Errorf("set role: %w (and failed to delete the user without role: %v)", err, deleteErr)
	}
	return fmt.Errorf("set role: %w", err)
}

func userFromToken(token *auth.Token) *entities.User {
	return &entities.User{
		ID:    token.UID,
		Email: stringClaim(token, "email"),
		Name:  stringClaim(token, "name"),
		Role:  entities.Role(stringClaim(token, roleClaim)),
	}
}

func stringClaim(token *auth.Token, name string) string {
	value, _ := token.Claims[name].(string)
	return value
}
