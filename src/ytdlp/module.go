package ytdlp

import (
	"time"

	"gorm.io/gorm"
)

type YtdlpModule struct {
	cache        *YtdlpMetadataStorage
	YtdlpService *YtdlpService
}

func NewYtdlpModule(
	db *gorm.DB,
	path string,
	maxCacheLifetime time.Duration,
	maxParallelism int,
) *YtdlpModule {
	ytdlp := newYtdlp(path)
	cache := newYtdlpMetadataStorage(db)
	service := newYtdlpService(ytdlp, cache, maxCacheLifetime, maxParallelism)

	return &YtdlpModule{
		cache:        cache,
		YtdlpService: service,
	}
}
