package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/philippgille/chromem-go"
)

func testdoc1() {
	// Set up chromem-go in-memory, for easy prototyping. Can add persistence easily!
	// We call it DB instead of client because there's no client-server separation. The DB is embedded.
	db := chromem.NewDB()

	/*
		使用Ollama模型进行嵌入，类似命令
		curl http://localhost:11434/api/embeddings -d '{
		  "model": "tazarov/all-minilm-l6-v2-f32",
		  "prompt": "Llamas are members of the camelid family"
		}'
	*/
	embeddingFunc := chromem.NewEmbeddingFuncOllama(
		"tazarov/all-minilm-l6-v2-f32",
		"http://192.168.30.59:11434/api")

	// Create collection. GetCollection, GetOrCreateCollection, DeleteCollection also available!
	collection, _ := db.CreateCollection("all-my-documents", nil, embeddingFunc)

	// Add docs to the collection. Update and delete will be added in the future.
	// Can be multi-threaded with AddConcurrently()!
	// We're showing the Chroma-like method here, but more Go-idiomatic methods are also available!
	_ = collection.Add(context.Background(),
		[]string{"doc1", "doc2"}, // unique ID for each doc
		nil,                      // We handle embedding automatically. You can skip that and add your own embeddings as well.
		[]map[string]string{{"source": "notion"}, {"source": "google-docs"}}, // Filter on these!
		[]string{"This is document1", "This is document2"},
	)

	// Query/search 2 most similar results. You can also get by ID.
	results, err := collection.Query(context.Background(),
		"This is a query document",
		2,
		map[string]string{"metadata_field": "is_equal_to_this"}, // optional filter
		map[string]string{"$contains": "search_string"},         // optional filter
	)
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("results: %#v\n", results)
	}

}

func testdoc2() {
	ctx := context.Background()

	db := chromem.NewDB()
	/*
		使用Ollama模型进行嵌入，类似命令
		curl http://localhost:11434/api/embeddings -d '{
		  "model": "tazarov/all-minilm-l6-v2-f32",
		  "prompt": "Llamas are members of the camelid family"
		}'
	*/
	embeddingFunc := chromem.NewEmbeddingFuncOllama(
		"tazarov/all-minilm-l6-v2-f32",
		"http://192.168.30.59:11434/api")

	c, err := db.CreateCollection("knowledge-base", nil, embeddingFunc)
	if err != nil {
		panic(err)
	}

	err = c.AddDocuments(ctx, []chromem.Document{
		{
			ID:      "1",
			Content: "天空呈蓝色是因为瑞利散射。",
		},
		{
			ID:      "2",
			Content: "树叶是绿色的因为叶绿素吸收红光和蓝光。",
		},
	}, runtime.NumCPU())
	if err != nil {
		panic(err)
	}

	res, err := c.Query(ctx, "为什么天空是蓝色的？", 1, nil, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("ID: %v\n相似度: %v\n内容: %v\n", res[0].ID, res[0].Similarity, res[0].Content)
}
