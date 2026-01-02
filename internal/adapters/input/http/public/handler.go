package public

import (
	"context"
	"log/slog"

	"github.com/walletera/accounts/internal/domain/accounts"
	"github.com/walletera/accounts/pkg/logattr"
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
	result, err := h.repository.SearchAccounts(ctx, params)
	if err != nil {
		h.logger.Error(
			"failed listing accounts",
			logattr.Error(err.Error()),
		)
		return &publicapi.ApiError{
			ErrorMessage: err.Message(),
			ErrorCode:    err.Code().String(),
		}, nil
	}
	var accountList []publicapi.Account
	for {
		ok, account, iteratorErr := result.Iterator.Next()
		if iteratorErr != nil {
			h.logger.Error(
				"failed listing accounts",
				logattr.Error(iteratorErr.Error()),
			)
			return &publicapi.ApiError{
				ErrorMessage: iteratorErr.Error(),
			}, nil
		}
		if !ok {
			break
		}
		accountList = append(accountList, account)
	}
	resp := publicapi.ListAccountsOKApplicationJSON(accountList)
	return &resp, nil
}

func (h Handler) CreateAccount(ctx context.Context, req *publicapi.Account, _ publicapi.CreateAccountParams) (publicapi.CreateAccountRes, error) {
	werr := h.repository.SaveAccount(ctx, *req)
	if werr != nil {
		h.logger.Error(
			"failed saving payment",
			logattr.Error(werr.Message()),
		)
		return &publicapi.CreateAccountInternalServerError{
			ErrorMessage: werr.Message(),
			ErrorCode:    werr.Code().String(),
		}, nil
	}
	h.logger.With(logattr.AccountId(req.ID.String())).Info("account saved")
	return req, nil
}
