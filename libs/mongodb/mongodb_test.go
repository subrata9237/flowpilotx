package mongodb

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const testCollectionName = "test_collection"

// TestDocument represents a test document structure
type TestDocument struct {
	Name  string    `bson:"name"`
	Value int       `bson:"value"`
	Time  time.Time `bson:"time"`
}

func setupTestEnvironment(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("FLOWPILOT_ENV", "development")
	os.Setenv("FLOWPILOT_MONGODB_URI", "mongodb://localhost:27017")
	os.Setenv("FLOWPILOT_MONGODB_DATABASE", "test")
	os.Setenv("FLOWPILOT_MONGODB_USERNAME", "")
	os.Setenv("FLOWPILOT_MONGODB_PASSWORD", "")
}

func cleanupTestEnvironment(t *testing.T) {
	// Clean up environment variables
	os.Unsetenv("FLOWPILOT_ENV")
	os.Unsetenv("FLOWPILOT_MONGODB_URI")
	os.Unsetenv("FLOWPILOT_MONGODB_DATABASE")
	os.Unsetenv("FLOWPILOT_MONGODB_USERNAME")
	os.Unsetenv("FLOWPILOT_MONGODB_PASSWORD")
}

// setupTestClient creates a new test client
func setupTestClient(t *testing.T) *Client {
	setupTestEnvironment(t)
	client, err := NewClient(nil) // Use environment variables
	require.NoError(t, err)
	require.NotNil(t, client)
	return client
}

// createTestCollection creates a test collection
func createTestCollection(t *testing.T, client *Client) {
	ctx := context.Background()
	coll := client.Collection(testCollectionName)
	_, err := coll.InsertOne(ctx, bson.M{"_id": "test"})
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		require.NoError(t, err)
	}
}

// cleanupTestData drops the test collection
func cleanupTestData(t *testing.T, client *Client) {
	ctx := context.Background()
	err := client.Collection(testCollectionName).Drop(ctx)
	require.NoError(t, err)
}

// insertTestDocuments inserts test documents
func insertTestDocuments(t *testing.T, client *Client, count int) {
	docs := make([]interface{}, count)
	for i := 0; i < count; i++ {
		docs[i] = TestDocument{
			Name:  fmt.Sprintf("test%d", i+1),
			Value: i + 1,
			Time:  time.Now(),
		}
	}

	coll := client.Collection(testCollectionName)
	_, err := coll.InsertMany(context.Background(), docs)
	require.NoError(t, err)
}

// assertDocumentCount checks if the collection has the expected number of documents
func assertDocumentCount(t *testing.T, client *Client, expected int64) {
	coll := client.Collection(testCollectionName)
	count, err := coll.CountDocuments(context.Background(), bson.M{})
	require.NoError(t, err)
	require.Equal(t, expected, count)
}

// assertDocumentNotExists checks if a document does not exist
func assertDocumentNotExists(t *testing.T, client *Client, filter bson.M) {
	coll := client.Collection(testCollectionName)
	var doc TestDocument
	err := coll.FindOne(context.Background(), filter).Decode(&doc)
	require.Equal(t, mongo.ErrNoDocuments, err)
}

// compareDocuments compares two test documents
func compareDocuments(t *testing.T, expected, actual TestDocument) {
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Value, actual.Value)
	require.WithinDuration(t, expected.Time, actual.Time, time.Second)
}

func TestNewClient(t *testing.T) {
	setupTestEnvironment(t)
	defer cleanupTestEnvironment(t)

	client, err := NewClient(nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	// Test connection
	err = client.Ping(context.Background())
	require.NoError(t, err)
}

func TestClient_Collection(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	defer cleanupTestEnvironment(t)

	coll := client.Collection(testCollectionName)
	require.NotNil(t, coll)

	// Test basic operations
	doc := bson.M{"test": "value"}
	_, err := coll.InsertOne(context.Background(), doc)
	require.NoError(t, err)

	var result bson.M
	err = coll.FindOne(context.Background(), bson.M{"test": "value"}).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "value", result["test"])

	// Cleanup
	err = coll.Drop(context.Background())
	require.NoError(t, err)
}

func TestClient_Database(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	defer cleanupTestEnvironment(t)

	db := client.Database()
	require.NotNil(t, db)
	assert.Equal(t, "test", db.Name())
}

func TestClient_CRUD(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	defer cleanupTestEnvironment(t)

	// Start fresh
	cleanupTestData(t, client)
	createTestCollection(t, client)
	defer cleanupTestData(t, client)

	// Test Insert
	insertTestDocuments(t, client, 3)

	// Test Find
	var doc TestDocument
	coll := client.Collection(testCollectionName)
	err := coll.FindOne(context.Background(), bson.M{"name": "test1"}).Decode(&doc)
	require.NoError(t, err)
	assert.Equal(t, "test1", doc.Name)
	assert.Equal(t, 1, doc.Value)

	// Test Update
	update := bson.M{"$set": bson.M{"value": 100}}
	_, err = coll.UpdateOne(context.Background(), bson.M{"name": "test1"}, update)
	require.NoError(t, err)

	err = coll.FindOne(context.Background(), bson.M{"name": "test1"}).Decode(&doc)
	require.NoError(t, err)
	assert.Equal(t, 100, doc.Value)

	// Test Delete
	_, err = coll.DeleteOne(context.Background(), bson.M{"name": "test1"})
	require.NoError(t, err)

	err = coll.FindOne(context.Background(), bson.M{"name": "test1"}).Decode(&doc)
	assert.Equal(t, mongo.ErrNoDocuments, err)
}

func TestClient_BulkOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	defer cleanupTestEnvironment(t)

	// Start fresh
	cleanupTestData(t, client)
	createTestCollection(t, client)
	defer cleanupTestData(t, client)

	// Test bulk insert
	docs := []interface{}{
		TestDocument{Name: "bulk1", Value: 1, Time: time.Now()},
		TestDocument{Name: "bulk2", Value: 2, Time: time.Now()},
		TestDocument{Name: "bulk3", Value: 3, Time: time.Now()},
		TestDocument{Name: "bulk4", Value: 4, Time: time.Now()},
	}

	coll := client.Collection(testCollectionName)
	result, err := coll.InsertMany(context.Background(), docs)
	require.NoError(t, err)
	require.Len(t, result.InsertedIDs, 4)

	// Test bulk update
	update := bson.M{"$inc": bson.M{"value": 10}}
	filter := bson.M{"name": bson.M{"$regex": "^bulk"}}
	updateResult, err := coll.UpdateMany(context.Background(), filter, update)
	require.NoError(t, err)
	require.Equal(t, int64(4), updateResult.ModifiedCount)

	// Test bulk delete
	deleteResult, err := coll.DeleteMany(context.Background(), filter)
	require.NoError(t, err)
	require.Equal(t, int64(4), deleteResult.DeletedCount)
}

func TestClient_Config(t *testing.T) {
	setupTestEnvironment(t)
	defer cleanupTestEnvironment(t)

	// Test with nil config (should use environment variables)
	client, err := NewClient(nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	cfg := client.Config()
	require.NotNil(t, cfg)
	assert.Equal(t, "mongodb://localhost:27017", cfg.URI)
	assert.Equal(t, "test", cfg.Database)

	// Test with custom config
	customConfig := &Config{
		URI:              "mongodb://custom:27017",
		Database:         "custom_db",
		ConnectTimeout:   15 * time.Second,
		OperationTimeout: 10 * time.Second,
		MaxPoolSize:      20,
		MinPoolSize:      2,
		RetryWrites:      true,
		RetryReads:       true,
		Direct:           false,
	}

	client, err = NewClient(customConfig)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	cfg = client.Config()
	require.NotNil(t, cfg)
	assert.Equal(t, customConfig.URI, cfg.URI)
	assert.Equal(t, customConfig.Database, cfg.Database)
	assert.Equal(t, customConfig.MaxPoolSize, cfg.MaxPoolSize)
}

func TestClient_Transactions(t *testing.T) {
	if !isReplicaSetAvailable() {
		t.Skip("Skipping transaction tests: requires replica set")
	}
	client := setupTestClient(t)
	defer client.Close()

	createTestCollection(t, client)
	defer cleanupTestData(t, client)

	ctx := context.Background()

	// Test successful transaction
	_, err := client.Transaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		doc := TestDocument{
			Name:  "transaction_test",
			Value: 1,
			Time:  time.Now(),
		}
		_, err := client.InsertOne(sessCtx, testCollectionName, doc)
		if err != nil {
			return nil, err
		}

		update := bson.M{"$set": bson.M{"value": 2}}
		_, err = client.UpdateOne(sessCtx, testCollectionName, bson.M{"name": "transaction_test"}, update)
		return nil, err
	})
	require.NoError(t, err)

	// Verify transaction results
	var result TestDocument
	err = client.FindOne(ctx, testCollectionName, bson.M{"name": "transaction_test"}, &result)
	require.NoError(t, err)
	require.Equal(t, 2, result.Value)

	// Test failed transaction
	_, err = client.Transaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		doc := TestDocument{
			Name:  "transaction_test", // Duplicate name should cause error
			Value: 3,
			Time:  time.Now(),
		}
		_, err := client.InsertOne(sessCtx, testCollectionName, doc)
		return nil, err
	})
	require.Error(t, err)

	// Verify transaction was rolled back
	err = client.FindOne(ctx, testCollectionName, bson.M{"value": 3}, &result)
	require.Error(t, err)
	require.Equal(t, mongo.ErrNoDocuments, err)
}

func TestClient_Aggregation(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	createTestCollection(t, client)
	defer cleanupTestData(t, client)

	// Insert test documents
	docs := []interface{}{
		TestDocument{Name: "doc1", Value: 5, Time: time.Now()},
		TestDocument{Name: "doc2", Value: 10, Time: time.Now()},
		TestDocument{Name: "doc3", Value: 15, Time: time.Now()},
	}
	_, err := client.InsertMany(context.Background(), testCollectionName, docs)
	require.NoError(t, err)

	pipeline := []bson.M{
		{"$group": bson.M{
			"_id": nil,
			"maxValue": bson.M{"$max": "$value"},
		}},
	}

	ctx := context.Background()
	cursor, err := client.Aggregate(ctx, testCollectionName, pipeline)
	require.NoError(t, err)

	var results []bson.M
	err = cursor.All(ctx, &results)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, int32(15), results[0]["maxValue"])
}

func TestClient_Indexes(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	createTestCollection(t, client)
	defer cleanupTestData(t, client)

	ctx := context.Background()

	// Create index
	indexName, err := client.CreateIndex(ctx, testCollectionName, bson.D{{Key: "value", Value: 1}},
		options.Index().SetUnique(true))
	require.NoError(t, err)
	require.NotEmpty(t, indexName)

	// Test unique constraint
	doc1 := TestDocument{
		Name:  "index_test_1",
		Value: 1,
		Time:  time.Now(),
	}
	_, err = client.InsertOne(ctx, testCollectionName, doc1)
	require.NoError(t, err)

	doc2 := TestDocument{
		Name:  "index_test_2",
		Value: 1, // Same value should violate unique constraint
		Time:  time.Now(),
	}
	_, err = client.InsertOne(ctx, testCollectionName, doc2)
	require.Error(t, err)
}

func TestClient_Watch(t *testing.T) {
	if !isReplicaSetAvailable() {
		t.Skip("Skipping change stream tests: requires replica set")
	}
	client := setupTestClient(t)
	defer client.Close()

	createTestCollection(t, client)
	defer cleanupTestData(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start watching collection changes
	pipeline := []bson.M{}
	changeStream, err := client.Watch(ctx, testCollectionName, pipeline)
	require.NoError(t, err)
	defer changeStream.Close(ctx)

	// Insert a document
	doc := TestDocument{
		Name:  "watch_test",
		Value: 1,
		Time:  time.Now(),
	}
	_, err = client.InsertOne(ctx, testCollectionName, doc)
	require.NoError(t, err)

	// Wait for change event
	require.True(t, changeStream.Next(ctx))
	var changeEvent bson.M
	err = changeStream.Decode(&changeEvent)
	require.NoError(t, err)

	operationType := changeEvent["operationType"]
	require.Equal(t, "insert", operationType)
}

func TestClient_ErrorHandling(t *testing.T) {
	client := setupTestClient(t)
	
	// Close the client to test error handling
	err := client.Close()
	require.NoError(t, err)

	// Attempt operations on closed client
	_, err = client.Collection(testCollectionName).InsertOne(context.Background(), bson.M{"test": "value"})
	require.Error(t, err)
	assert.Equal(t, "client is disconnected", err.Error())
}

func TestClient_Distinct(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	createTestCollection(t, client)
	defer cleanupTestData(t, client)

	// Insert test documents
	insertTestDocuments(t, client, 5)

	ctx := context.Background()
	values, err := client.Distinct(ctx, testCollectionName, "value", bson.M{})
	require.NoError(t, err)
	require.Len(t, values, 5)
}

func TestClient_FindOneAndModify(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	createTestCollection(t, client)
	defer cleanupTestData(t, client)

	ctx := context.Background()

	// Insert test document
	doc := TestDocument{
		Name:  "modify_test",
		Value: 1,
		Time:  time.Now(),
	}
	_, err := client.InsertOne(ctx, testCollectionName, doc)
	require.NoError(t, err)

	// Test FindOneAndUpdate
	update := bson.M{"$set": bson.M{"value": 2}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := client.FindOneAndUpdate(ctx, testCollectionName, bson.M{"name": "modify_test"}, update, opts)
	var updated TestDocument
	err = result.Decode(&updated)
	require.NoError(t, err)
	require.Equal(t, 2, updated.Value)

	// Test FindOneAndDelete
	result = client.FindOneAndDelete(ctx, testCollectionName, bson.M{"name": "modify_test"})
	var deleted TestDocument
	err = result.Decode(&deleted)
	require.NoError(t, err)
	require.Equal(t, "modify_test", deleted.Name)

	// Verify document was deleted
	assertDocumentNotExists(t, client, bson.M{"name": "modify_test"})
}

func TestClient_BulkWrite(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	createTestCollection(t, client)
	defer cleanupTestData(t, client)

	ctx := context.Background()

	// Prepare bulk write models
	models := []mongo.WriteModel{
		mongo.NewInsertOneModel().SetDocument(TestDocument{
			Name:  "bulk_1",
			Value: 1,
			Time:  time.Now(),
		}),
		mongo.NewInsertOneModel().SetDocument(TestDocument{
			Name:  "bulk_2",
			Value: 2,
			Time:  time.Now(),
		}),
		mongo.NewUpdateOneModel().
			SetFilter(bson.M{"name": "bulk_1"}).
			SetUpdate(bson.M{"$set": bson.M{"value": 3}}),
		mongo.NewDeleteOneModel().
			SetFilter(bson.M{"name": "bulk_2"}),
	}

	// Execute bulk write
	result, err := client.BulkWrite(ctx, testCollectionName, models)
	require.NoError(t, err)
	require.Equal(t, int64(2), result.InsertedCount)
	require.Equal(t, int64(1), result.ModifiedCount)
	require.Equal(t, int64(1), result.DeletedCount)

	// Verify results
	var doc TestDocument
	err = client.FindOne(ctx, testCollectionName, bson.M{"name": "bulk_1"}, &doc)
	require.NoError(t, err)
	require.Equal(t, 3, doc.Value)

	assertDocumentNotExists(t, client, bson.M{"name": "bulk_2"})
}

// Helper function to check if replica set is available
func isReplicaSetAvailable() bool {
	config := getTestConfig()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	opts := options.Client().ApplyURI(config.URI).
		SetAuth(options.Credential{
			Username: config.Username,
			Password: config.Password,
		})
	
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return false
	}
	defer client.Disconnect(ctx)
	
	result := client.Database("admin").RunCommand(ctx, bson.D{{Key: "replSetGetStatus", Value: 1}})
	return result.Err() == nil
}
