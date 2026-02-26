package main

import (
	"context"
	"fmt"

	"Eino-RAG-example/config"
	"github.com/cloudwego/eino-ext/components/document/loader/file"
	embedding "github.com/cloudwego/eino-ext/components/embedding/ark"
	redisInd "github.com/cloudwego/eino-ext/components/indexer/redis"
	"github.com/cloudwego/eino-ext/components/model/ark"
	redisRet "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

type RAGEngine struct {
	indexName string               // 索引的名字
	prefix    string               // 前缀，用于区分不同作用的键
	config    *config.ParamsConfig // 配置信息
	dimension int                  // 向量的维度
	redis     *redis.Client        // redis，存放向量
	embedder  *embedding.Embedder  // 向量模型
	Err       error                // 错误（统一处理）
	Loader    *file.FileLoader     // 读取外部文件
	Splitter  document.Transformer // 分割文件
	Indexer   *redisInd.Indexer    // 嵌入向量（本地存储到redis向量数据库中）
	Retriever *redisRet.Retriever  // 检索内容
	ChatModel *ark.ChatModel       // AI模型
}

func InitRAGEngine(ctx context.Context, index string, prefix string) (*RAGEngine, error) {
	r, err := initRAGEngine(ctx, index, prefix)
	if err != nil {
		return nil, err
	}
	// 文件加载
	r.newLoader(ctx)
	// 分离
	r.newSplitter(ctx)
	// 索引
	r.newIndexer(ctx)
	// 搜索，返回最近内容
	r.newRetriever(ctx)
	// 大模型
	r.newChatModel(ctx)

	return r, nil
}

func initRAGEngine(ctx context.Context, index string, prefix string) (*RAGEngine, error) {
	// 加载配置
	c := config.Map()

	// 创建 embedder 用于将文档转成向量，Indexer 与 Retriever 均依赖于该组件。
	embedder, err := embedding.NewEmbedder(ctx, &embedding.EmbeddingConfig{
		APIKey: c.ApiKey,
		Model:  c.Embedding,
	})
	if err != nil {
		return nil, err
	}

	return &RAGEngine{
		indexName: index,
		prefix:    prefix,
		config:    c,
		dimension: 4096,
		redis: redis.NewClient(&redis.Options{
			Addr:          fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port),
			Protocol:      2,
			UnstableResp3: true,
		}),
		embedder:  embedder,
		Loader:    nil,
		Splitter:  nil,
		Retriever: nil,
		Indexer:   nil,
		ChatModel: nil,
	}, nil
}

// 系统提示词：创建完索引并将向量存入数据库后，我们便可以在之后生成时，附带检索到的文档进行增强生成了
var systemPrompt = `
# Role: Student Learning Assistant

# Language: Chinese

- When providing assistance:
  • Be clear and concise
  • Include practical examples when relevant
  • Reference documentation when helpful
  • Suggest improvements or next steps if applicable

here's documents searched for you:
==== doc start ====
	  {documents}
==== doc end ====
`

// Stream 流式输出。从redis中查询相关的信息，把查询到的信息作为template模板发送给大模型，最后由大模型进行回复
func (r *RAGEngine) Stream(ctx context.Context, query string) (*schema.StreamReader[*schema.Message], error) {
	// 依据查询内容进行检索
	docs, err := r.Retriever.Retrieve(ctx, query)
	if err != nil {
		return nil, err
	}

	fmt.Println("-------------------------------------------")
	fmt.Println("===== 从知识库中查询到相关内容：=====")
	fmt.Println(docs)
	fmt.Println("-------------------------------------------")

	// 创建template模板并植入文档数据
	tpl := prompt.FromMessages(schema.FString, []schema.MessagesTemplate{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage("question: {content}"),
	}...)

	messages, err := tpl.Format(ctx, map[string]any{
		"documents": docs,
		"content":   query,
	})
	if err != nil {
		return nil, err
	}

	// LLM 进行文本生成
	return r.ChatModel.Stream(ctx, messages)
}
