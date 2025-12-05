package public

import (
    "context"
    "log/slog"

    "github.com/walletera/accounts/internal/domain/accounts"
    "github.com/walletera/accounts/publicapi"
)

type Handler struct {
    repository accounts.Repository
    logger     *slog.Logger
}

var _ publicapi.Handler = (*Handler)(nil)

func NewHandler(repository accounts.Repository, logger *slog.Logger) *Handler {
    return &Handler{repository: repository, logger: logger}
}

func (h Handler) ListAccounts(ctx context.Context, params publicapi.ListAccountsParams) (publicapi.ListAccountsRes, error) {
    //TODO implement me
    panic("implement me")
}

func (h Handler) PostAccount(ctx context.Context, req *publicapi.Account, params publicapi.PostAccountParams) (publicapi.PostAccountRes, error) {
    //TODO implement me
    panic("implement me")
}
