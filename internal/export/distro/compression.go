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
)

type gzipTransformer struct {
	writer *gzip.Writer
}

func (gzt *gzipTransformer) Transform(data []byte) []byte {
	var buf bytes.Buffer

	if gzt.writer == nil {
		gzt.writer = gzip.NewWriter(&buf)
	} else {
		gzt.writer.Reset(&buf)
	}

	gzt.writer.Write(data)
	gzt.writer.Close()

	return bytesBufferToBase64(buf)
}

type zlibTransformer struct {
	writer *zlib.Writer
}

func (zlt *zlibTransformer) Transform(data []byte) []byte {
	var buf bytes.Buffer

	if zlt.writer == nil {
		zlt.writer = zlib.NewWriter(&buf)
	} else {
		zlt.writer.Reset(&buf)
	}

	zlt.writer.Write(data)
	zlt.writer.Close()

	return bytesBufferToBase64(buf)
}

func bytesBufferToBase64(buf bytes.Buffer) []byte {
	dst := make([]byte, base64.StdEncoding.EncodedLen(buf.Len()))
	base64.StdEncoding.Encode(dst, buf.Bytes())
	return dst
}
