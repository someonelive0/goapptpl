package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

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
func CallLLM(document, query string) error {
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
