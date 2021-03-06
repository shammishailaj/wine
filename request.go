package wine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gopub/types"
)

const (
	ContentType = "Content-Type"
)

const (
	MIMETEXT              = "text/plain"
	MIMEJSON              = "application/json"
	MIMEHTML              = "text/html"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
	MIMEOctetStream       = "application/octet-stream"
)

// Request is a wrapper of http.Request, aims to provide more convenient interface
type Request struct {
	HTTPRequest *http.Request
	Parameters  types.M
}

type RequestParser interface {
	ParseHTTPRequest(req *http.Request, maxMemory int64) (parameters types.M, err error)
}

type DefaultRequestParser struct {
	headerFields map[string]bool
}

func NewDefaultRequestParser() *DefaultRequestParser {
	return &DefaultRequestParser{
		headerFields: map[string]bool{
			"sid":       true,
			"device_id": true,
		},
	}
}

func (p *DefaultRequestParser) ParseHTTPRequest(req *http.Request, maxMemory int64) (types.M, error) {
	startAt := time.Now()
	params := types.M{}
	for _, cookie := range req.Cookies() {
		params[cookie.Name] = cookie.Value
	}

	logger.Debug("Cost:", time.Since(startAt))
	for k, v := range req.Header {
		if strings.HasPrefix(k, "x-") || strings.HasPrefix(k, "X-") || p.headerFields[k] {
			params[strings.ToLower(k[2:])] = v
		}
	}

	logger.Debug("Cost:", time.Since(startAt))
	params.AddMap(convertToM(req.URL.Query()))

	contentType := req.Header.Get(ContentType)
	for i, ch := range contentType {
		if ch == ' ' || ch == ';' {
			contentType = contentType[:i]
			break
		}
	}

	logger.Debug("Cost:", time.Since(startAt))
	switch contentType {
	case MIMEHTML, MIMETEXT:
		break
	case MIMEJSON:
		d, e := ioutil.ReadAll(req.Body)
		if e != nil {
			logger.Error(e)
			break
		}

		logger.Debug("Cost:", time.Since(startAt))
		if len(d) > 0 {
			var m types.M
			e = jsonUnmarshal(d, &m)
			if e != nil {
				break
			}
			params.AddMap(m)
		}
		logger.Debug("Cost:", time.Since(startAt))
	case MIMEPOSTForm:
		logger.Debug("Cost:", time.Since(startAt))
		err := req.ParseForm()
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		logger.Debug("Cost:", time.Since(startAt))
		params.AddMap(convertToM(req.Form))
		logger.Debug("Cost:", time.Since(startAt))
	case MIMEMultipartPOSTForm:
		logger.Debug("Cost:", time.Since(startAt))
		err := req.ParseMultipartForm(maxMemory)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		logger.Debug("Cost:", time.Since(startAt))
		if req.MultipartForm != nil && req.MultipartForm.File != nil {
			params.AddMap(convertToM(req.MultipartForm.Value))
		}
		logger.Debug("Cost:", time.Since(startAt))
	default:
		if len(contentType) > 0 {
			err := errors.New(fmt.Sprintf("unsupported content type: %s", contentType))
			logger.Error(err)
			return nil, err
		}
	}

	logger.Debug("Cost:", time.Since(startAt))
	return params, nil
}

func convertToM(values map[string][]string) types.M {
	m := types.M{}
	for k, v := range values {
		i := strings.Index(k, "[]")
		if i >= 0 && i == len(k)-2 {
			k = k[0 : len(k)-2]
		}
		k = strings.ToLower(k)
		if len(v) > 1 || i >= 0 {
			m[k] = v
		} else if len(v) == 1 {
			m[k] = v[0]
		}
	}

	return m
}

func jsonUnmarshal(data []byte, pJSONObj interface{}) error {
	if len(data) == 0 {
		return errors.New("data is empty")
	}
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.UseNumber()
	err := decoder.Decode(pJSONObj)
	if err != nil {
		logger.Error(err)
	}
	return err
}
