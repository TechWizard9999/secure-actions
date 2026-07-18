package mongo

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type secretDoc struct {
	ID         bson.ObjectID `bson:"_id"`
	Identifier string        `bson:"identifier"`
	Value      string        `bson:"value"`
}

// SecretRepository implements secrets.Manager backed by MongoDB.
type SecretRepository struct {
	col *mongo.Collection
}

func NewSecretRepository(ctx context.Context, client *Client) (*SecretRepository, error) {
	col := client.Collection(CollectionSecrets)

	_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "identifier", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, fmt.Errorf("create identifier index: %w", err)
	}

	return &SecretRepository{col: col}, nil
}

func (r *SecretRepository) Set(ctx context.Context, identifier, encryptedValue string) error {
	filter := bson.D{{Key: "identifier", Value: identifier}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "identifier", Value: identifier},
		{Key: "value", Value: encryptedValue},
	}}}

	_, err := r.col.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("set secret: %w", err)
	}
	return nil
}

func (r *SecretRepository) Get(ctx context.Context, identifier string) (string, bool, error) {
	var doc secretDoc
	err := r.col.FindOne(ctx, bson.D{{Key: "identifier", Value: identifier}}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("get secret: %w", err)
	}
	return doc.Value, true, nil
}

func (r *SecretRepository) Delete(ctx context.Context, identifier string) error {
	_, err := r.col.DeleteOne(ctx, bson.D{{Key: "identifier", Value: identifier}})
	if err != nil {
		return fmt.Errorf("delete secret: %w", err)
	}
	return nil
}

func (r *SecretRepository) Keys(ctx context.Context) ([]string, error) {
	cursor, err := r.col.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}
	defer cursor.Close(ctx)

	var keys []string
	for cursor.Next(ctx) {
		var doc secretDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode secret: %w", err)
		}
		keys = append(keys, doc.Identifier)
	}
	return keys, cursor.Err()
}
