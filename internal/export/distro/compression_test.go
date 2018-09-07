//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"io/ioutil"
	"testing"
)

const (
	clearString = "This is the test string used for testing"
	gzipString  = "H4sIAAAJbogA/wrJyCxWyCxWKMlIVShJLS5RKC4pysxLVygtTk1RSMsvAgtm5qUDAgAA//8tdaMdKAAAAA=="
	zlibString  = "eJwKycgsVsgsVijJSFUoSS0uUSguKcrMS1coLU5NUUjLLwILZualAwIAAP//KucO4w=="
)

func TestGzip(t *testing.T) {

	comp := gzipTransformer{}
	enc := comp.Transform([]byte(clearString))

	compressed, err := base64.StdEncoding.DecodeString(string(enc))
	if err != nil {
		t.Fatal("Error base64 ", err)
	}

	var buf bytes.Buffer
	buf.Write(compressed)

	zr, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatal("Error decoding buffer ", err)
	}
	decoded, err := ioutil.ReadAll(zr)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if string(decoded) != clearString {
		t.Fatal("Decoded string ", string(enc), " is not ", clearString)
	}

	enc2 := comp.Transform([]byte(clearString))

	if !bytes.Equal(enc, enc2) {
		t.Fatal("Output should be the same", string(enc), " is not ", string(enc2))
	}
}

func TestZlib(t *testing.T) {

	comp := zlibTransformer{}
	enc := comp.Transform([]byte(clearString))

	compressed, err := base64.StdEncoding.DecodeString(string(enc))
	if err != nil {
		t.Fatal("Error base64 ", err)
	}

	var buf bytes.Buffer
	buf.Write(compressed)

	zr, err := zlib.NewReader(&buf)
	if err != nil {
		t.Fatal("Error decoding buffer ", err)
	}
	decoded, err := ioutil.ReadAll(zr)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if string(decoded) != clearString {
		t.Fatal("Decoded string ", string(enc), " is not ", clearString)
	}

	enc2 := comp.Transform([]byte(clearString))

	if !bytes.Equal(enc, enc2) {
		t.Fatal("Output should be the same", string(enc), " is not ", string(enc2))
	}
}

var result []byte

func BenchmarkGzip(b *testing.B) {

	comp := gzipTransformer{}

	var enc []byte
	for i := 0; i < b.N; i++ {
		enc = comp.Transform([]byte(clearString))
	}
	b.SetBytes(int64(len(enc)))
	result = enc
}

func BenchmarkZlib(b *testing.B) {

	comp := zlibTransformer{}

	var enc []byte
	for i := 0; i < b.N; i++ {
		enc = comp.Transform([]byte(clearString))
	}
	b.SetBytes(int64(len(enc)))
	result = enc
}
