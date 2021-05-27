package http2_test

import (
	"strings"
	"testing"

	http "github.com/zMrKrabz/fhttp"
	"github.com/zMrKrabz/fhttp/http2"
	"github.com/zMrKrabz/fhttp/httptrace"
)

func TestHeaderOrder(t *testing.T) {
	hk := ""

	trace := &httptrace.ClientTrace{
		WroteHeaderField: func(key string, values []string) {
			hk += key + " "
		},
	}
	req, err := http.NewRequest("GET", "https://httpbin.org/#/Request_inspection/get_headers", nil)
	req.Header.Add("experience", "pain")
	req.Header.Add("grind", "harder")
	req.Header.Add("live", "mas")
	req.Header[http.HeaderOrderKey] = []string{"grind", "experience", "live"}
	req.Header[http.PHeaderOrderKey] = []string{":method", ":authority", ":scheme", ":path"}

	if err != nil {
		t.Fatalf(err.Error())
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	tr := &http2.Transport{}

	_, err = tr.RoundTrip(req)

	if err != nil {
		t.Fatal(err.Error())
	}

	eq := strings.EqualFold(hk, ":method :authority :scheme :path grind experience live accept-encoding user-agent ")
	if !eq {
		t.Fatalf("Header order not set properly, \n Got %v \n Want: %v", hk, ":method :authority :scheme :path grind experience live accept-encoding user-agent")
	}
}
