package mongodb

import (
	"context"

	"github.com/walletera/accounts/publicapi"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Iterator struct {
	cursor *mongo.Cursor
}

func (m *Iterator) Next() (bool, publicapi.Account, error) {
	if !m.cursor.Next(context.Background()) {
		if err := m.cursor.Err(); err != nil {
			return false, publicapi.Account{}, err
		}
		return false, publicapi.Account{}, nil
	}

	var accountBSON AccountBSON
	if err := m.cursor.Decode(&accountBSON); err != nil {
		return false, publicapi.Account{}, err
	}

	return true, publicapi.Account(accountBSON), nil
}
