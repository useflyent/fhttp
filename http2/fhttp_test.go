package http2_test

import (
	"bytes"
	"fmt"
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

// Tests if connection settings are written correctly
func TestConnectionSettings(t *testing.T) {
	settings := []http2.Setting{
		{ID: http2.SettingHeaderTableSize, Val: 65536},
		{ID: http2.SettingMaxConcurrentStreams, Val: 1000},
		{ID: http2.SettingInitialWindowSize, Val: 6291456},
		{ID: http2.SettingMaxFrameSize, Val: 16384},
		{ID: http2.SettingMaxHeaderListSize, Val: 262144},
	}
	buf := new(bytes.Buffer)
	fr := http2.NewFramer(buf, buf)
	err := fr.WriteSettings(settings...)

	if err != nil {
		t.Fatalf(err.Error())
	}

	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatal(err.Error())
	}

	// dec := hpack.NewDecoder(2048, nil)
	// hf, err := dec.DecodeFull()
	// if err != nil {
	// 	t.Fatalf(err.Error())
	// }

	// for _, h := range hf {
	// 	fmt.Printf("%s\n", h.Name+h.Value)
	// }
	fmt.Println(f)
}
