# Contributing
This serves as a todo list for the library, anybody can contribute. For a branch to be merged, first open a pull request. For issues not on this page, open an issue first to see if the issue is within scope of this library.

## Migrating h2_bundle to http2/
Early on, I made all the http2 changes in the [h2_bundle file](h2_bundle.go). Due to its large size (10k+ lines), it becomes hard to maintain and add features. All changes made in h2_bundle should be migrated to the [http2 folder](http2) and all references to http2 in the library should import the http2 library. This should be done immediately, before implementing any of the other features.

## Add Enable Push as a field of http transport
The transports created by the HTTP1 client are empty http2 transports. There should be a way to add a default pushHandler, like the one in the [example](example_client_test.go) as a default pusher. This way, users can use Proxy as well, without a hacky http2 implementation.

## gzip, deflate, br
The `gzip, deflate, br` encoding should be implemented as an opt-in way for the client to use this instead of the standard gzip encoding. A helpful example can be found [here](https://play.golang.org/p/80HukFxfs4).

## Writing better tests
* Test fingerprinting bypass with this [site](https://privacycheck.sec.lrz.de/passive/fp_h2/fp_http2.html#fpDemoHttp2)
* Test for ENABLE_PUSH implementation, from [here](https://go-review.googlesource.com/c/net/+/181497/)
* Test for all features mentioned in [README](README.md), such as header order and pheader order with httptrace

## Merging upstream
When changes are made by the golang team on the [http]() or [http2](https://pkg.go.dev/golang.org/x/net/http2) library as a release branch,  
```
git remote add -f golang git@github.com:golang/go.git
git checkout -b golang-upstream golang/master
git subtree split -P src/crypto/tls/ -b golang-tls-upstream
git checkout master
git merge --no-commit golang-<http or http2>-upstream
```