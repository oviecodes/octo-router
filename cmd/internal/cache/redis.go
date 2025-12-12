package cache

import (
	"llm-router/types"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// have 1 interface and 2 structs that implement them namely SemanticCache and DefaultCache

type Cache interface {
	SetItem()
	GetItem()
}

type SemanticCache struct {
	client *redis.Client
	config cacheConfig
}

type DefaultCache struct {
	client *redis.Client
	config cacheConfig
}

type cacheConfig struct {
	strategy            string
	ttl                 int
	similarityThreshold float64
}

func (d *DefaultCache) GetItem() {}

func (d *DefaultCache) SetItem() {}

func (s *SemanticCache) GetItem() {}

func (s *SemanticCache) SetItem() {}

func NewCacheClient(config types.CacheData) (Cache, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	semanticEnabled := config.Semantic["enabled"]
	similarityThreshold := config.Semantic["similaritythreshold"]

	derivedSimilarityThreshold, _ := strconv.ParseFloat(similarityThreshold, 64)

	if semanticEnabled == "1" {
		return &SemanticCache{
			client: client,
			config: cacheConfig{
				strategy:            "semantic",
				similarityThreshold: derivedSimilarityThreshold,
				ttl:                 config.Ttl,
			},
		}, nil
	}

	return &DefaultCache{
		client: client,
		config: cacheConfig{
			strategy:            "default",
			similarityThreshold: derivedSimilarityThreshold,
			ttl:                 config.Ttl,
		},
	}, nil
}
