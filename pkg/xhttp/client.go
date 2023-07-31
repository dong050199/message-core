package xhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/sirupsen/logrus"
	"go.elastic.co/apm/module/apmhttp"
	"golang.org/x/net/context/ctxhttp"
)

const (
	defaultTimeout       = 30 * time.Second
	defaultLogBodyLength = 3000
	defaultNamespace     = "yams"
	defaultSubsystem     = "yams"
)

// nolint: lll
// Không cần check long line linter cho interface
type Client interface {
	PostJSONWithCustomHeader(ctx context.Context, url string, data, target interface{}, customHeader http.Header, method string, reqOpts ...RequestOption) (statusCode int, err error)
	PostJSON(c context.Context, url string, data, target interface{}, reqOptions ...RequestOption) (int, error)
	PostForm(c context.Context, url string, data, target interface{}, reqOptions ...RequestOption) (int, error)
	Get(c context.Context, url string, target interface{}, reqOptions ...RequestOption) (int, error)
	GetWithQuery(c context.Context, url string, data, target interface{}, reqOptions ...RequestOption) (int, error)
	GetWithQueryCustomHeader(c context.Context, url string, data, target interface{}, customHeader http.Header, reqOptions ...RequestOption) (int, error)
	GetWithoutEncodedQuery(c context.Context,
		url string, data, target interface{}, reqOptions ...RequestOption) (int, error)
	Do(ctx context.Context, request *http.Request, target interface{}, decodeNumber ...bool) (int, error)
	SendHTTPRequest(ctx context.Context, method string, path string, payload interface{}, outPut interface{}, reqOptions ...RequestOption) (int, error)
}

type client struct {
	client *http.Client
	opts   clientOptions
}

func NewClient(opts ...Option) Client {
	optsArg := getOptionsArg(opts)
	transport := NewTransport(optsArg)
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   optsArg.timeout,
	}
	c := &client{
		client: httpClient,
		opts:   optsArg,
	}
	return c
}

func getOptionsArg(opts []Option) clientOptions {
	// Init default options arg
	optsArgs := clientOptions{
		skipLog:         false,
		splitLogBody:    false,
		splitLogBodyLen: defaultLogBodyLength,
		timeout:         defaultTimeout,
	}

	for _, opt := range opts {
		opt.apply(&optsArgs)
	}
	return optsArgs
}

func (c *client) PostJSON(ctx context.Context,
	url string, data, target interface{}, reqOpts ...RequestOption) (statusCode int, err error) {
	header := c.getRequestHeader(reqOpts...)
	req, err := NewRequestBuilderWithCtx(ctx).
		WithMethod(http.MethodPost).
		WithURL(url).
		WithHeaders(header).
		WithBody(MIMEJSON, data).
		Build()
	if err != nil {
		return
	}
	return c.Do(ctx, req, target)
}

func (c *client) PostJSONWithCustomHeader(ctx context.Context,
	url string,
	data, target interface{},
	customHeader http.Header,
	method string,
	reqOpts ...RequestOption,
) (statusCode int, err error) {
	header := c.getRequestHeader(reqOpts...)
	req, err := NewRequestBuilderWithCtx(ctx).
		WithMethod(method).
		WithURL(url).
		WithHeaders(header).
		WithBody(MIMEJSON, data).
		Build()
	if err != nil {
		return
	}
	req.Header = customHeader
	decodeNumber := false
	if len(reqOpts) > 0 {
		decodeNumber = reqOpts[0].DecodeNumber
	}
	return c.Do(ctx, req, target, decodeNumber)
}

func (c *client) PostForm(ctx context.Context,
	url string, data, target interface{}, reqOpts ...RequestOption) (statusCode int, err error) {
	header := c.getRequestHeader(reqOpts...)
	req, err := NewRequestBuilderWithCtx(ctx).
		WithMethod(http.MethodPost).
		WithURL(url).
		WithHeaders(header).
		WithBody(MIMEPOSTForm, data).
		Build()
	if err != nil {
		return
	}
	return c.Do(ctx, req, target)
}

func (c *client) Get(ctx context.Context,
	url string, target interface{}, reqOpts ...RequestOption) (statusCode int, err error) {
	header := c.getRequestHeader(reqOpts...)
	req, err := NewRequestBuilderWithCtx(ctx).
		WithMethod(http.MethodGet).
		WithURL(url).
		WithHeaders(header).
		Build()
	if err != nil {
		return
	}
	return c.Do(ctx, req, target)
}

func (c *client) GetWithQuery(ctx context.Context,
	reqURL string, data, target interface{}, reqOpts ...RequestOption) (statusCode int, err error) {
	header := c.getRequestHeader(reqOpts...)
	req, err := NewRequestBuilderWithCtx(ctx).
		WithMethod(http.MethodGet).
		WithURL(reqURL).
		WithHeaders(header).
		Build()
	if err != nil {
		return
	}

	if data != nil {
		v, err := query.Values(data)
		if err != nil {
			return 0, err
		}
		nonEncodedValue, _ := url.PathUnescape(v.Encode())
		req.URL.RawQuery = nonEncodedValue
	}
	return c.Do(ctx, req, target)
}

func (c *client) GetWithoutEncodedQuery(ctx context.Context,
	reqURL string, data, target interface{}, reqOpts ...RequestOption) (statusCode int, err error) {
	header := c.getRequestHeader(reqOpts...)
	req, err := NewRequestBuilderWithCtx(ctx).
		WithMethod(http.MethodGet).
		WithURL(reqURL).
		WithHeaders(header).
		Build()
	if err != nil {
		return
	}

	if data != nil {
		v, err := query.Values(data)
		if err != nil {
			return 0, err
		}
		nonEncodedValue, _ := url.QueryUnescape(v.Encode())
		req.URL.RawQuery = nonEncodedValue
	}
	return c.Do(ctx, req, target)
}

func (c *client) Do(ctx context.Context, request *http.Request, target interface{}, decodeNumber ...bool) (int, error) {
	if requestID := request.Header.Get(RequestIDHeader); requestID == "" {
		request.Header.Set(RequestIDHeader, getContextIDFromCtx(ctx))
	}
	rsp, err := c.client.Do(request)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = rsp.Body.Close()
	}()

	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return 0, err
	}

	if len(bodyBytes) == 0 {
		return rsp.StatusCode, nil
	}

	if len(decodeNumber) > 0 && decodeNumber[0] == true {
		d := json.NewDecoder(bytes.NewBuffer(bodyBytes))
		d.UseNumber()
		return rsp.StatusCode, d.Decode(target)
	}

	return rsp.StatusCode, json.Unmarshal(bodyBytes, target)
}

func (c *client) getRequestHeader(reqOpts ...RequestOption) map[string]string {
	if len(reqOpts) == 0 {
		return nil
	}
	reqOpt := reqOpts[0]
	header := reqOpt.Header
	if header == nil {
		header = make(map[string]string)
	}
	if reqOpt.GroupPath != "" {
		header[GroupPathHeader] = reqOpt.GroupPath
	}
	return header
}

func getContextIDFromCtx(ctx context.Context) string {
	if result, ok := ctx.Value("context_id").(string); ok {
		return result
	}
	return ""
}

func (c *client) GetWithQueryCustomHeader(ctx context.Context,
	urlReq string, data, target interface{},
	customHeader http.Header, reqOpts ...RequestOption) (statusCode int, err error) {
	header := c.getRequestHeader(reqOpts...)
	req, err := NewRequestBuilderWithCtx(ctx).
		WithMethod(http.MethodGet).
		WithURL(urlReq).
		WithHeaders(header).
		Build()
	if err != nil {
		return
	}

	if data != nil {
		v, err := query.Values(data)
		if err != nil {
			return 0, err
		}
		nonEncodedValue, _ := url.PathUnescape(v.Encode())
		req.URL.RawQuery = nonEncodedValue
		// req.URL.RawQuery = v.Encode()
	}
	fmt.Println("111111111111111111111111111122221", data)
	fmt.Println("11111111111111111111111111111", req.URL)
	req.Header = customHeader
	return c.Do(ctx, req, target)
}

func (c *client) SendHTTPRequest(
	ctx context.Context,
	method string,
	path string,
	payload interface{},
	outPut interface{},
	reqOptions ...RequestOption,
) (status int, err error) {
	req, err := c.newRequest(ctx, method, path, payload, reqOptions)
	if err != nil {
		return -1, fmt.Errorf("failed to create %s request: %w", method, err)
	}

	status, err = c.doRequest(ctx, req, outPut)
	if err != nil {
		return
	}

	return
}

/*Internal implementation*/
func (c *client) newRequest(
	ctx context.Context,
	method,
	path string,
	payload interface{},
	reqOpts []RequestOption,
) (req *http.Request, err error) {
	header := c.getRequestHeader(reqOpts...)
	req, err = NewRequestBuilderWithCtx(ctx).
		WithMethod(method).
		WithBody(header[contentTypeField], payload).
		// WithURL(fmt.Sprintf("%s/%s", strings.TrimRight(c.Ops.URL, "/"), path)).
		WithURL(path).
		WithHeaders(header).
		Build()
	if err != nil {
		return
	}
	if reqOpts[0].HeaderCustom != nil {
		req.Header = reqOpts[0].HeaderCustom
	}

	return req, nil
}

func (h *client) doRequest(
	ctx context.Context,
	r *http.Request,
	outPut interface{},
) (status int, err error) {
	if requestID := r.Header.Get(RequestIDHeader); requestID == "" {
		r.Header.Set(RequestIDHeader, getContextIDFromCtx(ctx))
	}

	apmClient := apmhttp.WrapClient(h.client)
	resp, err := ctxhttp.Do(ctx, apmClient, r)
	if err != nil {
		logrus.WithField("MAKE-REQUEST-ERROR", err).WithField("Status-Code", resp.StatusCode).
			WithFields(
				logrus.Fields{
					"URL":    r.URL.String(),
					"Method": r.Method,
				}).
			WithError(err).
			Error()
		return -1, fmt.Errorf("failed to make request: %w", err)
	}

	if resp == nil {
		return
	}

	// return first if not need output return
	if outPut == nil {
		return
	}

	var buf bytes.Buffer
	dec := json.NewDecoder(io.TeeReader(resp.Body, &buf))
	if err := dec.Decode(outPut); err != nil {

		logrus.WithField("PARSE_RESPONSE_BODY_ERROR", err).
			WithFields(
				logrus.Fields{
					"URL":    r.URL.String(),
					"Method": r.Method,
					"Output": buf.String(),
				}).
			WithField("Status-Code", resp.StatusCode).
			WithError(err).
			Error()
		return -1, fmt.Errorf("could not parse response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logrus.WithField("DO_HTTP_REQUEST_ERROR", err).
			WithFields(
				logrus.Fields{"Status": resp.Status,
					"PostForm": r.PostForm,
					"Form":     r.Form,
					"Header":   r.Header,
					"URL":      r.URL.String(),
					"Method":   r.Method,
					"Output":   outPut,
				}).WithField("Status-Code", resp.StatusCode).
			WithError(err).
			Error()

		switch resp.StatusCode {
		case http.StatusInternalServerError:
			return
		default:
			return status, errors.New(resp.Status)
		}
	}

	defer resp.Body.Close() // nolint: errcheck

	return http.StatusOK, nil
}
