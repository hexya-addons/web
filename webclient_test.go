// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "github.com/hexya-addons/web/controllers"
	"github.com/hexya-erp/hexya/src/controllers"
	"github.com/hexya-erp/hexya/src/server"
	. "github.com/smartystreets/goconvey/convey"
)

func login() *http.Cookie {
	req := httptest.NewRequest(http.MethodPost, "/web/login", strings.NewReader(url.Values{
		"login":    []string{"admin"},
		"password": []string{"admin"},
	}.Encode()))
	w := httptest.NewRecorder()
	server.GetServer().ServeHTTP(w, req)
	for _, c := range w.Result().Cookies() {
		if c.Name == "hexya-session" {
			return c
		}
	}
	panic("No session cookie returned")
}

func performRequest(path string, data map[string]interface{}) *httptest.ResponseRecorder {
	params, _ := json.Marshal(data)
	reqData := server.RequestRPC{
		ID:     1,
		Method: "call",
		Params: params,
	}
	bs, _ := json.Marshal(reqData)
	body := bytes.NewReader(bs)
	sessionCookie := login()
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.AddCookie(sessionCookie)
	w := httptest.NewRecorder()
	server.GetServer().ServeHTTP(w, req)
	return w
}

func TestWebClient(t *testing.T) {
	controllers.BootStrap()
	Convey("Testing web client", t, func() {
		/*
			Convey("Read", func() {
			w := performRequest("/web/dataset/call_kw/", map[string]interface{}{
				"model":  "User",
				"method": "read",
				"args": []interface{}{
					[]interface{}{security.SuperUserID},
					[]interface{}{"name", "profile_id", "posts", "display_name", "__last_update"}},
				"kwargs": map[string]interface{}{
					"context": map[string]interface{}{
						"lang":     "en_US",
						"tz":       "",
						"uid":      1,
						"params":   map[string]interface{}{"action": "base_actions_ir_filters_view"},
						"bin_size": true,
					},
				},
			})
			respBytes, _ := ioutil.ReadAll(w.Body)
			fmt.Println(w.Code, string(respBytes))
		})*/
	})
}
