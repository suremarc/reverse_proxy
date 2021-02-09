package main

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"

	"strings"

	"github.com/gin-gonic/gin"
	apis "github.com/polygon-io/go-lib-apis"
	log "github.com/sirupsen/logrus"
)

type closeNotifyRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func (c *closeNotifyRecorder) CloseNotify() <-chan bool {
	return c.closed
}

func testingContext() (*gin.Context, *closeNotifyRecorder) {
	r := httptest.NewRecorder()
	r1 := &closeNotifyRecorder{
		ResponseRecorder: r,
		closed:           make(chan bool),
	}
	c, _ := gin.CreateTestContext(r1)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	return c, r1
}

// ReverseProxy is the last handler and it directs authenticated requests to the correct service
func ReverseProxy(backendURL url.URL, prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := backendURL.Host
		// if backendURL.Port() != "" {
		// 	host += ":" + backendURL.Port()
		// }

		c.Set(apis.ContextBackend, host)

		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = backendURL.Scheme
				req.URL.Host = host
			},
		}

		logger := apis.RequestLogger(c)
		//log.SetLevel(log.DebugLevel)
		logger.Debug(backendURL.Host)
		logger.Debug(host)
		proxy.ErrorHandler = func(rw http.ResponseWriter, request *http.Request, err error) {
			logger.WithError(err).Error("http: proxy error")
			rw.WriteHeader(http.StatusBadGateway)
		}

		// proxy.ModifyResponse = func(resp *http.Response) error {
		// 	// We do this because the reverse proxy already copies the relevant headers over from the response
		// 	// https://github.com/golang/go/blob/release-branch.go1.14/src/net/http/httputil/reverseproxy.go#L282
		// 	// Not doing this results in duplicated request ID's in the response header
		// 	resp.Header.Del(apis.HeaderRequestID)
		// 	return nil
		// }

		// c.Request.Header.Set(apis.HeaderAccountID, c.GetString(apis.ContextCallerAccountID))
		// p, err := authz.GetPermissions(c)
		// if err != nil && err != authz.ErrNoPermissions {
		// 	_ = c.Error(apis.DecorateError(err, "get_permissions"))
		// }

		// if err == nil {
		// 	if err := authz.SetPermissions(c, p); err != nil {
		// 		_ = c.Error(apis.DecorateError(err, "decorate_permissions"))
		// 	}
		// }

		// /accountservices/v1/accounts => /v1/accounts
		// /accountservices/internal/v1/accounts => /v1/accounts
		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, prefix)

		apis.RequestLogger(c).WithFields(log.Fields{
			"target_host": host,
			"path":        c.Request.URL.Path,
		}).Trace("proxying request")

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
