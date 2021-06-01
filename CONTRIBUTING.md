# Contributing
This serves as a todo list for the library, anybody can contribute. For a branch to be merged, first open a pull request. For issues not on this page, open an issue first to see if the issue is within scope of this library.

## gzip, deflate, br
The `gzip, deflate, br` encoding should be implemented as an opt-in way for the client to use this instead of the standard gzip encoding. A helpful example can be found [here](https://play.golang.org/p/80HukFxfs4).

## Writing better tests
* Test fingerprinting bypass with this [site](https://privacycheck.sec.lrz.de/passive/fp_h2/fp_http2.html#fpDemoHttp2)
* Test for ENABLE_PUSH implementation, from [here](https://go-review.googlesource.com/c/net/+/181497/)
* Test for all features mentioned in [README](README.md), such as header order and pheader order with httptrace

## Create a server fingerprint implementation
Will be able to use library in order to check fingerprint of incoming requests, and see what http2 setting is missing or wrong

## Fix push handler errors
The push handler has errors with reading responses, specifically, it will sometimes fail because it read to EOF, or `Client closed connection before receiving entire response`. Someone with knowledge of how pushed requests are sent and read should fix this issue, or see if something was copied wrong when implementing [the pull request](https://go-review.googlesource.com/c/net/+/181497/)

## Merging upstream
When changes are made by the golang team on the [http]() or [http2](https://pkg.go.dev/golang.org/x/net/http2) library as a release branch,  
```
git remote add -f golang git@github.com:golang/go.git
git checkout -b golang-upstream golang/master
git subtree split -P src/crypto/tls/ -b golang-tls-upstream
git checkout master
git merge --no-commit golang-<http or http2>-upstream
```
