package brewerydb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BreweryImporter struct {
	DBClient *BreweryDBClient
	MongoDB  *mongo.Database
}

func NewBreweryImporter(mongoURI string) (*BreweryImporter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	db := client.Database("breweries_db")
	return &BreweryImporter{
		DBClient: NewBreweryDBClient(),
		MongoDB:  db,
	}, nil
}

func (bi *BreweryImporter) ImportAllBreweries() error {
	breweries, err := bi.DBClient.GetAllBreweries()
	if err != nil {
		return fmt.Errorf("failed to get breweries: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	collection := bi.MongoDB.Collection("breweries_raw")

	// Clear existing data
	if _, err := collection.DeleteMany(ctx, bson.M{}); err != nil {
		return fmt.Errorf("failed to clear collection: %v", err)
	}

	var documents []interface{}
	for _, brewery := range breweries {
		doc := bson.M{
			"brewery":     brewery,
			"imported_at": time.Now(),
		}
		documents = append(documents, doc)
	}

	if len(documents) == 0 {
		fmt.Println("⚠️ No breweries to import")
		return nil
	}

	result, err := collection.InsertMany(ctx, documents)
	if err != nil {
		return fmt.Errorf("failed to insert documents: %v", err)
	}

	fmt.Printf("✅ Imported %d breweries into MongoDB\n", len(result.InsertedIDs))
	return nil
}

func (bi *BreweryImporter) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bi.MongoDB.Client().Disconnect(ctx)
}
