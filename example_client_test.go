package http_test

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"testing"

	http "github.com/useflyent/fhttp"
	"github.com/useflyent/fhttp/http2"
)

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

func getCharlesCert() (*x509.CertPool, error) {
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

func addCharlesToTransport(tr *http.Transport, proxy string) error {
	caCertPool, err := getCharlesCert()
	if err != nil {
		return err
	}
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return err
	}
	tr.TLSClientConfig = &tls.Config{
		RootCAs: caCertPool,
	}
	tr.Proxy = http.ProxyURL(proxyURL)

	return nil
}

func addWiresharkToTransport(tr *http.Transport) error {
	kl := flag.String("keylog", "ssl-keylog.txt", "file to dump ssl keys")
	keylog, err := os.OpenFile(*kl, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
		KeyLogWriter:       keylog,
	}
	return nil
}

// Test with Charles cert + proxy
func TestWithCert(t *testing.T) {
	h1t := &http.Transport{
		ForceAttemptHTTP2: true,
	}
	if err := addCharlesToTransport(h1t, "http://localhost:8888"); err != nil {
		t.Fatalf(err.Error())
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

	req, err := http.NewRequest("GET", "https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding", nil)

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
}

func TestHTTP1HeaderOrder(t *testing.T) {
	req, err := http.NewRequest("GET", "https://prod.jdgroupmesh.cloud/stores/size/products/16069871?channel=android-app-phone&expand=variations,informationBlocks,customisations", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	req.Header = http.Header{
		"X-NewRelic-ID":         {"12345"},
		"x-api-key":             {"ABCDE12345"},
		"MESH-Commerce-Channel": {"android-app-phone"},
		"mesh-version":          {"cart=4"},
		"User-Agent":            {"size/3.1.0.8355 (android-app-phone; Android 10; Build/CPH2185_11_A.28)"},
		"X-Request-Auth":        {"hawkHeader"},
		"X-acf-sensor-data":     {"3456"},
		"Content-Type":          {"application/json; charset=UTF-8"},
		"Accept":                {"application/json"},
		"Transfer-Encoding":     {"chunked"},
		"Host":                  {"prod.jdgroupmesh.cloud"},
		"Connection":            {"Keep-Alive"},
		"Accept-Encoding":       {"gzip"},
		http.HeaderOrderKey: {
			"X-NewRelic-ID",
			"x-api-key",
			"MESH-Commerce-Channel",
			"mesh-version",
			"User-Agent",
			"X-Request-Auth",
			"X-acf-sensor-data",
			"Transfer-Encoding",
			"Content-Type",
			"Accept",
			"Host",
			"Connection",
			"Accept-Encoding",
		},
		http.PHeaderOrderKey: {
			":method",
			":path",
			":authority",
			":scheme",
		},
	}

	var b []byte
	buf := bytes.NewBuffer(b)
	err = req.Write(buf)
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected := "GET /stores/size/products/16069871?channel=android-app-phone&expand=variations,informationBlocks,customisations HTTP/1.1\r\nX-NewRelic-ID: 12345\r\nx-api-key: ABCDE12345\r\nMESH-Commerce-Channel: android-app-phone\r\nmesh-version: cart=4\r\nUser-Agent: size/3.1.0.8355 (android-app-phone; Android 10; Build/CPH2185_11_A.28)\r\nX-Request-Auth: hawkHeader\r\nX-acf-sensor-data: 3456\r\nTransfer-Encoding: chunked\r\nContent-Type: application/json; charset=UTF-8\r\nAccept: application/json\r\nHost: prod.jdgroupmesh.cloud\r\nConnection: Keep-Alive\r\nAccept-Encoding: gzip\r\n\r\n"
	if expected != buf.String() {
		t.Fatalf("got: %swant: %s", buf.String(), expected)
	}
}

// Test with push handler
func TestEnablePush(t *testing.T) {
	t1 := &http.Transport{
		ForceAttemptHTTP2: true,
	}
	t2, err := http2.ConfigureTransports(t1)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t2.PushHandler = &http2.DefaultPushHandler{}
	t1.H2transport = t2
	c := &http.Client{
		Transport: t1,
	}
	var req *http.Request
	req, err = http.NewRequest("GET", "https://httpbin.org/headers", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	_, err = c.Do(req)
	if err != nil {
		t.Fatalf(err.Error())
	}

	req, err = http.NewRequest("POST", "https://httpbin.org/post", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	_, err = c.Do(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

// Test finishline
func TestFinishLine(t *testing.T) {
	t1 := &http.Transport{
		ForceAttemptHTTP2: true,
	}

	if err := addCharlesToTransport(t1, "http://localhost:8888"); err != nil {
		t.Fatalf(err.Error())
	}
	// if err := addWiresharkToTransport(t1); err != nil {
	// 	t.Fatalf(err.Error())
	// }
	t2, err := http2.ConfigureTransports(t1)
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
	t2.PushHandler = &http2.DefaultPushHandler{}
	t1.H2transport = t2

	c := &http.Client{
		Transport: t1,
	}
	req, err := http.NewRequest("GET", "https://www.finishline.com/", nil)
	if err != nil {
		t.Fatalf(err.Error())
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
		http.PHeaderOrderKey: {":method", ":authority", ":scheme", ":path"},
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("Got status %v from finishline, expected 200", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Printf("resp: %v\n", string(b)[1])
}

// Test compression brotli
func TestCompressionBrotli(t *testing.T) {
	t1 := &http.Transport{
		ForceAttemptHTTP2: true,
	}
	c := http.Client{
		Transport: t1,
	}
	req, _ := http.NewRequest("GET", "https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding", nil)
	req.Header = http.Header{
		"accept-encoding": {"br"},
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if h := resp.Header.Get("content-encoding"); h == "" || h != "br" {
		t.Fatalf("Got content-encoding header %v, expected br", h)
	}
}

// Test compression zlib deflate
func TestCompressionZlibDeflate(t *testing.T) {
	t1 := &http.Transport{
		ForceAttemptHTTP2: true,
	}
	addCharlesToTransport(t1, "http://localhost:8888")
	c := http.Client{
		Transport: t1,
	}
	req, _ := http.NewRequest("GET", "http://carsten.codimi.de/gzip.yaws/daniels.html?deflate=on&zlib=on", nil)
	req.Header = http.Header{
		"accept-encoding": {"deflate"},
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if h := resp.Header.Get("content-encoding"); h == "" || h != "deflate" {
		t.Fatalf("Expected content encoding deflate, got %v", h)
	}
}

// Test compression deflate
func TestCompressionDeflate(t *testing.T) {
	c := http.Client{}
	req, _ := http.NewRequest("GET", "http://carsten.codimi.de/gzip.yaws/daniels.html?deflate=on", nil)
	req.Header = http.Header{
		"accept-encoding": {"deflate"},
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if h := resp.Header.Get("content-encoding"); h == "" || h != "deflate" {
		t.Fatalf("Expected content encoding deflate, got %v", h)
	}
}

// Test with cookies
// Test with missing in header order, that should be added
// Test for UA that has empty string, excluding UA from being part of headers
