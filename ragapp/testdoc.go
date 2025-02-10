package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
		[]string{"doc1", "doc2", "doc3", "doc4", "doc5", "doc6"}, // unique ID for each doc
		nil, // We handle embedding automatically. You can skip that and add your own embeddings as well.
		[]map[string]string{
			{"source": "notion"},
			{"source": "google-docs"},
			{"source": "google-docs"},
			{"source": "google-docs"},
			{"source": "google-docs"},
			{"source": "google-docs"},
		}, // Filter on these!
		[]string{
			"Llamas are members of the camelid family, meaning they're closely related to vicuñas and camels.",
			"Llamas were first domesticated and used as pack animals 4,000 to 5,000 years ago in the Peruvian highlands.",
			"Llamas can grow as much as 6 feet tall, though the average llama is between 5 feet 6 inches and 5 feet 9 inches tall.",
			"Llamas weigh between 280 and 450 pounds and can carry 25 to 30 percent of their body weight.",
			"Llamas are vegetarians and have very efficient digestive systems.",
			"Llamas live to be about 20 years old, though some live up to 30 years."},
	)
	documents_collections := db.GetCollection("all-my-documents", embeddingFunc)
	fmt.Printf("all-my-documents count: %d\n", documents_collections.Count())

	// Query/search 2 most similar results. You can also get by ID.
	results, err := collection.Query(context.Background(),
		"What animals are llamas related to?",
		2,
		nil, // map[string]string{"metadata_field": "is_equal_to_this"}, // optional filter
		nil, // map[string]string{"$contains": "search_string"},         // optional filter
	)
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("What animals are llamas related to? results: %#v\n", len(results))
		for i, result := range results {
			fmt.Printf("result %d> %s %f: %s\n", i, result.ID, result.Similarity, result.Content)
		}
	}

}

// 测试中文问答
func testdoc2() {
	fmt.Printf("\ntestdoc2: =====================\n")

	ctx := context.Background()
	embeddingFunc := chromem.NewEmbeddingFuncOllama(
		"tazarov/all-minilm-l6-v2-f32",
		"http://192.168.30.59:11434/api")

	// db := chromem.NewDB()
	db, err := chromem.NewPersistentDB("ragapp/db", false)
	if err != nil {
		panic(err)
	}
	collections := db.ListCollections()
	fmt.Printf("collections: %#v\n", collections)
	knowledge := db.GetCollection("knowledge-base", embeddingFunc)
	fmt.Printf("knowledge-base count: %d\n", knowledge.Count())
	knowledge.Query(ctx, "天空是蓝色的原因是什么？", 1, nil, nil)
	res, err := knowledge.Query(ctx, "为什么天空是蓝色的？", 1, nil, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("ID: %v\n相似度: %v\n内容: %v\n", res[0].ID, res[0].Similarity, res[0].Content)

	fmt.Println("使用新建collection")

	/*
		使用Ollama模型进行嵌入，类似命令
		curl http://localhost:11434/api/embeddings -d '{
		  "model": "tazarov/all-minilm-l6-v2-f32",
		  "prompt": "Llamas are members of the camelid family"
		}'
	*/

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

	res, err = c.Query(ctx, "为什么天空是蓝色的？", 1, nil, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("ID: %v\n相似度: %v\n内容: %v\n", res[0].ID, res[0].Similarity, res[0].Content)

	callDeepSeek(res[0].Content, "为什么天空是蓝色的？")
}

/*
调用Ollama的API生成应答
"Using this information: {retrieved_document}, respond to the query: {query}"

	curl -X POST http://localhost:11434/api/generate -d '{
	  "model": "deepseek-custom",
	  "prompt": "Using this information: {retrieved_document}, respond to the query: {query}"
	}'

	响应数据是多行JSON，一行如下，每行表示一个生成的结果，其中"done"字段表示是否完成
	{"model":"deepseek-r1:1.5b","created_at":"2025-02-10T08:31:17.571108076Z","response":"\n","done":false}
*/
func callDeepSeek(document, query string) error {
	url := "http://192.168.30.59:11434/api/generate"
	// url := "http://localhost:11434/api/generate"
	contentType := "application/json"
	prompt := fmt.Sprintf("Using this information: '%s', respond to the query: '%s'", document, query)
	// llama3.1:latest deepseek-r1:1.5b
	jsonData := fmt.Sprintf(`{
		"model": "deepseek-r1:1.5b",
		"prompt": "%s"
	}`, prompt)
	fmt.Printf("prompt: %s\n", prompt)

	resp, err := http.Post(url, contentType, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)

	var result map[string]interface{}
	for scanner.Scan() {
		line := scanner.Text()
		if err = json.Unmarshal([]byte(line), &result); err != nil {
			fmt.Println("json.Unmarshal failed:", err)
			return err
		}
		if result["done"].(bool) {
			fmt.Printf("\ndone\n")
		} else {
			fmt.Printf("%s", result["response"].(string))
		}
		// fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("scanner error:", err)
	}

	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Println("Error reading response body:", err)
	// 	return err
	// }

	// fmt.Println("Response Status:", resp.Status)
	// fmt.Println("Response Body:", string(body))

	return nil
}
