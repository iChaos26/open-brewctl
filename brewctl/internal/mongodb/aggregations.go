package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ✅ CORREÇÃO: Pipeline corrigido com campos nomeados
func (s *AggregationService) RunSilverLayerAggregation() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("🔄 Running Silver Layer Aggregation...")

	pipeline := mongo.Pipeline{
		// ✅ CORREÇÃO: Todos os campos nomeados com bson.E
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "id", Value: 1},
			{Key: "name", Value: 1},
			{Key: "brewery_type", Value: 1},
			{Key: "address_1", Value: 1},
			{Key: "city", Value: bson.D{{Key: "$trim", Value: bson.D{{Key: "input", Value: "$city"}}}}},
			{Key: "state_province", Value: bson.D{{Key: "$trim", Value: bson.D{{Key: "input", Value: "$state_province"}}}}},
			{Key: "state", Value: bson.D{{Key: "$trim", Value: bson.D{{Key: "input", Value: "$state"}}}}},
			{Key: "country", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{{Key: "$eq", Value: bson.A{"$country", "United States"}}}},
					{Key: "then", Value: "US"},
					{Key: "else", Value: "$country"},
				}},
			}},
			{Key: "postal_code", Value: 1},
			{Key: "longitude", Value: bson.D{
				{Key: "$convert", Value: bson.D{
					{Key: "input", Value: "$longitude"},
					{Key: "to", Value: "double"},
					{Key: "onError", Value: nil},
					{Key: "onNull", Value: nil},
				}},
			}},
			{Key: "latitude", Value: bson.D{
				{Key: "$convert", Value: bson.D{
					{Key: "input", Value: "$latitude"},
					{Key: "to", Value: "double"},
					{Key: "onError", Value: nil},
					{Key: "onNull", Value: nil},
				}},
			}},
			{Key: "phone", Value: 1},
			{Key: "website_url", Value: 1},
			{Key: "street", Value: 1},
			// Data quality metrics
			{Key: "data_quality", Value: bson.D{
				{Key: "has_coordinates", Value: bson.D{
					{Key: "$and", Value: bson.A{
						bson.D{{Key: "$ne", Value: bson.A{"$longitude", nil}}},
						bson.D{{Key: "$ne", Value: bson.A{"$latitude", nil}}},
					}},
				}},
				{Key: "has_website", Value: bson.D{{Key: "$ne", Value: bson.A{"$website_url", nil}}}},
				{Key: "has_phone", Value: bson.D{{Key: "$ne", Value: bson.A{"$phone", nil}}}},
				{Key: "completeness_score", Value: bson.D{
					{Key: "$divide", Value: bson.A{
						bson.D{{Key: "$size", Value: bson.D{
							{Key: "$filter", Value: bson.D{
								{Key: "input", Value: bson.A{"$name", "$brewery_type", "$city", "$state", "$country"}},
								{Key: "as", Value: "field"},
								{Key: "cond", Value: bson.D{{Key: "$ne", Value: bson.A{"$$field", nil}}}},
							}},
						}}},
						5,
					}},
				}},
			}},
			{Key: "ingestion_date", Value: time.Now()},
			{Key: "last_updated", Value: time.Now()},
		}}},

		// ✅ CORREÇÃO: Merge corrigido
		{{Key: "$merge", Value: bson.D{
			{Key: "into", Value: "breweries_clean"},
			{Key: "on", Value: "_id"},
			{Key: "whenMatched", Value: "replace"},
			{Key: "whenNotMatched", Value: "insert"},
		}}},
	}

	// ✅ CORREÇÃO: Execução corrigida - não precisamos do cursor, só executar
	_, err := s.DB.Collection("breweries_raw").Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("silver aggregation failed: %v", err)
	}

	// Check if any documents were processed
	count, err := s.DB.Collection("breweries_clean").CountDocuments(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("failed to count documents: %v", err)
	}

	fmt.Printf("✅ Silver layer completed. Processed documents: %d\n", count)
	return nil
}
