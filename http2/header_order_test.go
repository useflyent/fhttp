package http2

import (
	"bytes"
	http "github.com/useflyent/fhttp"
	"github.com/useflyent/fhttp/httptrace"
	"strings"
	"testing"
)

func TestHeaderOrder(t *testing.T) {
	req, err := http.NewRequest("POST", "https://www.httpbin.org/headers", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	req.Header = http.Header{
		"sec-ch-ua":        {"\" Not;A Brand\";v=\"99\", \"Google Chrome\";v=\"91\", \"Chromium\";v=\"91\""},
		"accept":           {"*/*"},
		"x-requested-with": {"XMLHttpRequest"},
		"sec-ch-ua-mobile": {"?0"},
		"user-agent":       {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36\", \"I shouldn't be here"},
		"content-type":     {"application/json"},
		"origin":           {"https://www.size.co.uk/"},
		"sec-fetch-site":   {"same-origin"},
		"sec-fetch-mode":   {"cors"},
		"sec-fetch-dest":   {"empty"},
		"accept-language":  {"en-US,en;q=0.9"},
		"accept-encoding":  {"gzip, deflate, br"},
		"referer":          {"https://www.size.co.uk/product/white-jordan-air-1-retro-high/16077886/"},
		http.HeaderOrderKey: {
			"sec-ch-ua",
			"accept",
			"x-requested-with",
			"sec-ch-ua-mobile",
			"user-agent",
			"content-type",
			"sec-fetch-site",
			"sec-fetch-mode",
			"sec-fetch-dest",
			"referer",
			"accept-encoding",
			"accept-language",
		},
		http.PHeaderOrderKey: {
			":method",
			":authority",
			":scheme",
			":path",
		},
	}
	var b []byte
	buf := bytes.NewBuffer(b)
	err = req.Header.Write(buf)
	if err != nil {
		t.Fatalf(err.Error())
	}
	arr := strings.Split(buf.String(), "\n")
	var hdrs []string
	for _, v := range arr {
		a := strings.Split(v, ":")
		if a[0] == "" {
			continue
		}
		hdrs = append(hdrs, a[0])
	}

	for i := range req.Header[http.HeaderOrderKey] {
		if hdrs[i] != req.Header[http.HeaderOrderKey][i] {
			t.Errorf("want: %s\ngot: %s\n", req.Header[http.HeaderOrderKey][i], hdrs[i])
		}
	}
}

func TestHeaderOrder2(t *testing.T) {
	hk := ""
	trace := &httptrace.ClientTrace{
		WroteHeaderField: func(key string, values []string) {
			hk += key + " "
		},
	}
	req, err := http.NewRequest("GET", "https://httpbin.org/#/Request_inspection/get_headers", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	req.Header.Add("experience", "pain")
	req.Header.Add("grind", "harder")
	req.Header.Add("live", "mas")
	req.Header[http.HeaderOrderKey] = []string{"grind", "experience", "live"}
	req.Header[http.PHeaderOrderKey] = []string{":method", ":authority", ":scheme", ":path"}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	tr := &Transport{}
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer resp.Body.Close()

	eq := strings.EqualFold(hk, ":method :authority :scheme :path grind experience live accept-encoding user-agent ")
	if !eq {
		t.Fatalf("Header order not set properly, \n Got %v \n Want: %v", hk, ":method :authority :scheme :path grind experience live accept-encoding user-agent")
	}
}
