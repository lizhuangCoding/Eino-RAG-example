package main

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
)

// newLoader 读取文件的内容，转换成Eino框架的schema.Document
func (r *RAGEngine) newLoader(ctx context.Context) {
	l, err := file.NewFileLoader(ctx, &file.FileLoaderConfig{
		UseNameAsID: true,
		Parser:      nil,
	})
	if err != nil {
		r.Err = err
		return
	}
	r.Loader = l
}
