# fhttp 

**NOTE**
This maintenance of this library has move over to [userflynet](https://github.com/useflyent/fhttp). The only use for this repository is so imports will not break.

The f stands for flex. fhttp is a fork of net/http that provides an array of features pertaining to the fingerprint of the golang http client. Through these changes, the http client becomes much more flexible, and when combined with transports such as [uTLS](https://github.com/refraction-networking/utls) can mitigate servers from fingerprinting requests and see that it is made by golang, making them look like they originate from a regular chrome browser.

Documentation can be contributed, otherwise, look at tests and examples. The main one should be [example_client_test.go](example_client_test.go).

# Features

## Ordered Headers
Allows for pseudo header order and normal header order. Most of the code is taken from [this pr](https://go-review.googlesource.com/c/go/+/105755/).

## Connection settings
Has Chrome-like connection settings:
```
SETTINGS_HEADER_TABLE_SIZE = 65536 (2^16)
SETTINGS_ENABLE_PUSH = 1
SETTINGS_MAX_CONCURRENT_STREAMS = 1000
SETTINGS_INITIAL_WINDOW_SIZE = 6291456
SETTINGS_MAX_FRAME_SIZE = 16384 (2^14)
SETTINGS_MAX_HEADER_LIST_SIZE = 262144 (2^18)
```

Default net/http settings:
```
SETTINGS_HEADER_TABLE_SIZE = 4096
SETTINGS_ENABLE_PUSH = 0
SETTINGS_MAX_CONCURRENT_STREAMS = unlimited
SETTINGS_INITIAL_WINDOW_SIZE = 4194304
SETTINGS_MAX_FRAME_SIZE = 16384
SETTINGS_MAX_HEADER_LIST_SIZE = 10485760
```

ENABLE_PUSH implementation was merged from [this pull request](https://go-review.googlesource.com/c/net/+/181497/)

## gzip, deflate, br encoding 
Actually supports and implements encoding `gzip, deflate, br`

## Pseudo header order
Supports pseudo header order for http2 to mitigate fingerprinting. Read more about it [here](https://www.akamai.com/uk/en/multimedia/documents/white-paper/passive-fingerprinting-of-http2-clients-white-paper.pdf)

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

# Credits
Special thanks to the following people for helping me with this project.

* [cc](https://github.com/x04/) for guiding me when I first started this project and inspiring me with [cclient](https://github.com/x04/cclient)

* [umasi](https://github.com/umasii) for being good rubber ducky and giving me tips for http2 headers
