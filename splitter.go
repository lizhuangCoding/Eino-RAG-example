package main

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
)

// newSplitter 用于文档的分割，故称其为Splitter组件，其根据某些特征来将文本切分，便于降低文本的粒度，让LLM能够生成更精细准确的内容。
func (r *RAGEngine) newSplitter(ctx context.Context) {
	t, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			// "#": "title",
			"#":   "h1", // 一级标题
			"##":  "h2", // 二级标题
			"###": "h3", // 三级标题
		},
		TrimHeaders: false,
	})
	if err != nil {
		r.Err = err
		return
	}
	r.Splitter = t
}
