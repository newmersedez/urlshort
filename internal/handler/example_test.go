package handler_test

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Example_shortenURL демонстрирует сокращение URL через POST / с типом text/plain.
// Сервер возвращает 201 Created и короткую ссылку в теле ответа.
func Example_shortenURL() {
	srv, client, err := newExampleServer()
	if err != nil {
		panic(err)
	}
	defer srv.Close()

	resp, err := client.Post(
		srv.URL+"/", "text/plain",
		strings.NewReader("https://practicum.yandex.ru"),
	)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	// Output:
	// 201
}

// Example_shortenURLJSON демонстрирует сокращение URL через POST /api/shorten с JSON.
// Сервер возвращает 201 Created и JSON с полем result.
func Example_shortenURLJSON() {
	srv, client, err := newExampleServer()
	if err != nil {
		panic(err)
	}
	defer srv.Close()

	body := `{"url":"https://go.dev/doc/effective_go"}`
	resp, err := client.Post(
		srv.URL+"/api/shorten", "application/json",
		strings.NewReader(body),
	)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	// Output:
	// 201
}

// Example_shortenBatchURLs демонстрирует пакетное сокращение URL через POST /api/shorten/batch.
// Каждый элемент содержит correlation_id для сопоставления запроса и ответа.
func Example_shortenBatchURLs() {
	srv, client, err := newExampleServer()
	if err != nil {
		panic(err)
	}
	defer srv.Close()

	body := `[
		{"correlation_id":"1","original_url":"https://pkg.go.dev"},
		{"correlation_id":"2","original_url":"https://go.dev"}
	]`

	resp, err := client.Post(
		srv.URL+"/api/shorten/batch", "application/json",
		strings.NewReader(body),
	)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	// Output:
	// 201
}

// Example_redirect демонстрирует переход по короткой ссылке через GET /{id}.
// Сначала URL сокращается, затем выполняется GET-запрос.
// Сервер возвращает 307 Temporary Redirect с заголовком Location.
func Example_redirect() {
	srv, client, err := newExampleServer()
	if err != nil {
		panic(err)
	}
	defer srv.Close()

	// Сокращаем URL и получаем короткую ссылку из тела ответа.
	resp, err := client.Post(
		srv.URL+"/", "text/plain",
		strings.NewReader("https://golang.org"),
	)
	if err != nil {
		panic(err)
	}
	shortURL, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// Извлекаем путь короткой ссылки и переходим по нему.
	shortPath := strings.TrimPrefix(strings.TrimSpace(string(shortURL)), "http://localhost:8080")
	redirectResp, err := client.Get(srv.URL + shortPath)
	if err != nil {
		panic(err)
	}
	defer redirectResp.Body.Close()

	fmt.Println(redirectResp.StatusCode)
	fmt.Println(redirectResp.Header.Get("Location"))
	// Output:
	// 307
	// https://golang.org
}

// Example_getUserURLs демонстрирует получение списка URL пользователя через GET /api/user/urls.
// После сокращения двух ссылок сервер возвращает JSON-массив.
func Example_getUserURLs() {
	srv, client, err := newExampleServer()
	if err != nil {
		panic(err)
	}
	defer srv.Close()

	for _, u := range []string{"https://example.com", "https://example.org"} {
		resp, _ := client.Post(srv.URL+"/", "text/plain", strings.NewReader(u))
		resp.Body.Close()
	}

	resp, err := client.Get(srv.URL + "/api/user/urls")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var urls []struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
	json.NewDecoder(resp.Body).Decode(&urls)

	fmt.Println(resp.StatusCode)
	fmt.Println(len(urls))
	// Output:
	// 200
	// 2
}
