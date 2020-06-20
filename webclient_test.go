// Copyright 2020 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/hexya-addons/web/client"
	"github.com/hexya-addons/web/controllers"
	"github.com/hexya-addons/web/domains"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/server"
	. "github.com/smartystreets/goconvey/convey"
)

func TestWebclientCalls(t *testing.T) {
	hexyaURL := url.URL{
		Scheme: "http",
		Host:   "localhost:8585",
	}
	go func() { fmt.Println("server:", server.GetServer().Run(hexyaURL.Host)) }()
	cl := client.NewHexyaClient(hexyaURL.String())
	// Wait for server to come up
	time.Sleep(500 * time.Millisecond)
	Convey("Testing direct HTTP Calls to web client", t, func() {
		err := cl.Login("admin", "admin")
		So(err, ShouldBeNil)
		cookies := cl.Jar.Cookies(&hexyaURL)
		So(cookies, ShouldHaveLength, 1)
		So(cookies[0].Name, ShouldEqual, "hexya-session")
		Convey("Search-reading users", func() {
			raw, err := cl.RPC("/web/dataset/search_read", "call", controllers.SearchReadParams{
				Model: "User",
				Context: *types.NewContext().
					WithKey("company_id", 1).
					WithKey("lang", "en_US").
					WithKey("tz", "").
					WithKey("uid", 1),
				Domain: domains.Domain{},
				Fields: []string{"action_id", "active", "company_id", "company_ids", "email", "group_ids", "id", "image",
					"lang", "login", "name", "partner_id", "signature", "tz", "tz_offset", "display_name", "__last_update"},
				Limit: 80,
				Sort:  "",
			})
			So(err, ShouldBeNil)
			var res map[string]interface{}
			err = json.Unmarshal(raw, &res)
			So(err, ShouldBeNil)
			So(res, ShouldContainKey, "records")
			So(res, ShouldContainKey, "length")
			So(res["length"], ShouldEqual, 4)
			So(res["records"], ShouldHaveLength, 4)
		})
		Convey("Read on company", func() {
			raw, err := cl.RPC("/web/dataset/call_kw", "call", controllers.CallParams{
				Model:  "Company",
				Method: "read",
				Args: []json.RawMessage{
					json.RawMessage(`[1]`),
					json.RawMessage(`["display_name","name"]`),
				},
				KWArgs: map[string]json.RawMessage{
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"","uid":1}`),
				},
			})
			So(err, ShouldBeNil)
			var res []map[string]interface{}
			err = json.Unmarshal(raw, &res)
			So(err, ShouldBeNil)
			So(res, ShouldHaveLength, 1)
			So(res[0], ShouldContainKey, "display_name")
			So(res[0]["display_name"], ShouldEqual, "Your Company")
			So(res[0], ShouldContainKey, "name")
			So(res[0]["name"], ShouldEqual, "Your Company")
			So(res[0], ShouldContainKey, "id")
			So(res[0]["id"], ShouldEqual, 1)
		})
	})
}
