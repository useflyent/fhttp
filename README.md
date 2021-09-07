# fhttp

# Features

## Ordered Headers

The package allows for both pseudo header order and normal header order. Most of the code is from this [this Pull Request](https://go-review.googlesource.com/c/go/+/105755/).

**Note on HTTP/1.1 header order**
Although the header key is capitalized, the header order slice must be in lowercase.
```go
	req.Header = http.Header{
		"X-NewRelic-ID":         {"12345"},
		"x-api-key":             {"ABCDE12345"},
		"MESH-Commerce-Channel": {"android-app-phone"},
		"mesh-version":          {"cart=4"},
		"X-Request-Auth":        {"hawkHeader"},
		"X-acf-sensor-data":     {"3456"},
		"Content-Type":          {"application/json; charset=UTF-8"},
		"Accept":                {"application/json"},
		"Transfer-Encoding":     {"chunked"},
		"Host":                  {"example.com"},
		"Connection":            {"Keep-Alive"},
		"Accept-Encoding":       {"gzip"},
		HeaderOrderKey: {
			"x-newrelic-id",
			"x-api-key",
			"mesh-commerce-channel",
			"mesh-version",
			"user-agent",
			"x-request-auth",
			"x-acf-sensor-data",
			"transfer-encoding",
			"content-type",
			"accept",
			"host",
			"connection",
			"accept-encoding",
		},
		PHeaderOrderKey: {
			":method",
			":path",
			":authority",
			":scheme",
		},
	}
```

## Connection settings

fhttp has Chrome-like connection settings, as shown below:

```text
SETTINGS_HEADER_TABLE_SIZE = 65536 (2^16)
SETTINGS_ENABLE_PUSH = 1
SETTINGS_MAX_CONCURRENT_STREAMS = 1000
SETTINGS_INITIAL_WINDOW_SIZE = 6291456
SETTINGS_MAX_FRAME_SIZE = 16384 (2^14)
SETTINGS_MAX_HEADER_LIST_SIZE = 262144 (2^18)
```

The default net/http settings, on the other hand, are the following:

```text
SETTINGS_HEADER_TABLE_SIZE = 4096
SETTINGS_ENABLE_PUSH = 0
SETTINGS_MAX_CONCURRENT_STREAMS = unlimited
SETTINGS_INITIAL_WINDOW_SIZE = 4194304
SETTINGS_MAX_FRAME_SIZE = 16384
SETTINGS_MAX_HEADER_LIST_SIZE = 10485760
```

The ENABLE_PUSH implementation was merged from [this Pull Request](https://go-review.googlesource.com/c/net/+/181497/).

## Custom HTTP2 Transport Flow Settings
To set a custom value for update frame, create a `http2.FlowSettings` and add it to the `Transport.FlowSettings` field. For example, to mimic the Chrome update frame, set `FrameSettings`'s `Conn` value to `15663105`.

## gzip, deflate, and br encoding

`gzip`, `deflate`, and `br` encoding are all supported by the package.

## Pseudo header order

fhttp supports pseudo header order for http2, helping mitigate fingerprinting. You can read more about how it works [here](https://www.akamai.com/uk/en/multimedia/documents/white-paper/passive-fingerprinting-of-http2-clients-white-paper.pdf).

## Backward compatible with net/http

Although this library is an extension of `net/http`, it is also meant to be backward compatible. Replacing

```go
import (
   "net/http"
)
```

with

```go
import (
    http "github.com/useflyent/fhttp"
)
```

SHOULD not break anything.

## Versatile Content-Length and Transfer-Encoding headers

fhttp user to set custom Content-Length and Transfer-Encoding headers of all types.

### To set an empty Content-Length header
```go
req.Header = http.Header{
	"Content-Length": {http.ContentLengthEmpty},
}
```

### To ignore setting the Content-Length header
```go
req.Header = http.Header{
    "Content-Length": {http.ContentLengthDelete},
}
```

Any Content-Length or Transfer-Encoding headers set will be prioritized and fhttp will not set proper Content-length or Transfer-Encoding headers

## Credits

Special thanks to the following people for helping me with this project.

* [cc](https://github.com/x04/) for guiding me when I first started this project and inspiring me with [cclient](https://github.com/x04/cclient)

* [umasi](https://github.com/umasii) for being good rubber ducky and giving me tips for http2 headers

* [car](https://github.com/Carcraftz) for pointing the custom http2 update frame to me
