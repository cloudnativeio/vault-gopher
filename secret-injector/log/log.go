// NOTE: something in my mind, log http from the module itself.
// take the request and response and dump it and get the details
package log

import (
"fmt"
"net/http"
"os"
"regexp"
"strconv"
"strings"
"time"

"github.com/sirupsen/logrus"
)

type GopherLogger struct {
	*logrus.Logger
}

type Http struct {
	Request  *http.Request
	Response *http.Response
}

type Fields struct {
	method   string
	protocol string
	code     int
	path     string
	host     string
}

// Instantiate a new logger
func NewLogger() *GopherLogger {
	var baseLogger = logrus.New()

	var gopherLogger = &GopherLogger{baseLogger}
	gopherLogger.Level = logrus.DebugLevel
	gopherLogger.Formatter = &logrus.TextFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339,
	}
	gopherLogger.Out = os.Stdout

	return gopherLogger
}

// Method that implements fields
func (h *Http) getFields() []string {
	var f Fields
	fields := make([]string, 0)

	if h.Request.Method != "" {
		f.method = h.Request.Method
		fields = append(fields, f.method)
	}

	if h.Response.StatusCode != 0 {
		f.code = h.Response.StatusCode
		fields = append(fields, strconv.Itoa(f.code))
	}

	if h.Request.Proto != "" {
		f.protocol = h.Request.Proto
		fields = append(fields, f.protocol)
	}

	if h.Request.URL.Host != "" {
		f.path = h.Request.URL.Path
		fields = append(fields, f.path)
	}

	if h.Request.Host != "" {
		f.host = h.Request.Host
		fields = append(fields, f.host)
	}

	return fields
}

// Method that implements logging of http request and response
func (g *GopherLogger) LogGopher(resp *http.Response, req *http.Request) {
	var f Fields
	data := f.parameters(resp, req)

	if ok, _ := regexp.MatchString("^5", strconv.Itoa(resp.StatusCode)); ok {
		g.log5xx(data)
	}
	if ok, _ := regexp.MatchString("^4", strconv.Itoa(resp.StatusCode)); ok {
		g.log4xx(data)
	}
	if ok, _ := regexp.MatchString("^3", strconv.Itoa(resp.StatusCode)); ok {
		g.log3xx(data)
	}
	if ok, _ := regexp.MatchString("^2", strconv.Itoa(resp.StatusCode)); ok {
		g.log2xx(data)
	}
}

// Log 5xx status code
func (g *GopherLogger) log5xx(s string) {
	g.Fatal(s)
}

// Log 4xx status code
func (g *GopherLogger) log4xx(s string) {
	g.Warn(s)
}

// Log 2xx status code
func (g *GopherLogger) log3xx(s string) {
	g.Info(s)
}

// Log 2xx status code
func (g *GopherLogger) log2xx(s string) {
	g.Info(s)
}

// Method that implements field parameters
func (f *Fields) parameters(resp *http.Response, req *http.Request) string {
	p := &Http{req, resp}
	d := p.getFields()
	return strings.Trim(fmt.Sprintf("%v", d), "[ ]")
}
