package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ollama/ollama/api"
)

/*
 * Embedding use ollama api to create embeddings
 * Ollama url like http://localhost:11434
 * Use embedding model tazarov/all-minilm-l6-v2-f32
 */
type Embedding struct {
	OllamaUrl string `json:"ollama_url"`
	Model     string `json:"model"`
}

// embedding text to embedding vector with model of ollama
func (p *Embedding) Embed(text string) ([]float64, error) {
	// Create a new client to ollama url
	httpclient := &http.Client{}
	u, err := url.Parse(p.OllamaUrl)
	if err != nil {
		return nil, err
	}
	cli := api.NewClient(u, httpclient)

	req := &api.EmbeddingRequest{
		Model:  "tazarov/all-minilm-l6-v2-f32", // "mxbai-embed-large"
		Prompt: text,
	}

	// resp.Embedding is a []float64 with 4096 entries
	resp, err := cli.Embeddings(context.Background(), req)
	if err != nil {
		return nil, err
	}
	// log.Printf("%#v", resp.Embedding)

	return resp.Embedding, nil
}

func Floats32ToString(floats []float32) string {
	strVals := make([]string, len(floats))
	for i, val := range floats {
		strVals[i] = fmt.Sprintf("%f", val)
	}
	joined := strings.Join(strVals, ", ")
	return "[" + joined + "]"
}

func Floats64ToString(floats []float64) string {
	strVals := make([]string, len(floats))
	for i, val := range floats {
		strVals[i] = fmt.Sprintf("%f", val)
	}
	joined := strings.Join(strVals, ", ")
	return "[" + joined + "]"
}
