package mongodb

import (
    "context"

    "github.com/walletera/accounts/internal/domain/accounts"
    "go.mongodb.org/mongo-driver/v2/mongo"
)

type Iterator struct {
    cursor *mongo.Cursor
}

func (m *Iterator) Next() (bool, accounts.Account, error) {
    if !m.cursor.Next(context.Background()) {
        if err := m.cursor.Err(); err != nil {
            return false, accounts.Account{}, err
        }
        return false, accounts.Account{}, nil
    }

    var paymentBSON AccountBSON
    if err := m.cursor.Decode(&paymentBSON); err != nil {
        return false, accounts.Account{}, err
    }

    payment := accounts.Account{
        ID:               paymentBSON.ID,
        AggregateVersion: paymentBSON.AggregateVersion,
        Data:             paymentBSON.Data,
    }
    return true, payment, nil
}
