package services

import (
	"io"
	"net/http"
	"net/url"
)

func CheckGrammar(text string) string {
	apiURL := "https://api.languagetool.org/v2/check"
	data := url.Values{}
	data.Set("text", text)
	data.Set("language", "en-US")

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return "Grammar check failed."
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Grammar check failed."
	}

	return string(body)
}
