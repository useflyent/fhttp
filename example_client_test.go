package http_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	http "github.com/zMrKrabz/fhttp"
	"github.com/zMrKrabz/fhttp/http2"
)

type testPushHandlerPushReadResponse struct {
	promise       *http.Request
	origReqURL    *url.URL
	origReqHeader http.Header
	push          *http.Response
	pushErr       error
	done          chan struct{}
}

func (ph *testPushHandlerPushReadResponse) HandlePush(r *http2.PushedRequest) {
	ph.promise = r.Promise
	ph.origReqURL = r.OriginalRequestURL
	ph.origReqHeader = r.OriginalRequestHeader
	ph.push, ph.pushErr = r.ReadResponse(r.Promise.Context())
	close(ph.done)
}

// Basic http test with Header Order + enable push
func TestExample(t *testing.T) {
	c := http.Client{}

	req, err := http.NewRequest("GET", "https://httpbin.org/headers", strings.NewReader(""))

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	req.Header = http.Header{
		"sec-ch-ua":                 {"\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
		"sec-ch-ua-mobile":          {"?0"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36"},
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-user":            {"?1"},
		"sec-fetch-dest":            {"document"},
		"accept-encoding":           {"gzip, deflate, br"},
		http.HeaderOrderKey: {
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"upgrade-insecure-requests",
			"user-agent",
			"accept",
			"sec-fetch-site",
			"sec-fetch-mode",
			"sec-fetch-user",
			"sec-fetch-dest",
			"accept-encoding",
		},
	}

	resp, err := c.Do(req)

	if err != nil {
		t.Errorf(err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %v", resp.StatusCode)
	}

	var data interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)

	if err != nil {
		t.Errorf(err.Error())
	}
}

func getCert() (*x509.CertPool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	caCert, err := os.ReadFile(fmt.Sprintf("%v/charles_cert.pem", home))
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)
	return certPool, nil
}

// Test with Charles cert + proxy
func TestWithCert(t *testing.T) {
	caCertPool, err := getCert()
	if err != nil {
		t.Fatalf(err.Error())
	}
	proxyURL, err := url.Parse("http://localhost:8888")
	if err != nil {
		t.Fatal(err.Error())
	}

	h1t := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
		Proxy:             http.ProxyURL(proxyURL),
		ForceAttemptHTTP2: true,
	}

	t2, err := http2.ConfigureTransports(h1t)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t2.Settings = []http2.Setting{
		{ID: http2.SettingMaxConcurrentStreams, Val: 1000},
		{ID: http2.SettingMaxFrameSize, Val: 16384},
		{ID: http2.SettingMaxHeaderListSize, Val: 262144},
	}
	t2.InitialWindowSize = 6291456
	t2.HeaderTableSize = 65536
	h1t.H2transport = t2

	client := http.Client{
		Transport: h1t,
	}

	req, err := http.NewRequest("GET", "https://httpbin.org/headers", nil)

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	req.Header = http.Header{
		"sec-ch-ua":                 {"\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
		"sec-ch-ua-mobile":          {"?0"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36", "I shouldn't be here"},
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-user":            {"?1"},
		"cookie":                    {"cf_clearance=67f509a97bae8bb8349523a14c0ca3d7d8460c93-1620778862-0-250", "wp_customerGroup=NOT+LOGGED+IN"},
		"sec-fetch-dest":            {"document"},
		"accept-encoding":           {"gzip, deflate, br"},
		"not-included-header":       {"should be last"},
		http.HeaderOrderKey: {
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"upgrade-insecure-requests",
			"user-agent",
			"cookie",
			"accept",
			"sec-fetch-site",
			"sec-fetch-mode",
			"sec-fetch-user",
			"sec-fetch-dest",
			"accept-encoding",
		},
		http.PHeaderOrderKey: {":method", ":authority", ":scheme", ":path"},
	}

	resp, err := client.Do(req)

	if err != nil {
		t.Errorf(err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %v", resp.StatusCode)
	}

	var data interface{}
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Error(err.Error())
	}
}

// Test with push handler
func TestEnablePush(t *testing.T) {
	caCertPool, err := getCert()
	if err != nil {
		t.Fatalf(err.Error())
	}
	proxyURL, err := url.Parse("http://localhost:8888")
	if err != nil {
		t.Fatal(err.Error())
	}
	t1 := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
		Proxy:             http.ProxyURL(proxyURL),
		ForceAttemptHTTP2: true,
	}
	t2, err := http2.ConfigureTransports(t1)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t2.PushHandler = &testPushHandlerPushReadResponse{}
	t1.H2transport = t2
	c := &http.Client{
		Transport: t1,
	}
	req, err := http.NewRequest("GET", "https://httpbin.org/headers", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	_, err = c.Do(req)
	if err != nil {
		t.Fatalf(err.Error())
	}

	req, err = http.NewRequest("POST", "https://httpbin.org/post", nil)
	_, err = c.Do(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

// Test with cookies
// Test with missing in header order, that should be added
// Test for UA that has empty string, excluding UA from being part of headers
