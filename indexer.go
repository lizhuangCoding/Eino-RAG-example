package main

import (
	"context"

	redisInd "github.com/cloudwego/eino-ext/components/indexer/redis"
)

// newIndexer 在以 Redis-stack 为向量数据库的场景下，其主要目的便是将embedder返回的向量数据存储在Redis向量数据库中，便于后续进行查询。
func (r *RAGEngine) newIndexer(ctx context.Context) {
	i, err := redisInd.NewIndexer(ctx, &redisInd.IndexerConfig{
		Client:           r.redis,
		KeyPrefix:        r.prefix,
		DocumentToHashes: nil,
		BatchSize:        10,
		Embedding:        r.embedder,
	})
	if err != nil {
		r.Err = err
	}
	r.Indexer = i
}

func (r *RAGEngine) InitVectorIndex(ctx context.Context) error {
	// FT.INFO <indexName> 指令用于获取索引的详细信息，这里如果索引不存在会返回Err
	// 若能够获取到索引信息则Err为空，说明索引已存在，不需要重复创建
	if _, err := r.redis.Do(ctx, "FT.INFO", r.indexName).Result(); err == nil {
		return nil
	}

	createIndexArgs := []interface{}{
		"FT.CREATE", r.indexName,
		"ON", "HASH",
		"PREFIX", "1", r.prefix,
		"SCHEMA",
		"content", "TEXT",
		"vector_content", "VECTOR", "FLAT",
		"6",
		"TYPE", "FLOAT32",
		"DIM", r.dimension,
		"DISTANCE_METRIC", "COSINE",
	}

	if err := r.redis.Do(ctx, createIndexArgs...).Err(); err != nil {
		return err
	}

	// 再次检查索引是否已创建。
	if _, err := r.redis.Do(ctx, "FT.INFO", r.indexName).Result(); err != nil {
		return err
	}
	return nil
}
