// Copyright 2020 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

// Package client provides utilities to make client calls to the
// Hexya server. This is mainly useful for tests.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

// Hexya wraps a http.Client
type Hexya struct {
	http.Client
	Url string
}

// NewHexyaClient returns a new Hexya Client to make requests from.
func NewHexyaClient(url string) *Hexya {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		panic(err)
	}
	return &Hexya{
		Client: http.Client{
			Jar: jar,
		},
		Url: url,
	}
}

// Login to the Hexya server with the given username and password
func (hc *Hexya) Login(username, password string) error {
	vals := make(url.Values)
	vals.Add("login", username)
	vals.Add("password", password)
	_, err := hc.Client.PostForm(hc.Url+"/web/login", vals)
	if err != nil {
		return fmt.Errorf("error while logging: %s", err)
	}
	return nil
}

// RPC executes a RPC call to the given method on the given URI with the given params.
// params must be json serializable.
// Returned value is the result message.
func (hc *Hexya) RPC(uri, method string, params interface{}) (json.RawMessage, error) {
	dataMap := map[string]interface{}{
		"id":      0,
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	data, err := json.Marshal(dataMap)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling data: %s", err)
	}
	res, err := hc.Client.Post(hc.Url+uri, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error while sending request: %s", err)
	}
	resData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading response body: %s", err)
	}
	var resStruct struct {
		Result json.RawMessage `json:"result"`
	}
	err = json.Unmarshal(resData, &resStruct)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshalling response: %s. Response: %s", err, string(resData))
	}
	return resStruct.Result, nil
}
