package mongodb

import (
    "context"

    "github.com/google/uuid"
    "github.com/walletera/accounts/internal/domain/accounts"
    "github.com/walletera/accounts/publicapi"
    "github.com/walletera/werrors"
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

type AccountBSON struct {
    // Walletera account unique identifier.
    ID uuid.UUID `bson:"_id"`
    // Walletera unique identifier of the customer (institution) that owns the account.
    CustomerId uuid.UUID `bson:"customerId"`
    // Identifier assigned to the account by the customer.
    CustomerAccountId publicapi.OptString `bson:"customerAccountId"`
    InstitutionName   publicapi.OptString `bson:"institutionName"`
    InstitutionId     publicapi.OptString `bson:"institutionId"`
    Currency          publicapi.Currency  `bson:"currency"`
    // Extra account details. The details depend on the accountType.
    AccountDetails publicapi.AccountAccountDetails `bson:"accountDetails"`
}

type AccountsRepository struct {
    client         *mongo.Client
    dbName         string
    collectionName string
}

func NewAccountsRepository(client *mongo.Client, dbName string, collectionName string) *AccountsRepository {
    return &AccountsRepository{client: client, dbName: dbName, collectionName: collectionName}
}

func (a AccountsRepository) GetAccount(ctx context.Context, id uuid.UUID) (publicapi.Account, werrors.WError) {
    //TODO implement me
    panic("implement me")
}

func (a AccountsRepository) SaveAccount(ctx context.Context, account publicapi.Account) werrors.WError {
    accountBSON := AccountBSON(account)
    coll := a.client.Database(a.dbName).Collection(a.collectionName)
    _, err := coll.InsertOne(ctx, accountBSON)
    if err != nil {
        if mongo.IsDuplicateKeyError(err) {
            return werrors.NewNonRetryableInternalError("duplicate key error: %s", err.Error())
        }
        return werrors.NewRetryableInternalError("failed to save account: %s", err.Error())
    }
    return nil
}

func (a AccountsRepository) SearchAccounts(ctx context.Context, listAccountsParams publicapi.ListAccountsParams) (accounts.QueryResult, werrors.WError) {
    filter := bson.M{}

    if listAccountsParams.ID.IsSet() {
        filter["_id"] = listAccountsParams.ID.Value
    }

    if listAccountsParams.ID.IsSet() {
        filter["accountDetails.cvu"] = listAccountsParams.Cvu.Value
    }

    if listAccountsParams.ID.IsSet() {
        filter["accountDetails.account_number"] = listAccountsParams.DinoPayAccountNumber.Value
    }

    coll := a.client.Database(a.dbName).Collection(a.collectionName)

    total, err := coll.CountDocuments(ctx, filter)
    if err != nil {
        return accounts.QueryResult{}, werrors.NewRetryableInternalError("failed to count accounts: %s", err.Error())
    }

    findOpts := options.Find()

    cursor, err := coll.Find(ctx, filter, findOpts)
    if err != nil {
        return accounts.QueryResult{}, werrors.NewRetryableInternalError("failed to find accounts: %s", err.Error())
    }

    iterator := &Iterator{cursor: cursor}
    return accounts.QueryResult{
        Iterator: iterator,
        Total:    uint64(total),
    }, nil
}
