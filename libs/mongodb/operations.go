package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertOne inserts a single document into the collection
func (c *Client) InsertOne(ctx context.Context, collection string, document interface{}) (*mongo.InsertOneResult, error) {
	if c.client == nil {
		return nil, ErrClientClosed
	}
	return c.Collection(collection).InsertOne(ctx, document)
}

// InsertMany inserts multiple documents into the collection
func (c *Client) InsertMany(ctx context.Context, collection string, documents []interface{}) (*mongo.InsertManyResult, error) {
	if c.client == nil {
		return nil, ErrClientClosed
	}
	return c.Collection(collection).InsertMany(ctx, documents)
}

// FindOne finds a single document matching the filter
func (c *Client) FindOne(ctx context.Context, collection string, filter interface{}, result interface{}) error {
	if c.client == nil {
		return ErrClientClosed
	}
	return c.Collection(collection).FindOne(ctx, filter).Decode(result)
}

// FindMany finds all documents matching the filter
func (c *Client) FindMany(ctx context.Context, collection string, filter interface{}) (*mongo.Cursor, error) {
	if c.client == nil {
		return nil, ErrClientClosed
	}
	return c.Collection(collection).Find(ctx, filter)
}

// UpdateOne updates a single document matching the filter
func (c *Client) UpdateOne(ctx context.Context, collection string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	if c.client == nil {
		return nil, ErrClientClosed
	}
	return c.Collection(collection).UpdateOne(ctx, filter, update)
}

// UpdateMany updates all documents matching the filter
func (c *Client) UpdateMany(ctx context.Context, collection string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	if c.client == nil {
		return nil, ErrClientClosed
	}
	return c.Collection(collection).UpdateMany(ctx, filter, update)
}

// DeleteOne deletes a single document matching the filter
func (c *Client) DeleteOne(ctx context.Context, collection string, filter interface{}) (*mongo.DeleteResult, error) {
	if c.client == nil {
		return nil, ErrClientClosed
	}
	return c.Collection(collection).DeleteOne(ctx, filter)
}

// DeleteMany deletes all documents matching the filter
func (c *Client) DeleteMany(ctx context.Context, collection string, filter interface{}) (*mongo.DeleteResult, error) {
	if c.client == nil {
		return nil, ErrClientClosed
	}
	return c.Collection(collection).DeleteMany(ctx, filter)
}

// CountDocuments counts the number of documents matching the filter
func (c *Client) CountDocuments(ctx context.Context, collection string, filter interface{}) (int64, error) {
	if c.client == nil {
		return 0, ErrClientClosed
	}
	return c.Collection(collection).CountDocuments(ctx, filter)
}

// Aggregate performs an aggregation pipeline
func (c *Client) Aggregate(ctx context.Context, collection string, pipeline interface{}) (*mongo.Cursor, error) {
	if c.client == nil {
		return nil, ErrClientClosed
	}
	return c.Collection(collection).Aggregate(ctx, pipeline)
}

// Distinct finds the distinct values for a specified field
func (c *Client) Distinct(ctx context.Context, collection string, fieldName string, filter interface{}) ([]interface{}, error) {
	if c.client == nil {
		return nil, ErrClientClosed
	}
	return c.Collection(collection).Distinct(ctx, fieldName, filter)
}

// CreateIndex creates an index for the collection
func (c *Client) CreateIndex(ctx context.Context, collection string, keys interface{}, opts ...*options.IndexOptions) (string, error) {
	if c.client == nil {
		return "", ErrClientClosed
	}
	model := mongo.IndexModel{
		Keys:    keys,
		Options: opts[0],
	}
	return c.Collection(collection).Indexes().CreateOne(ctx, model)
}

// DropCollection drops the specified collection
func (c *Client) DropCollection(ctx context.Context, collection string) error {
	if c.client == nil {
		return ErrClientClosed
	}
	return c.Collection(collection).Drop(ctx)
}

// Transaction executes the provided function within a transaction
func (c *Client) Transaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	if c.client == nil {
		return nil, ErrClientClosed
	}

	session, err := c.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	return session.WithTransaction(ctx, fn)
}

// BulkWrite performs a bulk write operation
func (c *Client) BulkWrite(ctx context.Context, collection string, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	if c.client == nil {
		return nil, ErrClientClosed
	}
	return c.Collection(collection).BulkWrite(ctx, models, opts...)
}

// Watch creates a change stream for the specified collection
func (c *Client) Watch(ctx context.Context, collection string, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	if c.client == nil {
		return nil, ErrClientClosed
	}
	return c.Collection(collection).Watch(ctx, pipeline, opts...)
}

// FindOneAndUpdate finds a single document and updates it
func (c *Client) FindOneAndUpdate(ctx context.Context, collection string, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	if c.client == nil {
		return c.Collection(collection).FindOne(ctx, filter) // This will be a no-op since client is nil
	}
	return c.Collection(collection).FindOneAndUpdate(ctx, filter, update, opts...)
}

// FindOneAndDelete finds a single document and deletes it
func (c *Client) FindOneAndDelete(ctx context.Context, collection string, filter interface{}, opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult {
	if c.client == nil {
		return c.Collection(collection).FindOne(ctx, filter) // This will be a no-op since client is nil
	}
	return c.Collection(collection).FindOneAndDelete(ctx, filter, opts...)
}

// FindOneAndReplace finds a single document and replaces it
func (c *Client) FindOneAndReplace(ctx context.Context, collection string, filter interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *mongo.SingleResult {
	if c.client == nil {
		return c.Collection(collection).FindOne(ctx, filter) // This will be a no-op since client is nil
	}
	return c.Collection(collection).FindOneAndReplace(ctx, filter, replacement, opts...)
}
