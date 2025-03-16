package cache

import (
	"fmt"

	"github.com/chats/go-user-api/config"
	"github.com/rs/zerolog/log"
)

// Factory is an interface for creating cache connections
type Factory interface {
	Create(config config.CacheConfig) (Cache, error)
}

// CacheFactory implements the Factory interface
type CacheFactory struct{}

// NewCacheFactory creates a new CacheFactory
func NewCacheFactory() Factory {
	return &CacheFactory{}
}

// Create creates a new cache connection based on the provided configuration
func (f *CacheFactory) Create(config config.CacheConfig) (Cache, error) {
	switch config.Type {
	case "redis":
		log.Info().Msg("Creating Redis cache connection")
		return NewRedis(config)
	//case "memcached":
	//	log.Info().Msg("Creating Memcached cache connection")
	//	return NewMemcached(config)
	default:
		return nil, fmt.Errorf("unsupported cache type: %s", config.Type)
	}
}
