// +build compat_test

package secretstream

// #cgo LDFLAGS: -lsodium
// #include <sodium.h>
// #include <sodium/crypto_secretstream_xchacha20poly1305.h>
import "C"
import (
	"errors"
)

type sodiumStream struct {
	state C.crypto_secretstream_xchacha20poly1305_state
}

func NewSodiumSendStream(key []byte) (Encryptor, []byte, error) {
	res := &sodiumStream{}

	ckey := C.CBytes(key)
	defer C.free(ckey)

	header := C.malloc(C.crypto_secretstream_xchacha20poly1305_HEADERBYTES)
	defer C.free(header)

	rc := C.crypto_secretstream_xchacha20poly1305_init_push(&res.state, (*C.uchar)(header), (*C.uchar)(ckey))
	if rc != 0 {
		return nil, nil, cryptoFailure
	}
	// fmt.Printf("sodium: %+v\n", res.state)

	return res, C.GoBytes(header, C.crypto_secretstream_xchacha20poly1305_HEADERBYTES), nil
}

func NewSodiumRecvStream(key []byte, header []byte) (Decryptor, error) {
	res := &sodiumStream{}

	chdr := C.CBytes(header)
	defer C.free(chdr)

	ckey := C.CBytes(key)
	defer C.free(ckey)

	rc := C.crypto_secretstream_xchacha20poly1305_init_pull(&res.state, (*C.uchar)(chdr), (*C.uchar)(ckey))
	if rc != 0 {
		return nil, cryptoFailure
	}
	// fmt.Printf("receiver = %+v", res.state)

	return res, nil
}

func (s *sodiumStream) Push(plaintext []byte, tag byte) ([]byte, error) {
	pt := C.CBytes(plaintext)
	defer C.free(pt)

	cipher_len := len(plaintext) + C.crypto_secretstream_xchacha20poly1305_ABYTES
	ct := C.malloc((C.size_t)(cipher_len))
	defer C.free(ct)

	cipher_len_ull := C.ulonglong(cipher_len)
	pt_len_ull := C.ulonglong(len(plaintext))
	if C.crypto_secretstream_xchacha20poly1305_push(&s.state,
		(*C.uchar)(ct), &cipher_len_ull,
		(*C.uchar)(pt), pt_len_ull,
		nil, 0,
		C.uchar(tag)) != 0 {
		return nil, errors.New("sodium error")
	}

	return C.GoBytes(ct, C.int(cipher_len)), nil
}

func (s *sodiumStream) Pull(ciphertext []byte) ([]byte, byte, error) {
	ctc := C.CBytes(ciphertext)
	defer C.free(ctc)

	mlen := C.ulong(len(ciphertext) - C.crypto_secretstream_xchacha20poly1305_ABYTES)
	mlen_ull := C.ulonglong(mlen)
	msg := C.malloc((C.size_t)(mlen))
	var tag C.uchar

	if C.crypto_secretstream_xchacha20poly1305_pull(&s.state,
		(*C.uchar)(msg), &mlen_ull, &tag,
		(*C.uchar)(ctc), C.ulonglong(len(ciphertext)),
		nil, 0) != 0 {
		return nil, 0, cryptoFailure
	}
	// fmt.Printf("receiver = %+v", s.state)

	return C.GoBytes(msg, C.int(mlen)), byte(tag), nil
}
