package main

import (
	"context"

	redisRet "github.com/cloudwego/eino-ext/components/retriever/redis"
)

// Retriever 把 Indexer 构建索引之后的内容进行召回

// newRetriever 检索，其目的是在Redis中进行最近邻搜索，返回最近的TopK个文档。
func (r *RAGEngine) newRetriever(ctx context.Context) {
	re, err := redisRet.NewRetriever(ctx, &redisRet.RetrieverConfig{
		Client:            r.redis,
		Index:             r.indexName,
		VectorField:       "vector_content",
		DistanceThreshold: nil,
		Dialect:           2,
		ReturnFields:      []string{"vector_content", "content"},
		DocumentConverter: nil,
		TopK:              1,
		Embedding:         r.embedder,
	})
	if err != nil {
		r.Err = err
		return
	}
	r.Retriever = re
}
