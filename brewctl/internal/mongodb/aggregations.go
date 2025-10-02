package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

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

	// ✅ CORREÇÃO: Execução corrigida
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

// ✅ IMPLEMENTAÇÃO: Gold Layer Aggregation faltante
// ✅ ADICIONAR: RunGoldLayerAggregation faltante
func (s *AggregationService) RunGoldLayerAggregation() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("🔄 Running Gold Layer Aggregation...")

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "data_quality.completeness_score", Value: bson.D{{Key: "$gte", Value: 0.6}}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "country", Value: "$country"},
				{Key: "state", Value: "$state"},
				{Key: "brewery_type", Value: "$brewery_type"},
			}},
			{Key: "total_breweries", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "breweries_with_website", Value: bson.D{
				{Key: "$sum", Value: bson.D{
					{Key: "$cond", Value: bson.A{bson.D{{Key: "$ne", Value: bson.A{"$website_url", nil}}}, 1, 0}},
				}},
			}},
			{Key: "breweries_with_phone", Value: bson.D{
				{Key: "$sum", Value: bson.D{
					{Key: "$cond", Value: bson.A{bson.D{{Key: "$ne", Value: bson.A{"$phone", nil}}}, 1, 0}},
				}},
			}},
			{Key: "breweries_with_coordinates", Value: bson.D{
				{Key: "$sum", Value: bson.D{
					{Key: "$cond", Value: bson.A{"$data_quality.has_coordinates", 1, 0}},
				}},
			}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "country", Value: "$_id.country"},
			{Key: "state", Value: "$_id.state"},
			{Key: "brewery_type", Value: "$_id.brewery_type"},
			{Key: "total_breweries", Value: 1},
			{Key: "breweries_with_website", Value: 1},
			{Key: "breweries_with_phone", Value: 1},
			{Key: "breweries_with_coordinates", Value: 1},
		}}},
		{{Key: "$merge", Value: bson.D{
			{Key: "into", Value: "breweries_aggregated"},
			{Key: "on", Value: bson.A{"country", "state", "brewery_type"}},
			{Key: "whenMatched", Value: "replace"},
			{Key: "whenNotMatched", Value: "insert"},
		}}},
	}

	_, err := s.DB.Collection("breweries_clean").Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("gold aggregation failed: %v", err)
	}

	count, err := s.DB.Collection("breweries_aggregated").CountDocuments(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("failed to count aggregated documents: %v", err)
	}

	fmt.Printf("✅ Gold layer completed. Aggregated records: %d\n", count)
	return nil
}

// ✅ IMPLEMENTAÇÃO: GetTopStates faltante
func (s *AggregationService) GetTopStates(limit int) ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$state"},
			{Key: "total_breweries", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "total_breweries", Value: -1}}}},
		{{Key: "$limit", Value: int64(limit)}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "state", Value: "$_id"},
			{Key: "total_breweries", Value: 1},
		}}},
	}

	cursor, err := s.DB.Collection("breweries_clean").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// ✅ IMPLEMENTAÇÃO: GetBreweryTypesDistribution faltante
func (s *AggregationService) GetBreweryTypesDistribution() ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$brewery_type"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "states", Value: bson.D{{Key: "$addToSet", Value: "$state"}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "brewery_type", Value: "$_id"},
			{Key: "count", Value: 1},
			{Key: "states_covered", Value: bson.D{{Key: "$size", Value: "$states"}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
	}

	cursor, err := s.DB.Collection("breweries_clean").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// ✅ IMPLEMENTAÇÃO: GetGeographicDistribution (opcional, se necessário)
func (s *AggregationService) GetGeographicDistribution() ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "data_quality.has_coordinates", Value: true},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "name", Value: 1},
			{Key: "brewery_type", Value: 1},
			{Key: "city", Value: 1},
			{Key: "state", Value: 1},
			{Key: "country", Value: 1},
			{Key: "longitude", Value: 1},
			{Key: "latitude", Value: 1},
		}}},
		{{Key: "$limit", Value: 100}},
	}

	cursor, err := s.DB.Collection("breweries_clean").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
