package gorequests

import (
	"code.byted.org/kite/kitutil"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// doRequest send request
func (r *Request) doRequest() error {
	return r.doRequestFactor(r.doInternalRequest)
}

// doRequest send request
func (r *Request) doInternalRequest() error {
	if r.isRequest {
		return nil
	}

	r.cachedurl = r.parseRequestURL()

	r.logger.Info(r.Context(), "[gorequests] %s: %s, body=%s, header=%+v", r.method, r.cachedurl, r.rawBody, r.header)

	if r.persistentJar != nil {
		defer func() {
			if err := r.persistentJar.Save(); err != nil {
				r.logger.Error(r.Context(), "save cookie failed: %s", err)
			}
		}()
	}

	req, err := http.NewRequest(r.method, r.cachedurl, r.body)
	if err != nil {
		return fmt.Errorf("[gorequest] %s %s new request failed: %w", r.method, r.cachedurl, err)
	}

	req.Header = r.header

	// TODO: reuse client
	c := &http.Client{
		Timeout: r.timeout,
	}
	if r.isIgnoreSSL {
		c.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	if r.persistentJar != nil {
		c.Jar = r.persistentJar
	}
	if r.isNoRedirect {
		c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	r.reqTime = time.Now()
	resp, err := c.Do(req)
	r.respTime = time.Now()
	r.doErr = err
	r.resp = resp
	r.isRequest = true
	if err != nil {
		return fmt.Errorf("[gorequest] %s %s send request failed: %w", r.method, r.cachedurl, err)
	}
	err = r.doProduceLog()
	if err != nil {
		r.logger.Error(r.Context(), "produce log failed: %s", err)
	}
	return nil
}

// doRead send request and read response
func (r *Request) doRead() error {
	return r.doRequestFactor(func() error {
		if err := r.doInternalRequest(); err != nil {
			return err
		}

		if r.isRead {
			return nil
		}

		var err error
		r.bytes, err = ioutil.ReadAll(r.resp.Body)
		r.isRead = true
		if err != nil {
			return fmt.Errorf("[gorequest] %s %s read response failed: %w", r.method, r.cachedurl, err)
		}

		r.logger.Info(r.Context(), "[gorequests] %s: %s, status_code: %d, header: %s, doRead: %s", r.method, r.cachedurl, r.resp.StatusCode, r.resp.Header, r.bytes)
		return nil
	})
}

func (r *Request) doProduceLog() error {
	if logProducer == nil {
		return nil
	}

	if err := r.doInternalRequest(); err != nil {
		return err
	}

	if r.isSend {
		return nil
	}

	message := LogMessage{
		Method:            r.method,
		Url:               r.cachedurl,
		RequestBody:       string(r.rawBody),
		RequestHeader:     r.header,
		RequestTime:       r.reqTime,
		ResponseBody:      string(r.bytes),
		ResponseHeader:    r.resp.Header,
		ResponseStateCode: r.resp.StatusCode,
		ResponseTime:      r.respTime,
		TimeConsuming:     r.respTime.UnixMilli() - r.reqTime.UnixMilli(),
		LogId:             r.getStrCtx(kitutil.LOGIDKEY),
		RequestType:       RequestMessageType_Out,
	}
	if r.doErr != nil {
		message.ErrorMessage = r.doErr.Error()
	}
	r.log = &message
	data, _ := json.Marshal(message)

	var err error
	err = logProducer.SendLogMessage(r.context, data)
	r.isSend = true
	if err != nil {
		return fmt.Errorf("[gorequest] %s %s send log message failed %w, message: %s", r.method, r.cachedurl, err, string(data))
	}

	r.logger.Info(r.Context(), "[gorequests] %s: %s, produce log: %s", r.method, r.cachedurl, string(data))
	return nil
}

func (r *Request) doRequestFactor(f func() error) error {
	if r.err != nil {
		return r.err
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	r.err = f()
	return r.err
}
