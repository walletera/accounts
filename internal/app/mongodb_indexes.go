package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/walletera/accounts/pkg/logattr"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var indexes = map[string]mongo.IndexModel{
	"accounts_dinopay_account_number_unique": {
		Keys: bson.D{
			{Key: "accountDetails.oneof.dinopayaccountdetails.accountType", Value: 1},
			{Key: "accountDetails.oneof.dinopayaccountdetails.accountNumber", Value: 1},
		},
		Options: options.Index().
			SetUnique(true).
			SetName("accounts_dinopay_account_number_unique").
			SetPartialFilterExpression(bson.D{
				{Key: "accountDetails.oneof.dinopayaccountdetails.accountType", Value: bson.D{{Key: "$gt", Value: ""}}},
				{Key: "accountDetails.oneof.dinopayaccountdetails.accountNumber", Value: bson.D{{Key: "$gt", Value: ""}}},
			}),
	},
	"accounts_cvu_unique": {
		Keys: bson.D{
			{Key: "accountDetails.oneof.cvuaccountdetails.accountType", Value: 1},
			{Key: "accountDetails.oneof.cvuaccountdetails.routingInfo.oneof.cvucvuroutinginfo.cvu", Value: 1},
		},
		Options: options.Index().
			SetUnique(true).
			SetName("accounts_cvu_unique").
			SetPartialFilterExpression(bson.D{
				{Key: "accountDetails.oneof.cvuaccountdetails.accountType", Value: bson.D{{Key: "$gt", Value: ""}}},
				{Key: "accountDetails.oneof.cvuaccountdetails.routingInfo.oneof.cvucvuroutinginfo.cvu", Value: bson.D{{Key: "$gt", Value: ""}}},
			}),
	},
}

func createMongoDBUniqueIndexes(ctx context.Context, app *App) error {
	coll := app.mongoClient.Database(accountsDatabase).Collection(accountsCollection)

	for indexName, index := range indexes {
		exists, err := indexExistsByName(ctx, coll, indexName, app.logger)
		if err != nil {
			return fmt.Errorf("failed to check indexes: %w", err)
		}
		if exists {
			app.logger.With("index_name", indexName).Info("index already exists")
			continue
		}
		_, err = coll.Indexes().CreateOne(ctx, index)
		if err != nil {
			return fmt.Errorf("failed creating index: %w", err)
		}
		app.logger.With("index_name", indexName).Info("index created")
	}

	return nil
}

func indexExistsByName(ctx context.Context, coll *mongo.Collection, indexName string, logger *slog.Logger) (bool, error) {
	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		return false, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			logger.Error("error closing cursor", logattr.Error(err.Error()))
		}
	}(cursor, ctx)

	for cursor.Next(ctx) {
		var idx bson.M
		if err := cursor.Decode(&idx); err != nil {
			return false, err
		}
		if name, ok := idx["name"].(string); ok && name == indexName {
			return true, nil
		}
	}
	if err := cursor.Err(); err != nil {
		return false, err
	}
	return false, nil
}
