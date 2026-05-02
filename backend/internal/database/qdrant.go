package database

import (
	"context"
	"log"
	"net/url"
	"strings"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
)

// QdrantClient manages the connection to Qdrant Cloud.
type QdrantClient struct {
	Client *qdrant.Client
}

// NewQdrantClient establishes a gRPC connection to the Qdrant Cloud cluster.
func NewQdrantClient(urlStr string, apiKey string) (*QdrantClient, error) {
	host := urlStr
	if parsed, err := url.Parse(urlStr); err == nil && parsed.Hostname() != "" {
		host = parsed.Hostname()
	} else if strings.Contains(urlStr, "://") {
		// Fallback
		parts := strings.Split(urlStr, "://")
		if len(parts) > 1 {
			host = strings.Split(parts[1], ":")[0]
		}
	}

	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   host, // e.g. 70629f26-...aws.cloud.qdrant.io
		Port:   6334,
		APIKey: apiKey,
		UseTLS: true,
	})

	if err != nil {
		return nil, err
	}

	return &QdrantClient{
		Client: client,
	}, nil
}

// InitQdrantSchema ensures the `dealna_items` collection and indexes exist.
func (q *QdrantClient) InitQdrantSchema(ctx context.Context) error {
	collectionName := "dealna_items"

	// 1. Create the Collection (384 Dimensions for bge-small-en-v1.5)
	err := q.Client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     384,
					Distance: qdrant.Distance_Cosine,
				},
			},
		},
	})
	if err != nil {
		log.Printf("Qdrant Collection '%s' might already exist or failed to create: %v", collectionName, err)
	} else {
		log.Printf("Qdrant Collection '%s' created successfully.", collectionName)
	}

	// 2. Create Payload Indexes
	isTenant := true

	// A. University ID (Tenant Index)
	_, err = q.Client.CreateFieldIndex(ctx, &qdrant.CreateFieldIndexCollection{
		CollectionName: collectionName,
		FieldName:      "university_id",
		FieldType:      qdrant.FieldType_FieldTypeKeyword.Enum(),
		FieldIndexParams: &qdrant.PayloadIndexParams{
			IndexParams: &qdrant.PayloadIndexParams_KeywordIndexParams{
				KeywordIndexParams: &qdrant.KeywordIndexParams{
					IsTenant: &isTenant, // Physically partitions the HNSW graph!
				},
			},
		},
	})
	if err != nil {
		log.Printf("Failed to create university_id index: %v", err)
	}

	// B. Category (Keyword Index)
	_, err = q.Client.CreateFieldIndex(ctx, &qdrant.CreateFieldIndexCollection{
		CollectionName: collectionName,
		FieldName:      "category",
		FieldType:      qdrant.FieldType_FieldTypeKeyword.Enum(),
	})
	if err != nil {
		log.Printf("Failed to create category index: %v", err)
	}

	// C. Status (Keyword Index)
	_, err = q.Client.CreateFieldIndex(ctx, &qdrant.CreateFieldIndexCollection{
		CollectionName: collectionName,
		FieldName:      "status",
		FieldType:      qdrant.FieldType_FieldTypeKeyword.Enum(),
	})
	if err != nil {
		log.Printf("Failed to create status index: %v", err)
	}

	// D. Condition (Keyword Index)
	_, err = q.Client.CreateFieldIndex(ctx, &qdrant.CreateFieldIndexCollection{
		CollectionName: collectionName,
		FieldName:      "condition",
		FieldType:      qdrant.FieldType_FieldTypeKeyword.Enum(),
	})
	if err != nil {
		log.Printf("Failed to create condition index: %v", err)
	}

	// E. Price (Float Index)
	_, err = q.Client.CreateFieldIndex(ctx, &qdrant.CreateFieldIndexCollection{
		CollectionName: collectionName,
		FieldName:      "price",
		FieldType:      qdrant.FieldType_FieldTypeFloat.Enum(),
	})
	if err != nil {
		log.Printf("Failed to create price index: %v", err)
	}

	// F. IsGiveaway (Bool Index)
	_, err = q.Client.CreateFieldIndex(ctx, &qdrant.CreateFieldIndexCollection{
		CollectionName: collectionName,
		FieldName:      "is_giveaway",
		FieldType:      qdrant.FieldType_FieldTypeBool.Enum(),
	})
	if err != nil {
		log.Printf("Failed to create is_giveaway index: %v", err)
	}

	log.Println("Qdrant collection and payload indexes successfully initialized!")
	return nil
}

// SearchItems queries Qdrant using the provided vector, enforcing tenant and status constraints.
func (q *QdrantClient) SearchItems(ctx context.Context, vector []float32, filter domain.ItemFilter) ([]uuid.UUID, error) {
	collectionName := "dealna_items"

	// 1. Mandatory Filters: Tenant (University) and active status.
	// We also exclude the current user's items.
	mustConditions := []*qdrant.Condition{
		{
			ConditionOneOf: &qdrant.Condition_Field{
				Field: &qdrant.FieldCondition{
					Key: "university_id",
					Match: &qdrant.Match{
						MatchValue: &qdrant.Match_Keyword{
							Keyword: filter.RequesterUniversityID.String(),
						},
					},
				},
			},
		},
		{
			ConditionOneOf: &qdrant.Condition_Field{
				Field: &qdrant.FieldCondition{
					Key: "status",
					Match: &qdrant.Match{
						MatchValue: &qdrant.Match_Keyword{
							Keyword: "AVAILABLE",
						},
					},
				},
			},
		},
	}

	// 2. Optional UI Filters
	if filter.CategoryID != nil {
		mustConditions = append(mustConditions, &qdrant.Condition{
			ConditionOneOf: &qdrant.Condition_Field{
				Field: &qdrant.FieldCondition{
					Key: "category",
					Match: &qdrant.Match{
						MatchValue: &qdrant.Match_Keyword{
							Keyword: filter.CategoryID.String(),
						},
					},
				},
			},
		})
	}

	if filter.MinPrice != nil || filter.MaxPrice != nil {
		rangeCondition := &qdrant.Range{}
		if filter.MinPrice != nil {
			val := float64(*filter.MinPrice)
			rangeCondition.Gte = &val
		}
		if filter.MaxPrice != nil {
			val := float64(*filter.MaxPrice)
			rangeCondition.Lte = &val
		}
		mustConditions = append(mustConditions, &qdrant.Condition{
			ConditionOneOf: &qdrant.Condition_Field{
				Field: &qdrant.FieldCondition{
					Key:   "price",
					Range: rangeCondition,
				},
			},
		})
	}

	qdrantFilter := &qdrant.Filter{
		Must: mustConditions,
	}

	// 3. Execute Search
	offset := uint64(filter.Offset)
	limitUint64 := uint64(filter.Limit)

	// Set a reasonable threshold to prevent completely unrelated items from matching.
	// Since we use Cosine distance, scores generally range from 0 to 1.
	var threshold float32 = 0.35

	searchResult, err := q.Client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQuery(vector...),
		Filter:         qdrantFilter,
		Limit:          &limitUint64,
		Offset:         &offset,
		ScoreThreshold: &threshold,
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: false}},
	})
	if err != nil {
		return nil, err
	}

	// 4. Extract UUIDs
	var ids []uuid.UUID
	for _, point := range searchResult {
		if idStr, ok := point.Id.PointIdOptions.(*qdrant.PointId_Uuid); ok {
			parsedID, err := uuid.Parse(idStr.Uuid)
			if err == nil {
				// Only include it if it's NOT the excluded owner (if we indexed owner_id we'd filter it in Qdrant, but for now we'll do it later in PG or we can add it to Qdrant. Since owner_id is not in Qdrant payload, Postgres will filter it natively during hydration using an intersection).
				ids = append(ids, parsedID)
			}
		}
	}

	return ids, nil
}
