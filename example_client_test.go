package http_test

import (
	"strings"
	"testing"

	http "github.com/zMrKrabz/fhttp"
)

func TestExample(t *testing.T) {
	c := http.Client{}
	req, err := http.NewRequest("GET", "https://www.topps.com/", strings.NewReader(""))

	if err != nil {
		t.Errorf(err.Error())
	}

	resp, err := c.Do(req)

	if err != nil {
		t.Errorf(err.Error())
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %v", resp.StatusCode)
	}
}
