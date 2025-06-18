package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ollama/ollama/api"
)

func embdding_text(text string) error {

	httpclient := &http.Client{}
	u, err := url.Parse("http://192.168.30.59:11434")
	if err != nil {
		return err
	}
	cli := api.NewClient(u, httpclient)

	req := &api.EmbeddingRequest{
		Model:  "tazarov/all-minilm-l6-v2-f32", // "mxbai-embed-large"
		Prompt: text,
	}
	resp, err := cli.Embeddings(context.Background(), req)
	// resp.Embedding is a []float64 with 4096 entries
	if err != nil {
		return err
	}
	log.Printf("%#v", resp.Embedding)

	// write the embedding to pgvector
	connString := "postgres://pgvector:pgvector@192.168.30.59:54333/ragdb?sslmode=disable"
	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		return err
	}
	defer dbpool.Close()

	sqltext := `
	INSERT INTO documents (content, embedding)
	VALUES ($1, $2::vector)`
	_, err = dbpool.Exec(context.Background(), sqltext, req.Prompt, floats64ToString(resp.Embedding))
	if err != nil {
		fmt.Printf("failed to insert document: %s\n", err)
		return err
	}

	return nil
}

func floats32ToString(floats []float32) string {
	strVals := make([]string, len(floats))
	for i, val := range floats {
		strVals[i] = fmt.Sprintf("%f", val)
	}
	joined := strings.Join(strVals, ", ")
	return "[" + joined + "]"
}

func floats64ToString(floats []float64) string {
	strVals := make([]string, len(floats))
	for i, val := range floats {
		strVals[i] = fmt.Sprintf("%f", val)
	}
	joined := strings.Join(strVals, ", ")
	return "[" + joined + "]"
}

func searchSimilarDocuments(query string, k int) ([]string, error) {

	httpclient := &http.Client{}
	u, err := url.Parse("http://192.168.30.59:11434")
	if err != nil {
		return nil, err
	}
	cli := api.NewClient(u, httpclient)

	req := &api.EmbeddingRequest{
		Model:  "tazarov/all-minilm-l6-v2-f32", // "mxbai-embed-large"
		Prompt: query,
	}
	resp, err := cli.Embeddings(context.Background(), req)
	// resp.Embedding is a []float64 with 4096 entries
	if err != nil {
		return nil, err
	}
	log.Printf("resp len %#v", len(resp.Embedding))

	connString := "postgres://pgvector:pgvector@192.168.30.59:54333/ragdb?sslmode=disable"
	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		return nil, err
	}
	defer dbpool.Close()

	sqltext := fmt.Sprintf(`
        SELECT content
        FROM documents
        ORDER BY embedding <-> $1
        LIMIT $2;
    `)
	rows, err := dbpool.Query(context.Background(), sqltext, floats64ToString(resp.Embedding), k)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()
	fmt.Printf("rows: %#v\n", rows)

	var contents []string
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		contents = append(contents, content)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	fmt.Printf("Found %d similar documents:\n", len(contents))
	fmt.Printf("%#v\n", contents)
	return contents, nil
}

/* pgvector

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS documents (
        id SERIAL PRIMARY KEY,
        content TEXT,
        embedding vector(1536)
);

CREATE INDEX IF NOT EXISTS documents_embedding_idx
ON documents USING ivfflat (embedding vector_l2_ops) WITH (lists = 100);

*/
