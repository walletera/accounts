package accounts

import (
    "context"

    "github.com/google/uuid"
    "github.com/walletera/accounts/publicapi"
    "github.com/walletera/werrors"
)

type Iterator interface {
    Next() (bool, publicapi.Account, error)
}

type QueryResult struct {
    Iterator Iterator
    Total    uint64
}

type Repository interface {
    GetAccount(ctx context.Context, id uuid.UUID) (publicapi.Account, werrors.WError)
    SaveAccount(ctx context.Context, payment publicapi.Account) werrors.WError
    SearchAccounts(ctx context.Context, listAccountsParams publicapi.ListAccountsParams) (QueryResult, werrors.WError)
}
