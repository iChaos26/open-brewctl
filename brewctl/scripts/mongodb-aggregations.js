// MongoDB Aggregation Scripts for Breweries Data

// 1. Silver Layer Transformation - Data Cleaning
db.breweries_raw.aggregate([
    {
        $project: {
            _id: 1,
            id: 1,
            name: 1,
            brewery_type: 1,
            address_1: 1,
            city: { $trim: { input: "$city" } },
            state_province: { $trim: { input: "$state_province" } },
            state: { $trim: { input: "$state" } },
            country: { 
                $cond: {
                    if: { $eq: ["$country", "United States"] },
                    then: "US",
                    else: "$country"
                }
            },
            postal_code: 1,
            longitude: { $convert: { input: "$longitude", to: "double", onError: null } },
            latitude: { $convert: { input: "$latitude", to: "double", onError: null } },
            phone: 1,
            website_url: 1,
            street: 1,
            data_quality: {
                has_coordinates: { $and: [{ $ne: ["$longitude", null] }, { $ne: ["$latitude", null] }] },
                has_website: { $ne: ["$website_url", null] },
                has_phone: { $ne: ["$phone", null] },
                completeness_score: {
                    $divide: [
                        {
                            $size: {
                                $filter: {
                                    input: ["$name", "$brewery_type", "$city", "$state", "$country"],
                                    as: "field",
                                    cond: { $ne: ["$$field", null] }
                                }
                            }
                        },
                        5
                    ]
                }
            },
            ingestion_date: { $dateToString: { format: "%Y-%m-%d", date: new Date() } },
            last_updated: new Date()
        }
    },
    {
        $merge: {
            into: "breweries_clean",
            on: "_id",
            whenMatched: "replace",
            whenNotMatched: "insert"
        }
    }
]);

// 2. Gold Layer Aggregation - Business Metrics
db.breweries_clean.aggregate([
    {
        $match: {
            "data_quality.completeness_score": { $gte: 0.8 }
        }
    },
    {
        $group: {
            _id: {
                country: "$country",
                state: "$state",
                brewery_type: "$brewery_type"
            },
            total_breweries: { $sum: 1 },
            breweries_with_website: {
                $sum: { $cond: [{ $ne: ["$website_url", null] }, 1, 0] }
            },
            breweries_with_phone: {
                $sum: { $cond: [{ $ne: ["$phone", null] }, 1, 0] }
            },
            breweries_with_coordinates: {
                $sum: { $cond: ["$data_quality.has_coordinates", 1, 0] }
            },
            avg_completeness_score: { $avg: "$data_quality.completeness_score" },
            example_breweries: { $push: { name: "$name", city: "$city" } }
        }
    },
    {
        $project: {
            _id: 0,
            country: "$_id.country",
            state: "$_id.state",
            brewery_type: "$_id.brewery_type",
            total_breweries: 1,
            breweries_with_website: 1,
            breweries_with_phone: 1,
            breweries_with_coordinates: 1,
            website_coverage: {
                $round: [
                    { $multiply: [{ $divide: ["$breweries_with_website", "$total_breweries"] }, 100] },
                    2
                ]
            },
            phone_coverage: {
                $round: [
                    { $multiply: [{ $divide: ["$breweries_with_phone", "$total_breweries"] }, 100] },
                    2
                ]
            },
            coordinates_coverage: {
                $round: [
                    { $multiply: [{ $divide: ["$breweries_with_coordinates", "$total_breweries"] }, 100] },
                    2
                ]
            },
            avg_completeness_score: { $round: ["$avg_completeness_score", 3] },
            example_breweries: { $slice: ["$example_breweries", 3] },
            aggregation_date: { $dateToString: { format: "%Y-%m-%d", date: new Date() } }
        }
    },
    {
        $merge: {
            into: "breweries_aggregated",
            on: ["country", "state", "brewery_type"],
            whenMatched: "replace",
            whenNotMatched: "insert"
        }
    }
]);

// 3. Query Examples for Analysis

// Top 10 States by Brewery Count
db.breweries_clean.aggregate([
    {
        $group: {
            _id: "$state",
            total_breweries: { $sum: 1 }
        }
    },
    {
        $sort: { total_breweries: -1 }
    },
    {
        $limit: 10
    },
    {
        $project: {
            _id: 0,
            state: "$_id",
            total_breweries: 1
        }
    }
]);

// Brewery Type Distribution
db.breweries_clean.aggregate([
    {
        $group: {
            _id: "$brewery_type",
            count: { $sum: 1 },
            states: { $addToSet: "$state" }
        }
    },
    {
        $project: {
            _id: 0,
            brewery_type: "$_id",
            count: 1,
            states_covered: { $size: "$states" }
        }
    },
    {
        $sort: { count: -1 }
    }
]);

// Geographic Distribution (for mapping)
db.breweries_clean.aggregate([
    {
        $match: {
            "data_quality.has_coordinates": true
        }
    },
    {
        $project: {
            _id: 0,
            name: 1,
            brewery_type: 1,
            city: 1,
            state: 1,
            country: 1,
            longitude: 1,
            latitude: 1,
            website_url: 1
        }
    },
    {
        $limit: 100
    }
]);