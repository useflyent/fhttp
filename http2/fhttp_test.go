package http2_test

import (
	"bytes"
	"errors"
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

func compareSettings(ID http2.SettingID, output uint32, expected uint32) error {
	if output != expected {
		return errors.New(fmt.Sprintf("Setting %v, expected %d got %d", ID, expected, output))
	}
	return nil
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

	sf := f.(*http2.SettingsFrame)
	n := sf.NumSettings()
	if n != len(settings) {
		t.Fatalf("Expected %d settings, got %d", len(settings), n)
	}

	for i := 0; i < n; i++ {
		s := sf.Setting(i)
		var err error
		switch s.ID {
		case http2.SettingHeaderTableSize:
			err = compareSettings(s.ID, s.Val, 65536)
		case http2.SettingMaxConcurrentStreams:
			err = compareSettings(s.ID, s.Val, 1000)
		case http2.SettingInitialWindowSize:
			err = compareSettings(s.ID, s.Val, 6291456)
		case http2.SettingMaxFrameSize:
			err = compareSettings(s.ID, s.Val, 16384)
		case http2.SettingMaxHeaderListSize:
			err = compareSettings(s.ID, s.Val, 262144)
		}

		if err != nil {
			t.Fatal(err.Error())
		}
	}
}

// Round trip test, makes sure that the changes made doesn't break the library
func TestRoundTrip(t *testing.T) {
	settings := []http2.Setting{
		{ID: http2.SettingHeaderTableSize, Val: 65536},
		{ID: http2.SettingMaxConcurrentStreams, Val: 1000},
		{ID: http2.SettingInitialWindowSize, Val: 6291456},
		{ID: http2.SettingMaxFrameSize, Val: 16384},
		{ID: http2.SettingMaxHeaderListSize, Val: 262144},
	}
	tr := http2.Transport{
		Settings: settings,
	}
	req, err := http.NewRequest("GET", "www.google.com", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	tr.RoundTrip(req)
}
