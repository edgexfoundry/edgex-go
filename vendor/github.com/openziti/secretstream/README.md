# secretstream
Implementation of [libsodium](https://github.com/jedisct1/libsodium)'s [secretstream](https://libsodium.gitbook.io/doc/secret-key_cryptography/secretstream) in Go

The main goal of this project is allow using `secretstream` between programs using libsodium and
programs written in Go without resorting to wrapping libsodium in Go. golang.org/x/crypto has all necessary
algorithms to make that happen.

## Testing against libsodium
It is important that this implementation is compatible with libsodium. Tests tagged with `compat_test` use libsodium to test compatibility.

make sure you have libsodium installed and ready to be used
```bash
$ sudo apt install libsodium libsodium-dev
```
_other platforms something similar_

You're ready to run tests!
```bash
$ go test --tags=compat_test ./...
```
