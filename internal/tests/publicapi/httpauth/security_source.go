package httpauth

import (
	"context"

	"github.com/walletera/accounts/publicapi"
)

type SecuritySource struct {
	token string
}

func NewSecuritySource(token string) *SecuritySource {
	return &SecuritySource{token: token}
}

func (s *SecuritySource) BearerAuth(ctx context.Context, operationName publicapi.OperationName) (publicapi.BearerAuth, error) {
	return publicapi.BearerAuth{Token: s.token}, nil
}
