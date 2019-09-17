// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hexya-addons/web/controllers"
	"github.com/hexya-addons/web/webdata"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	. "github.com/smartystreets/goconvey/convey"
)

func TestExecute(t *testing.T) {
	var newPartnerID, newCompanyID int64
	Convey("Testing Execute function", t, func() {
		Convey("Creating a Company", func() {
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "Company",
				Method: "create",
				Args: []json.RawMessage{json.RawMessage(`{"logo": false, "name": "Company4", "tagline": false, 
"street": false, "street2": false, "city": false, "state_id": false, "zip": false, "country_id": false, 
"website": false, "phone": false, "fax": false, "email": false, "vat": false, "company_registry": false, 
"parent_id": false, "currency_id": 45}`)},
				KWArgs: map[string]json.RawMessage{
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"","uid":1,"params":{"action":"base_action_res_company_form"}}`)},
			})
			So(err, ShouldBeNil)
			rID, ok := res.(int64)
			So(ok, ShouldBeTrue)
			So(rID, ShouldNotEqual, 0)
			newCompanyID = rID
		})
		Convey("Creating and SearchReading a User", func() {
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "User",
				Method: "create",
				Args: []json.RawMessage{json.RawMessage(fmt.Sprintf(`{"active":true,"image":false,"name":"User","email":false,
"login":"user@example.com","company_ids":[[6,false,[%d]]],"company_id":%d,"group_ids":[],"lang":false,"tz":false,
"action_id":false,"signature":false,"parent_id":3,"user_id":1,"categories_ids":[]}`, newCompanyID, newCompanyID))},
				KWArgs: map[string]json.RawMessage{
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"","uid":1,"params":{"action":"base_action_res_users"}}`),
				},
			})
			So(err, ShouldBeNil)
			rID, ok := res.(int64)
			So(ok, ShouldBeTrue)
			So(rID, ShouldNotEqual, 0)

			res, err = controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "User",
				Method: "search_read",
				Args: []json.RawMessage{
					json.RawMessage(fmt.Sprintf(`[["id","in",[%d]]]`, rID)),
					json.RawMessage(`["action_id","active","company_id","company_ids","email","group_ids","id","image",
"lang","login","name","partner_id","signature","tz","tz_offset","display_name","__last_update"]`),
				},
				KWArgs: map[string]json.RawMessage{
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"","uid":1,
"params":{"action":"base_action_res_users","view_type":"form","model":"User","_push_me":false},
"bin_size":true,"active_test":false}`),
				},
			})
			So(err, ShouldBeNil)
			srr, ok := res.([]models.RecordData)
			So(ok, ShouldBeTrue)
			So(srr, ShouldHaveLength, 1)
			So(srr[0].Underlying().Keys(), ShouldContain, "company_id")
			n := srr[0].Underlying().Get("name")
			So(n, ShouldEqual, "User")
			c := srr[0].Underlying().Get("company_id")
			cID, _ := c.(webdata.RecordIDWithName)
			So(cID.ID, ShouldEqual, newCompanyID)
			So(cID.Name, ShouldEqual, "Company4")
			a := srr[0].Underlying().Get("active")
			So(a, ShouldBeTrue)
		})

		Convey("Create a Partner", func() {
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "Partner",
				Method: "create",
				Args: []json.RawMessage{json.RawMessage(`{"active":true,"image_medium":false,"is_company":false,
"company_type":"person","name":"Nicolas PIGANEAU","parent_id":3,"company_name":false,"type":"contact","website":false,
"categories_ids":[],"function":false,"phone":false,"mobile":false,"fax":false,"users_ids":[],"email":false,
"title_id":false,"lang":"en_US","children_ids":[],"comment":false,"customer":true,"user_id":1,"supplier":false,
"ref":false,"company_id":false}`)},
				KWArgs: map[string]json.RawMessage{
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"","uid":1,
"search_default_customer":true,"params":{"action":"base_action_partner_form"}}`),
				},
			})
			So(err, ShouldBeNil)
			rID, ok := res.(int64)
			So(ok, ShouldBeTrue)
			So(rID, ShouldNotEqual, 0)
			newPartnerID = rID
		})

		Convey("Updating a Partner", func() {
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "Partner",
				Method: "write",
				Args: []json.RawMessage{
					json.RawMessage(fmt.Sprintf(`[%d]`, newPartnerID)),
					json.RawMessage(`{"company_type":"company","children_ids":[[0,false,{"type":"contact",
"street":"31 Hong Kong street","street2":false,"city":"Taipei","state_id":false,"zip":"106","country_id":221,
"name":"Nick Jr","title_id":false,"function":false,"email":false,"phone":false,"mobile":false,"comment":false,
"supplier":false,"customer":true,"lang":"en_US","image":false}]]}`),
				},
				KWArgs: map[string]json.RawMessage{
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"","uid":1,
"search_default_customer":true,"params":{"action":"base_action_partner_form"}}`),
				},
			})
			So(err, ShouldBeNil)
			rb, ok := res.(bool)
			So(ok, ShouldBeTrue)
			So(rb, ShouldBeTrue)
		})

		Convey("Read on Company", func() {
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "Company",
				Method: "read",
				Args: []json.RawMessage{
					json.RawMessage(fmt.Sprintf(`[%d]`, newCompanyID)),
					json.RawMessage(`["display_name","name"]`),
				},
				KWArgs: map[string]json.RawMessage{
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"","uid":1}`),
				},
			})
			So(err, ShouldBeNil)
			rr, ok := res.([]models.RecordData)
			So(rr, ShouldHaveLength, 1)
			So(ok, ShouldBeTrue)
			So(rr[0].Underlying().Keys(), ShouldContain, "display_name")
			n := rr[0].Underlying().Get("name")
			So(n, ShouldEqual, "Company4")
			id := rr[0].Underlying().Get("id")
			So(id, ShouldEqual, newCompanyID)
			dn := rr[0].Underlying().Get("display_name")
			So(dn, ShouldEqual, "Company4")
		})

		Convey("DefaultGet on Partner", func() {
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "Partner",
				Method: "default_get",
				Args: []json.RawMessage{
					json.RawMessage(`["active","categories_ids","children_ids","city","comment","commercial_partner_id",
"company_id","company_name","company_type","country_id","customer","email","fax","function","image_medium","is_company",
"lang","mobile","name","parent_id","phone","ref","state_id","street","street2","supplier","title_id","type","user_id",
"users_ids","website","zip"]`),
				},
				KWArgs: map[string]json.RawMessage{
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"","uid":1,
"params":{"action":"base_action_partner_form","menu_id":"menu_partner","_push_me":false},
"search_default_customer":true}`),
				},
			})
			So(err, ShouldBeNil)
			rr, ok := res.(models.RecordData)
			So(ok, ShouldBeTrue)
			So(rr.Underlying().FieldMap, ShouldContainKey, "active")
			So(rr.Underlying().FieldMap["active"], ShouldBeTrue)
			So(rr.Underlying().FieldMap, ShouldContainKey, "is_company")
			So(rr.Underlying().FieldMap["is_company"], ShouldBeFalse)
			So(rr.Underlying().FieldMap, ShouldContainKey, "company_type")
			So(rr.Underlying().FieldMap["company_type"], ShouldEqual, "person")
		})

		Convey("DefaultGet on User (embedding model)", func() {
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "User",
				Method: "default_get",
				Args: []json.RawMessage{
					json.RawMessage(`["action_id","active","company_id","company_ids","email","group_ids","id","image","lang","login","name","partner_id","signature","tz","tz_offset"]`),
				},
				KWArgs: map[string]json.RawMessage{
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"","uid":1,"params":{"action":"base_action_res_users"}}`),
				},
			})
			So(err, ShouldBeNil)
			rr, ok := res.(models.RecordData)
			So(ok, ShouldBeTrue)
			So(rr.Underlying().FieldMap, ShouldContainKey, "active")
			So(rr.Underlying().FieldMap["active"], ShouldBeTrue)
			So(rr.Underlying().FieldMap, ShouldContainKey, "company_id")
			So(rr.Underlying().FieldMap["company_id"], ShouldHaveSameTypeAs, webdata.RecordIDWithName{})
			So(rr.Underlying().FieldMap["company_id"].(webdata.RecordIDWithName).ID, ShouldEqual, 1)
			So(rr.Underlying().FieldMap["company_id"].(webdata.RecordIDWithName).Name, ShouldEqual, "Your Company")
			So(rr.Underlying().FieldMap, ShouldContainKey, "lang")
			So(rr.Underlying().FieldMap["lang"], ShouldEqual, "en_US")
		})

		Convey("Onchange call on Partner before creation", func() {
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "Partner",
				Method: "onchange",
				Args: []json.RawMessage{
					json.RawMessage(`[]`),
				},
				KWArgs: map[string]json.RawMessage{
					"values": json.RawMessage(`{"id":false,"active":true,"image_medium":false,"is_company":false,
"commercial_partner_id":3,"company_type":"person","name":false,"parent_id":3,"company_name":false,
"type":"contact","street":false,"street2":false,"city":false,"state_id":false,"zip":false,"country_id":false,
"website":false,"categories_ids":[],"function":false,"phone":false,"mobile":false,"fax":false,"users_ids":[],
"email":false,"title_id":false,"lang":"en_US","children_ids":[],"comment":false,"customer":true,"user_id":false,
"supplier":false,"ref":false,"company_id":false}`),
					"field_name": json.RawMessage(`["active","image_medium","is_company","commercial_partner_id",
"company_type","name","parent_id","company_name","type","street","street2","city","state_id","zip","country_id",
"website","categories_ids","function","phone","mobile","fax","users_ids","email","title_id","lang","children_ids",
"comment","customer","user_id",
"supplier","ref","company_id"]`),
					"field_onchange": json.RawMessage(`{"active":"","image_medium":"","is_company":"",
"commercial_partner_id":"","company_type":"1","name":"","parent_id":"1","company_name":"","type":"","street":"",
"street2":"","city":"","state_id":"","zip":"","country_id":"","website":"","categories_ids":"","function":"",
"phone":"","mobile":"","fax":"","users_ids":"","email":"1","title_id":"","lang":"","children_ids":"",
"children_ids.city":"","children_ids.comment":"","children_ids.country_id":"","children_ids.customer":"",
"children_ids.email":"1","children_ids.function":"","children_ids.image":"","children_ids.lang":"",
"children_ids.mobile":"","children_ids.name":"","children_ids.phone":"","children_ids.state_id":"",
"children_ids.street":"","children_ids.street2":"","children_ids.supplier":"","children_ids.title_id":"",
"children_ids.type":"","children_ids.zip":"","children_ids.active":"","children_ids.display_name":"",
"children_ids.is_company":"","children_ids.parent_id":"1","children_ids.user_id":"","comment":"",
"customer":"","user_id":"","supplier":"","ref":"","company_id":""}`),
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"","uid":1,
"search_default_customer":true,"params":{"action":"base_action_partner_form"}}`),
				},
			})
			So(err, ShouldBeNil)
			ocr, ok := res.(models.OnchangeResult)
			So(ok, ShouldBeTrue)
			fm := ocr.Value.Underlying().FieldMap
			So(fm, ShouldHaveLength, 7)
			So(fm, ShouldContainKey, "is_company")
			So(fm["is_company"], ShouldBeFalse)
		})

		Convey("Onchange call on Partner during modification", func() {
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "Partner",
				Method: "onchange",
				Args: []json.RawMessage{
					json.RawMessage(`[]`),
				},
				KWArgs: map[string]json.RawMessage{
					"values": json.RawMessage(`{"id":false,"active":true,"image_medium":false,"is_company":false,
"commercial_partner_id":3,"company_type":"company","name":false,"parent_id":3,"company_name":false,
"type":"contact","street":false,"street2":false,"city":false,"state_id":false,"zip":false,"country_id":false,
"website":false,"categories_ids":[],"function":false,"phone":false,"mobile":false,"fax":false,"users_ids":[],
"email":false,"title_id":false,"lang":"en_US","children_ids":[],"comment":false,"customer":true,"user_id":false,
"supplier":false,"ref":false,"company_id":false}`),
					"field_name": json.RawMessage(`["company_type"]`),
					"field_onchange": json.RawMessage(`{"active":"","image_medium":"","is_company":"",
"commercial_partner_id":"","company_type":"1","name":"","parent_id":"1","company_name":"","type":"","street":"",
"street2":"","city":"","state_id":"","zip":"","country_id":"","website":"","categories_ids":"","function":"",
"phone":"","mobile":"","fax":"","users_ids":"","email":"1","title_id":"","lang":"","children_ids":"",
"children_ids.city":"","children_ids.comment":"","children_ids.country_id":"","children_ids.customer":"",
"children_ids.email":"1","children_ids.function":"","children_ids.image":"","children_ids.lang":"",
"children_ids.mobile":"","children_ids.name":"","children_ids.phone":"","children_ids.state_id":"",
"children_ids.street":"","children_ids.street2":"","children_ids.supplier":"","children_ids.title_id":"",
"children_ids.type":"","children_ids.zip":"","children_ids.active":"","children_ids.display_name":"",
"children_ids.is_company":"","children_ids.parent_id":"1","children_ids.user_id":"","comment":"","customer":"",
"user_id":"","supplier":"","ref":"","company_id":""}`),
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"","uid":1,
"search_default_customer":true,"params":{"action":"base_action_partner_form"}}`),
				},
			})
			So(err, ShouldBeNil)
			ocr, ok := res.(models.OnchangeResult)
			So(ok, ShouldBeTrue)
			fm := ocr.Value.Underlying().FieldMap
			So(fm, ShouldHaveLength, 1)
			So(fm, ShouldContainKey, "is_company")
			So(fm["is_company"], ShouldBeTrue)

		})

		Convey("NameSearch on Country", func() {
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "Country",
				Method: "name_search",
				Args:   []json.RawMessage{},
				KWArgs: map[string]json.RawMessage{
					"name":     json.RawMessage(`""`),
					"args":     json.RawMessage(`[]`),
					"operator": json.RawMessage(`"ilike"`),
					"context":  json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"","uid":1}`),
					"limit":    json.RawMessage(`8`),
				},
			})
			So(err, ShouldBeNil)
			rin, ok := res.([]webdata.RecordIDWithName)
			So(ok, ShouldBeTrue)
			So(rin, ShouldHaveLength, 8)
			So(rin[0].ID, ShouldEqual, 1)
			So(rin[0].Name, ShouldEqual, "Afghanistan, Islamic State of")
		})

		Convey("SearchReading Currencies", func() {
			var srp controllers.SearchReadParams
			data := []byte(`{"model":"Currency","fields":["name","symbol","rates_ids","date","rate","active"],
"domain":[],"context":{"company_id":1,"lang":"en_US","tz":"","uid":1,"params":{"action":"base_action_currency_form",
"menu_id":"menu_currency","_push_me":false},"active_test":false,"bin_size":true},"offset":0,"limit":10,"sort":""}`)
			err := json.Unmarshal(data, &srp)
			So(err, ShouldBeNil)
			res, err := controllers.SearchRead(security.SuperUserID, srp)
			So(err, ShouldBeNil)
			So(res.Length, ShouldEqual, 173)
			So(res.Records, ShouldHaveLength, 10)
		})
	})
}
