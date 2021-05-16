# fhttp 
The f stands for flex. fhttp is a fork of net/http that provides an array of features pertaining to the fingerprint of the golang http client. Through these changes, the http client becomes much more flexible, and when combined with transports such as [uTLS](https://github.com/refraction-networking/utls) can mitigate servers from fingerprinting requests and see that it is made by golang, making them look like they originate from a regular chrome browser.

Documentation can be contributed, otherwise, look at tests and examples. The main one should be [example_client_test.go](example_client_test.go).

# Features
## Ordered Headers
Allows for pseudo header order and normal header order. Most of the code is taken from [this pr](https://go-review.googlesource.com/c/go/+/105755/).

## Connection settings (TODO)
Has Chrome-like connection settings:
```
SETTINGS_HEADER_TABLE_SIZE
```
(will add the rest later, as i can't connect to [notion](https://www.notion.so/) from my computer)

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
	http "github.com/zMrKrabz/fhttp"
)
```

SHOULD not break anything. 
