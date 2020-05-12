package httpclient

import (
	"context"
	"github.com/b2wdigital/restQL-golang/internal/domain"
	"github.com/b2wdigital/restQL-golang/internal/platform/conf"
	"github.com/b2wdigital/restQL-golang/internal/platform/logger"
	"github.com/b2wdigital/restQL-golang/internal/platform/plugins"
	"github.com/pkg/errors"
	"github.com/rs/dnscache"
	"github.com/valyala/fasthttp"
	"net"
	"sync"
	"time"
)

type httpResult struct {
	target   string
	duration time.Duration
	err      error
	response *fasthttp.Response
}

type HttpClient struct {
	client        *fasthttp.Client
	log           *logger.Logger
	pluginManager plugins.Manager
	responsePool  *sync.Pool
}

func New(log *logger.Logger, pm plugins.Manager, cfg *conf.Config) *HttpClient {
	clientCfg := cfg.Web.Client

	r := &dnscache.Resolver{}
	go func() {
		t := time.NewTicker(10 * time.Minute)
		defer t.Stop()
		for range t.C {
			r.Refresh(true)
		}
	}()
	dialer := &fasthttp.TCPDialer{
		Resolver: &net.Resolver{
			PreferGo:     true,
			StrictErrors: false,
			Dial: func(ctx context.Context, network, address string) (conn net.Conn, err error) {
				host, port, err := net.SplitHostPort(address)
				if err != nil {
					return nil, err
				}
				ips, err := r.LookupHost(context.Background(), host)
				if err != nil {
					return nil, err
				}
				for _, ip := range ips {
					var dialer net.Dialer
					conn, err = dialer.Dial(network, net.JoinHostPort(ip, port))
					if err == nil {
						break
					}
				}
				return
			},
		},
	}
	c := &fasthttp.Client{
		Name:                          "restql",
		NoDefaultUserAgentHeader:      false,
		DisableHeaderNamesNormalizing: true,
		Dial:                          dialer.Dial,
		ReadTimeout:                   clientCfg.ReadTimeout,
		WriteTimeout:                  clientCfg.WriteTimeout,
		MaxConnsPerHost:               clientCfg.MaxConnsPerHost,
		MaxIdleConnDuration:           clientCfg.MaxIdleConnDuration,
		MaxConnDuration:               clientCfg.MaxConnDuration,
		MaxConnWaitTimeout:            clientCfg.MaxConnWaitTimeout,
	}

	rp := &sync.Pool{
		New: func() interface{} {
			return make(chan httpResult)
		},
	}

	return &HttpClient{client: c, log: log, pluginManager: pm, responsePool: rp}
}

func (hc *HttpClient) Do(ctx context.Context, request domain.HttpRequest) (domain.HttpResponse, error) {
	requestCtx := hc.pluginManager.RunBeforeRequest(ctx, request)

	c := hc.responsePool.Get().(chan httpResult)

	go func() {
		req := fasthttp.AcquireRequest()

		err := setupRequest(request, req)
		if err != nil {
			hc.log.Error("failed to setup http client request", err)
			fasthttp.ReleaseRequest(req)
			c <- httpResult{target: request.Host, err: err, duration: 0}
			return
		}

		res := fasthttp.AcquireResponse()
		start := time.Now()
		err = hc.client.DoTimeout(req, res, request.Timeout)
		finish := time.Since(start)

		reqUri := req.URI().String()
		fasthttp.ReleaseRequest(req)

		c <- httpResult{target: reqUri, err: err, duration: finish, response: res}
	}()

	hr := <-c
	hc.responsePool.Put(c)

	switch {
	case hr.err == fasthttp.ErrTimeout:
		hc.log.Info("request timed out", "url", hr.target, "method", request.Method, "duration-ms", hr.duration.Milliseconds())
		response := makeErrorResponse(hr.target, hr.duration, hr.err)

		fasthttp.ReleaseResponse(hr.response)

		hc.pluginManager.RunAfterRequest(requestCtx, request, response, hr.err)

		return response, domain.ErrRequestTimeout
	case hr.err != nil:
		response := makeErrorResponse(hr.target, hr.duration, hr.err)

		if hr.response != nil {
			fasthttp.ReleaseResponse(hr.response)
		}

		hc.pluginManager.RunAfterRequest(requestCtx, request, response, hr.err)

		return response, errors.Wrap(hr.err, "request execution failed")
	}

	response, err := makeResponse(hr.target, hr.response, hr.duration)
	if err != nil {
		response = makeErrorResponse(hr.target, hr.duration, err)

		fasthttp.ReleaseResponse(hr.response)

		hc.pluginManager.RunAfterRequest(requestCtx, request, response, err)

		return response, err
	}

	fasthttp.ReleaseResponse(hr.response)

	hc.pluginManager.RunAfterRequest(requestCtx, request, response, err)

	return response, nil
}
