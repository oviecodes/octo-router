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

func NewRedisClient(config types.RedisData) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})
}

func NewCacheClient(cfg types.CacheData, redisConfig types.RedisData) (Cache, error) {

	client := NewRedisClient(redisConfig)

	semanticEnabled := cfg.Semantic["enabled"]
	similarityThreshold := cfg.Semantic["similaritythreshold"]

	derivedSimilarityThreshold, _ := strconv.ParseFloat(similarityThreshold, 64)

	if semanticEnabled == "1" {
		return &SemanticCache{
			client: client,
			config: cacheConfig{
				strategy:            "semantic",
				similarityThreshold: derivedSimilarityThreshold,
				ttl:                 cfg.Ttl,
			},
		}, nil
	}

	return &DefaultCache{
		client: client,
		config: cacheConfig{
			strategy:            "default",
			similarityThreshold: derivedSimilarityThreshold,
			ttl:                 cfg.Ttl,
		},
	}, nil
}
