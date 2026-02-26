package helpers

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func DoPost(url string, body any) (*http.Response, error) {
	b, _ := json.Marshal(body)
	return http.Post(url, "application/json", bytes.NewReader(b))
}

func DoGet(url string) (*http.Response, error) {
	return http.Get(url)
}
