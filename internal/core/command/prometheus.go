package command

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	dflBuckets = []float64{0.3, 1.0, 2.5, 5.0}
)

const (
	requestName = "command_requests_total"
	latencyName = "command_request_duration_seconds"
	responseName = "command_response_data_bytes"
)

// Opts specifies options how to create new PrometheusMiddleware.
type Opts struct {
	// Buckets specifies an custom buckets to be used in request histograpm.
	Buckets []float64
}

// PrometheusMiddleware specifies the metrics that is going to be generated
type PrometheusMiddleware struct {
	request *prometheus.CounterVec
	latency *prometheus.HistogramVec
	response  *prometheus.GaugeVec
}

// NewPrometheusMiddleware creates a new PrometheusMiddleware instance
func NewPrometheusMiddleware(opts Opts) *PrometheusMiddleware {
	var prometheusMiddleware PrometheusMiddleware

	prometheusMiddleware.request = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: requestName,
			Help: "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
		},
		[]string{"code", "service","device","command","method", "path"},
	)

	if err := prometheus.Register(prometheusMiddleware.request); err != nil {
		log.Println("prometheusMiddleware.request was not registered:", err)
	}

	buckets := opts.Buckets
	if len(buckets) == 0 {
		buckets = dflBuckets
	}

	prometheusMiddleware.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    latencyName,
		Help:    "How long it took to process the request, partitioned by status code, method and HTTP path.",
		Buckets: buckets,
	},
		[]string{"code", "service","device","command","method", "path"},
	)

	if err := prometheus.Register(prometheusMiddleware.latency); err != nil {
		log.Println("prometheusMiddleware.latency was not registered:", err)
	}

	prometheusMiddleware.response = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: responseName,
			Help: "Response data bytes , partitioned by status code, method and HTTP path.",
		},
		[]string{"code", "service","device","command","method", "path"},
	)

	if err := prometheus.Register(prometheusMiddleware.response); err != nil {
		log.Println("prometheusMiddleware.response was not registered:", err)
	}

	return &prometheusMiddleware
}

// InstrumentHandlerDuration is a middleware that wraps the http.Handler and it record
// how long the handler took to run, which path was called, and the status code.
// This method is going to be used with gorilla/mux.
func (p *PrometheusMiddleware) PrometheusHandlerMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var deviceName,commandName string

		begin := time.Now()

		delegate := &responseWriterDelegator{ResponseWriter: w}
		next.ServeHTTP(delegate, r) // call original

		//route := mux.CurrentRoute(r)
		//path, _ := route.GetPathTemplate()
		path := r.URL.Path
		tokens := strings.Split(path,"/")
		serviceName := r.Host

		if len(tokens) == 7 {
			deviceName = tokens[5]
			commandName = tokens[6]
		} else{
			deviceName = ""
			commandName = ""
		}

		//log.Println("r:",r.Host)
		//log.Println("path:",path)
		//log.Println("tokens:", tokens)
		//log.Println("tokens:", len(tokens))

		code := sanitizeCode(delegate.status)
		method := sanitizeMethod(r.Method)
		length := delegate.written;

		go p.request.WithLabelValues(
			code,
			serviceName,
			deviceName,
			commandName,
			method,
			path,
		).Inc()

		go p.latency.WithLabelValues(
			code,
			serviceName,
			deviceName,
			commandName,
			method,
			path,
		).Observe(float64(time.Since(begin)) / float64(time.Second))

		go p.response.WithLabelValues(
			code,
			serviceName,
			deviceName,
			commandName,
			method,
			path,
		).Set(float64(length))
	})
}

type responseWriterDelegator struct {
	http.ResponseWriter
	status      int
	written     int64
	wroteHeader bool
}

func (r *responseWriterDelegator) WriteHeader(code int) {
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseWriterDelegator) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

func sanitizeMethod(m string) string {
	return strings.ToLower(m)
}

func sanitizeCode(s int) string {
	return strconv.Itoa(s)
}
