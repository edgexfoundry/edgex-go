// +build windows

package zmq4

/*
#include <zmq.h>
#include <winsock2.h>
#include "zmq4.h"
*/
import "C"

// winsock2.h needed for ZeroMQ version 4.3.3

import (
	"unsafe"
)

/*
ZMQ_FD: Retrieve file descriptor associated with the socket

See: http://api.zeromq.org/4-1:zmq-getsockopt#toc9
*/
func (soc *Socket) GetFd() (uintptr, error) {
	value := C.SOCKET(0)
	size := C.size_t(unsafe.Sizeof(value))
	var i C.int
	var err error
	for {
		i, err = C.zmq4_getsockopt(soc.soc, C.ZMQ_FD, unsafe.Pointer(&value), &size)
		// not really necessary because Windows doesn't have EINTR
		if i == 0 || !soc.ctx.retry(err) {
			break
		}
	}
	if i != 0 {
		return uintptr(0), errget(err)
	}
	return uintptr(value), nil
}
