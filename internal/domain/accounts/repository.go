package accounts

import (
    "context"

    "github.com/google/uuid"
    "github.com/walletera/accounts/publicapi"
    "github.com/walletera/werrors"
)

type Account struct {
    ID               uuid.UUID
    AggregateVersion uint64
    Data             publicapi.Account
}

type Iterator interface {
    Next() (bool, Account, error)
}

type QueryResult struct {
    Iterator Iterator
    Total    uint64
}

type Repository interface {
    GetAccount(ctx context.Context, id uuid.UUID) (Account, werrors.WError)
    SaveAccount(ctx context.Context, payment Account) werrors.WError
    SearchAccounts(ctx context.Context, listAccountsParams publicapi.ListAccountsParams) (QueryResult, werrors.WError)
}
