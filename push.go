package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	_ "log"
	"net/http"
	"strings"
	"time"
)

func push(method, requrl, contentType string, data interface{}, timeout int) (error, []byte) {
	var req *http.Request
	var err error
	var res []byte
	switch method {
	case "GET":
		req, err = http.NewRequest(method, requrl, nil)
		if err != nil {
			//log.Printf("Request url %s fail: new request error: %s", requrl, err)
			return err, res
		}
	case "POST":
		if contentType == "application/json" {
			body, err := json.Marshal(data)
			if err != nil {
				//log.Printf("Request url %s fail: parse body data %v error: %s", requrl, data, err)
				return err, res
			}
			req, err = http.NewRequest(method, requrl, strings.NewReader(string(body)))
			//req, err = http.NewRequest(method, requrl, bytes.NewBuffer(body))
			if err != nil {
				//log.Printf("Request url %s fail: new request error: %s", requrl, err)
				return err, res
			}
			req.Header.Set("Content-Type", "application/json")
		} else {
			//log.Printf("Request url %s fail: http Content-Type %s is not supported", requrl, method)
			return errors.New("content-type err"), res
		}
	default:
		//log.Printf("Request url %s fail: http method %s is not supported", requrl, method)
		return errors.New("method not supported"), res
	}
	client := &http.Client{}
	if timeout != 0 {
		client.Timeout = time.Duration(timeout) * time.Second
	}
	resp, err := client.Do(req)
	if err != nil {
		//log.Printf("Request url %s fail: query error: %s", requrl, err)
		return err, res
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//log.Printf("Request url %s get request body fail", requrl)
		return err, res
	}
	return nil, body
}
