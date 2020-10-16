package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sschwartz96/stockpile/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient holds the connection to the database
type MongoClient struct {
	*mongo.Client
	collectionMap map[string]*mongo.Collection
	searchIndices map[string](map[string]bool)
}

// NewMongoClient makes a connection with the mongo client
func NewMongoClient(dbName string, collections []string, opts *options.ClientOptions, searchIndices map[string](map[string]bool)) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// connect to client
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	// confirm the connection with a ping
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &MongoClient{
		Client:        client,
		collectionMap: createCollectionMap(client.Database(dbName), collections),
		searchIndices: searchIndices,
	}, nil
}

// createCollectionMap creates a map of mongo collections so the program doesn't
// reallocate space for a collection every time a request is called
func createCollectionMap(db *mongo.Database, collections []string) map[string]*mongo.Collection {
	collectionMap := make(map[string]*mongo.Collection, len(collections))
	for _, collection := range collections {
		collectionMap[collection] = db.Collection(collection)
	}
	return collectionMap
}

func (m *MongoClient) Open(ctx context.Context) error {
	// already opened when using CreateMongoClient
	return nil
}

func (m *MongoClient) Close(ctx context.Context) error {
	return m.Disconnect(ctx)
}

// Insert takes a collection name and interface object and inserts into collection
func (c *MongoClient) Insert(collection string, object interface{}) error {
	col := c.collectionMap[collection]

	res, err := col.InsertOne(context.Background(), object)
	if err != nil {
		return err
	}

	if res.InsertedID != nil {
		return nil
	}
	return errors.New("failed to insert object into: " + collection)
}

func (m *MongoClient) FindOne(collection string, object interface{}, filter *db.Filter, opts *db.Options) error {
	col := m.collectionMap[collection]
	f := db.ConvertToMongoFilter(filter)
	o := db.ConvertToFindOneOptions(opts)
	res := col.FindOne(context.Background(), f, o)
	return res.Decode(object)
}

// FindAll finds all within the collection, using filter and options if applicable
func (m *MongoClient) FindAll(collection string, object interface{}, filter *db.Filter, opts *db.Options) error {
	col := m.collectionMap[collection]
	f := db.ConvertToMongoFilter(filter)
	o := db.ConvertToFindOptions(opts)
	cur, err := col.Find(context.Background(), f, o)
	if err != nil {
		return err
	}
	err = cur.All(context.Background(), object)
	return err
}

func (m *MongoClient) Update(collection string, object interface{}, filter *db.Filter) error {
	col := m.collectionMap[collection]
	f := db.ConvertToMongoFilter(filter)
	u := bson.M{"$set": object}
	res, err := col.UpdateOne(context.Background(), f, u)
	if err != nil {
		return err
	}
	if res.MatchedCount > 1 {
		if res.ModifiedCount == 0 {
			return fmt.Errorf("error mongo update: matched %v, but didn't modify", res.MatchedCount)
		}
	} else {
		return fmt.Errorf("error mongo update: did not match any documents")
	}
	return nil
}

// Upsert updates or inserts object within collection with premade filter
func (c *MongoClient) Upsert(collection string, object interface{}, filter *db.Filter) error {
	col := c.collectionMap[collection]
	update := bson.M{"$set": object}
	f := db.ConvertToMongoFilter(filter)

	upsert := true
	opts := &options.UpdateOptions{Upsert: &upsert}

	_, err := col.UpdateOne(context.Background(), f, update, opts)
	if err != nil {
		return err
	}

	return nil
}

// Delete deletes the certain document based on param and value
func (c *MongoClient) Delete(collection string, filter *db.Filter) error {
	col := c.collectionMap[collection]
	f := db.ConvertToMongoFilter(filter)
	res, err := col.DeleteOne(context.Background(), f)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("error mongo delete: deleted count == 0")
	}
	return nil
}

// Find takes collection, param & value to build filter, and object pointer to receive data
func (c *MongoClient) Find(collection, param string, value interface{}, object interface{}) error {
	filter := bson.D{{
		Key:   param,
		Value: value,
	}}

	return c.FindWithBSON(collection, filter, options.FindOne(), object)
}

// FindWithBSON takes in object and already made bson filter
func (c *MongoClient) FindWithBSON(collection string, filter interface{}, opts *options.FindOneOptions, object interface{}) error {
	var err error

	// get collection
	col := c.collectionMap[collection]

	// find operation
	if opts == nil {
		opts = options.FindOne()
	}
	result := col.FindOne(context.Background(), filter, opts)
	err = result.Err()
	if err != nil {
		return err
	}
	// decode one
	err = result.Decode(object)

	return err
}

// FindAllWithBSON takes collection string, bson filter, options.FindOptions
// and decodes into pointer to the slice
func (c *MongoClient) FindAllWithBSON(collection string, filter interface{}, opts *options.FindOptions, slice interface{}) error {
	// get collection
	col := c.collectionMap[collection]

	// find operation
	cur, err := col.Find(context.Background(), filter, opts)
	if err != nil {
		return err
	}
	// decode all
	err = cur.All(context.Background(), slice)
	return err

}

// UpdateWithBSON takes in collection string & bson filter and update object
func (c *MongoClient) UpdateWithBSON(collection string, filter, update interface{}) error {
	col := c.collectionMap[collection]
	r, err := col.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if r.ModifiedCount != 1 {
		if r.MatchedCount == 1 {
			return errors.New("matched but not updated")
		}
		return errors.New("object failed to update")
	}
	return nil
}

// Exists checks if the document exists within the collection based on the filter
func (c *MongoClient) Exists(collection string, filter interface{}) (bool, error) {
	col := c.collectionMap[collection]

	// setup limit in FindOptions
	limit := int64(1)
	opts := options.FindOptions{Limit: &limit}

	cur, err := col.Find(context.Background(), filter, &opts)
	if err != nil {
		return false, err
	}

	return cur.TryNext(context.Background()), nil
}

func (c *MongoClient) doesIndexExists(collection string, fields []string) bool {
	indexMap, ok := c.searchIndices[collection]
	if !ok {
		return false
	}
	for _, field := range fields {
		if !indexMap[field] {
			return false
		}
	}
	return true
}

//func (c *mongoClient) createSearchIndex(collection string, fields []string) error {
//	col := c.collectionMap[collection]
//	var indexes bson.M
//	for _, field := range fields {
//		indexes[field] = "text"
//	}
//	indexModel := mongo.IndexModel{Keys: indexes}
//
//	index, err := col.Indexes().CreateOne(context.Background(), indexModel)
//	if err != nil {
//		return fmt.Errorf("createSearchIndex() error: %v", err)
//	}
//	fmt.Println("created index named: ", index)
//	return nil
//}

// Search takes a collection, search string, and slice of fields to search upon.
// The results are unmarshalled into slice interface
func (c *MongoClient) Search(collection, search string, fields []string, slice interface{}) error {
	col := c.collectionMap[collection]
	if !c.doesIndexExists(collection, fields) {
		// TODO: create indices??? no, because we should have them already created
		return errors.New("Search() search indices do not exist")
	}
	// create search filter
	filter := bson.M{
		"$text": bson.M{
			"$search": search,
		},
		"score": bson.M{"$meta": "textScore"},
	}
	// sort by score
	opts := options.Find().SetSort(bson.M{"score": bson.M{"$meta": "textScore"}})
	// run search
	cur, err := col.Find(context.Background(), filter, opts)
	if err != nil {
		return err
	}
	return cur.All(context.Background(), slice)
}

// Aggregate takes in a collection string, filter, pipeline, and pointer to object
// returns error if anything is malformed
func (c *MongoClient) Aggregate(collection string, pipeline mongo.Pipeline, object interface{}) error {
	col := c.collectionMap[collection]
	cur, err := col.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	return cur.All(context.Background(), object)
}
