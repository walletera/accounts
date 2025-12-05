package mongodb

import (
    "context"

    "github.com/google/uuid"
    "github.com/walletera/accounts/internal/domain/accounts"
    "github.com/walletera/accounts/publicapi"
    "github.com/walletera/werrors"
    "go.mongodb.org/mongo-driver/v2/mongo"
)

type AccountBSON struct {
    ID               uuid.UUID         `bson:"_id"`
    AggregateVersion uint64            `bson:"version"`
    Data             publicapi.Account `bson:"data"`
}

type AccountsRepository struct {
    client         *mongo.Client
    dbName         string
    collectionName string
}

func NewAccountsRepository(client *mongo.Client, dbName string, collectionName string) *AccountsRepository {
    return &AccountsRepository{client: client, dbName: dbName, collectionName: collectionName}
}

func (a AccountsRepository) GetAccount(ctx context.Context, id uuid.UUID) (accounts.Account, werrors.WError) {
    //TODO implement me
    panic("implement me")
}

func (a AccountsRepository) SaveAccount(ctx context.Context, payment accounts.Account) werrors.WError {
    //TODO implement me
    panic("implement me")
}

func (a AccountsRepository) SearchAccounts(ctx context.Context, listAccountsParams publicapi.ListAccountsParams) (accounts.QueryResult, werrors.WError) {
    //TODO implement me
    panic("implement me")
}
