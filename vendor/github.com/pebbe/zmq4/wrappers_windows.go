// +build windows

package zmq4

/*

#include <errno.h>
#include <zmq.h>

#if ZMQ_VERSION_MINOR < 2
// Version < 4.2.x
#include <zmq_utils.h>
int zmq_curve_public (char *z85_public_key, const char *z85_secret_key);
#endif // Version < 4.2.x

#if ZMQ_VERSION_MINOR < 1
const char *zmq_msg_gets (zmq_msg_t *msg, const char *property);
#if ZMQ_VERSION_PATCH < 5
// Version < 4.0.5
int zmq_proxy_steerable (const void *frontend, const void *backend, const void *capture, const void *control);
#endif // Version < 4.0.5
#endif // Version == 4.0.x

int zmq4_bind (void *socket, const char *endpoint)
{
    int i;
    i = zmq_bind(socket, endpoint);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

int zmq4_close (void *socket)
{
    int i;
    i = zmq_close(socket);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

int zmq4_connect (void *socket, const char *endpoint)
{
    int i;
    i = zmq_connect(socket, endpoint);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

int zmq4_ctx_get (void *context, int option_name)
{
    int i;
    i = zmq_ctx_get(context, option_name);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

void *zmq4_ctx_new ()
{
    void *v;
    v = zmq_ctx_new();
    if (v == NULL)
        errno = zmq_errno();
    return v;
}

int zmq4_ctx_set (void *context, int option_name, int option_value)
{
    int i;
    i = zmq_ctx_set(context, option_name, option_value);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

int zmq4_ctx_term (void *context)
{
    int i;
    i = zmq_ctx_term(context);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

int zmq4_curve_keypair (char *z85_public_key, char *z85_secret_key)
{
    int i;
    i = zmq_curve_keypair(z85_public_key, z85_secret_key);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

int zmq4_curve_public (char *z85_public_key, char *z85_secret_key)
{
    int i;
    i = zmq_curve_public(z85_public_key, z85_secret_key);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

int zmq4_disconnect (void *socket, const char *endpoint)
{
    int i;
    i = zmq_disconnect(socket, endpoint);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

int zmq4_getsockopt (void *socket, int option_name, void *option_value, size_t *option_len)
{
    int i;
    i = zmq_getsockopt(socket, option_name, option_value, option_len);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

const char *zmq4_msg_gets (zmq_msg_t *message, const char *property)
{
    const char *s;
    s = zmq_msg_gets(message, property);
    if (s == NULL)
        errno = zmq_errno();
    return s;
}

int zmq4_msg_recv (zmq_msg_t *msg, void *socket, int flags)
{
    int i;
    i = zmq_msg_recv(msg, socket, flags);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

int zmq4_poll (zmq_pollitem_t *items, int nitems, long timeout)
{
    int i;
    i = zmq_poll(items, nitems, timeout);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

int zmq4_proxy (void *frontend, void *backend, void *capture)
{
    int i;
    i = zmq_proxy(frontend, backend, capture);
    errno = zmq_errno();
    return i;
}

int zmq4_proxy_steerable (void *frontend, void *backend, void *capture, void *control)
{
    int i;
    i = zmq_proxy_steerable(frontend, backend, capture, control);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

int zmq4_send (void *socket, void *buf, size_t len, int flags)
{
    int i;
    i = zmq_send(socket, buf, len, flags);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

int zmq4_setsockopt (void *socket, int option_name, const void *option_value, size_t option_len)
{
    int i;
    i = zmq_setsockopt(socket, option_name, option_value, option_len);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

void *zmq4_socket (void *context, int type)
{
    void *v;
    v = zmq_socket(context, type);
    if (v == NULL)
        errno = zmq_errno();
    return v;
}

int zmq4_socket_monitor (void *socket, char *endpoint, int events)
{
    int i;
    i = zmq_socket_monitor(socket, endpoint, events);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

int zmq4_unbind (void *socket, const char *endpoint)
{
    int i;
    i = zmq_unbind(socket, endpoint);
    if (i < 0)
        errno = zmq_errno();
    return i;
}

*/
import "C"
