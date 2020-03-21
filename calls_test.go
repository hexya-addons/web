// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hexya-addons/web/controllers"
	"github.com/hexya-addons/web/webtypes"
	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/operator"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/m"
	. "github.com/smartystreets/goconvey/convey"
)

func TestExecute(t *testing.T) {
	var newPartnerID, newCompanyID int64
	Convey("Testing Execute function", t, func() {
		Convey("Creating a Company", func() {
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "Company",
				Method: "create",
				Args: []json.RawMessage{json.RawMessage(`{"logo": false, "name": "Company4", 
"street": false, "street2": false, "city": false, "state_id": false, "zip": false, "country_id": false, 
"website": false, "phone": false, "email": false, "vat": false, "company_registry": false, 
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
"action_id":false,"signature":false,"parent_id":3,"user_id":1,"category_ids":[]}`, newCompanyID, newCompanyID))},
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
			n := srr[0].Underlying().Get(models.NewFieldName("Name", "name"))
			So(n, ShouldEqual, "User")
			c := srr[0].Underlying().Get(models.NewFieldName("Company", "company_id"))
			cID, _ := c.(webtypes.RecordIDWithName)
			So(cID.ID, ShouldEqual, newCompanyID)
			So(cID.Name, ShouldEqual, "Company4")
			a := srr[0].Underlying().Get(models.NewFieldName("Active", "active"))
			So(a, ShouldBeTrue)
		})

		Convey("Create a Partner", func() {
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "Partner",
				Method: "create",
				Args: []json.RawMessage{json.RawMessage(`{"active":true,"image_medium":false,"is_company":false,
"company_type":"person","name":"Nicolas PIGANEAU","parent_id":3,"company_name":false,"type":"contact","website":false,
"category_ids":[],"function":false,"phone":false,"mobile":false,"user_ids":[],"email":false,
"title_id":false,"lang":"en_US","children_ids":[],"comment":false,"user_id":1,
"ref":false,"company_id":false}`)},
				KWArgs: map[string]json.RawMessage{
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"","uid":1,
"params":{"action":"base_action_partner_form"}}`),
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
"lang":"en_US","image":false}]]}`),
				},
				KWArgs: map[string]json.RawMessage{
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"","uid":1,
"params":{"action":"base_action_partner_form"}}`),
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
			n := rr[0].Underlying().Get(models.NewFieldName("Name", "name"))
			So(n, ShouldEqual, "Company4")
			id := rr[0].Underlying().Get(models.ID)
			So(id, ShouldEqual, newCompanyID)
			dn := rr[0].Underlying().Get(models.NewFieldName("DisplayName", "display_name"))
			So(dn, ShouldEqual, "Company4")
		})

		Convey("DefaultGet on User", func() {
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "User",
				Method: "default_get",
				Args: []json.RawMessage{
					json.RawMessage(`["id","groups_count","accesses_count","rules_count","active_partner","partner_id","image_1920","name","email","login","company_ids","company_id","companies_count","active","lang","tz","tz_offset","action_id","signature"]`),
				},
				KWArgs: map[string]json.RawMessage{
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"Europe/Brussels","uid":1,"allowed_company_ids":[1],
"params":{"action":"base_action_partner_form","menu_id":"menu_partner","_push_me":false}}`),
				},
			})
			So(err, ShouldBeNil)
			rr, ok := res.(models.RecordData)
			So(ok, ShouldBeTrue)
			So(rr.Underlying().FieldMap, ShouldContainKey, "company_id")
			So(rr.Underlying().FieldMap["company_id"], ShouldEqual, 1)
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
			So(rr.Underlying().FieldMap["company_id"], ShouldHaveSameTypeAs, int64(1))
			So(rr.Underlying().FieldMap["company_id"], ShouldEqual, 1)
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
					"values": json.RawMessage(`{"id":false,"active":true,"image":false,"is_company":false,
"commercial_partner_id":false,"company_type":false,"name":false,"parent_id":false,"company_name":false,
"type":"contact","street":false,"street2":false,"city":false,"state_id":false,"zip":false,"country_id":false,
"website":false,"category_ids":[],"function":false,"phone":false,"mobile":false,"user_ids":[],
"email":false,"title_id":false,"lang":"en_US","children_ids":[],"comment":false,"user_id":false,
"ref":false,"company_id":1}`),
					"field_name": json.RawMessage(`["active","image","is_company","commercial_partner_id",
"company_type","name","parent_id","company_name","type","street","street2","city","state_id","zip","country_id",
"website","category_ids","function","phone","mobile","user_ids","email","title_id","lang","children_ids",
"comment","user_id","ref","company_id"]`),
					"field_onchange": json.RawMessage(`{"active":"1","image":"","is_company":"1",
"commercial_partner_id":"1","company_type":"1","name":"1","parent_id":"1","company_name":"1","type":"1","street":"",
"street2":"","city":"","state_id":"","zip":"","country_id":"1","website":"","category_ids":"","function":"","phone":"",
"mobile":"","user_ids":"","user_ids.login_date":"","user_ids.lang":"","user_ids.name":"","user_ids.login":"1",
"email":"1","title_id":"","lang":"","children_ids":"1","children_ids.comment":"","children_ids.function":"",
"children_ids.color":"","children_ids.image":"","children_ids.street":"","children_ids.city":"",
"children_ids.display_name":"","children_ids.zip":"","children_ids.title_id":"","children_ids.country_id":"1",
"children_ids.parent_id":"1","children_ids.email":"1","children_ids.is_company":"1",
"children_ids.street2":"","children_ids.lang":"",
"children_ids.name":"1","children_ids.phone":"","children_ids.mobile":"","children_ids.type":"1",
"children_ids.state_id":"","comment":"","user_id":"","ref":"","company_id":""}`),
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"Europe/Brussels","uid":1,
"params":{"action":"base_action_partner_form"}}`),
				},
			})
			So(err, ShouldBeNil)
			ocr, ok := res.(webtypes.OnChangeResult)
			So(ok, ShouldBeTrue)
			fm := ocr.Value.Underlying().FieldMap
			So(fm, ShouldHaveLength, 2)
			So(fm, ShouldContainKey, "company_type")
			So(fm, ShouldContainKey, "commercial_partner_id")
			So(fm["company_type"], ShouldEqual, "person")
			So(fm["commercial_partner_id"], ShouldBeFalse)
			So(ocr.Warning, ShouldBeEmpty)
		})

		Convey("Onchange call on Partner during modification of company_type", func() {
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "Partner",
				Method: "onchange",
				Args: []json.RawMessage{
					json.RawMessage(`[]`),
				},
				KWArgs: map[string]json.RawMessage{
					"values": json.RawMessage(`{"id":false,"active":true,"image":false,"is_company":false,
"commercial_partner_id":false,"company_type":"company","name":false,"parent_id":false,"company_name":false,
"type":"contact","street":false,"street2":false,"city":false,"state_id":false,"zip":false,"country_id":false,
"website":false,"category_ids":[],"function":false,"phone":false,"mobile":false,"user_ids":[],
"email":false,"title_id":false,"lang":"en_US","children_ids":[],"comment":false,"user_id":false,
"ref":false,"company_id":1}`),
					"field_name": json.RawMessage(`["company_type"]`),
					"field_onchange": json.RawMessage(`{"active":"1","image":"","is_company":"1",
"commercial_partner_id":"1","company_type":"1","name":"1","parent_id":"1","company_name":"1","type":"1","street":"",
"street2":"","city":"","state_id":"","zip":"","country_id":"1","website":"","category_ids":"","function":"",
"phone":"","mobile":"","user_ids":"","user_ids.login_date":"","user_ids.lang":"","user_ids.name":"",
"user_ids.login":"1","email":"1","title_id":"","lang":"","children_ids":"1","children_ids.comment":"",
"children_ids.function":"","children_ids.color":"","children_ids.image":"","children_ids.street":"",
"children_ids.city":"","children_ids.display_name":"","children_ids.zip":"","children_ids.title_id":"",
"children_ids.country_id":"1","children_ids.parent_id":"1","children_ids.email":"1",
"children_ids.is_company":"1","children_ids.street2":"",
"children_ids.lang":"","children_ids.name":"1","children_ids.phone":"","children_ids.mobile":"","children_ids.type":"1",
"children_ids.state_id":"","comment":"","user_id":"","ref":"","company_id":""}`),
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"Europe/Brussels","uid":1,
"params":{"action":"base_action_partner_form"}}`),
				},
			})
			So(err, ShouldBeNil)
			ocr, ok := res.(webtypes.OnChangeResult)
			So(ok, ShouldBeTrue)
			fm := ocr.Value.Underlying().FieldMap
			So(fm, ShouldHaveLength, 2)
			So(fm, ShouldContainKey, "is_company")
			So(fm, ShouldContainKey, "commercial_partner_id")
			So(fm["is_company"], ShouldBeTrue)
			So(fm["commercial_partner_id"], ShouldBeFalse)
			So(ocr.Warning, ShouldBeEmpty)
		})

		Convey("Onchange call on Partner during modification of country", func() {
			var belgiumID int64
			So(models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				belgiumID = h.Country().NewSet(env).GetRecord("base_be").ID()
			}), ShouldBeNil)
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "Partner",
				Method: "onchange",
				Args: []json.RawMessage{
					json.RawMessage(`[]`),
				},
				KWArgs: map[string]json.RawMessage{
					"values": json.RawMessage(fmt.Sprintf(`{"id":false,"active":true,"image":false,"is_company":false,
"commercial_partner_id":false,"company_type":"company","name":false,"parent_id":false,"company_name":false,
"type":"contact","street":false,"street2":false,"city":false,"state_id":false,"zip":false,"country_id":%d,
"website":false,"category_ids":[],"function":false,"phone":false,"mobile":false,"user_ids":[],
"email":false,"title_id":false,"lang":"en_US","children_ids":[],"comment":false,"user_id":false,
"ref":false,"company_id":1}`, belgiumID)),
					"field_name": json.RawMessage(`["country_id"]`),
					"field_onchange": json.RawMessage(`{"active":"1","image":"","is_company":"1",
"commercial_partner_id":"1","company_type":"1","name":"1","parent_id":"1","company_name":"1","type":"1","street":"",
"street2":"","city":"","state_id":"","zip":"","country_id":"1","website":"","category_ids":"","function":"",
"phone":"","mobile":"","user_ids":"","user_ids.login_date":"","user_ids.lang":"","user_ids.name":"",
"user_ids.login":"1","email":"1","title_id":"","lang":"","children_ids":"1","children_ids.comment":"",
"children_ids.function":"","children_ids.color":"","children_ids.image":"","children_ids.street":"",
"children_ids.city":"","children_ids.display_name":"","children_ids.zip":"","children_ids.title_id":"",
"children_ids.country_id":"1","children_ids.parent_id":"1","children_ids.email":"1",
"children_ids.is_company":"1","children_ids.street2":"",
"children_ids.lang":"","children_ids.name":"1","children_ids.phone":"","children_ids.mobile":"","children_ids.type":"1",
"children_ids.state_id":"","comment":"","user_id":"","ref":"","company_id":""}`),
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"Europe/Brussels","uid":1,
"params":{"action":"base_action_partner_form"}}`),
				},
			})
			So(err, ShouldBeNil)
			ocr, ok := res.(webtypes.OnChangeResult)
			So(ok, ShouldBeTrue)
			fm := ocr.Value.Underlying().FieldMap
			So(fm, ShouldHaveLength, 2)
			So(fm, ShouldContainKey, "is_company")
			So(fm, ShouldContainKey, "commercial_partner_id")
			So(fm["is_company"], ShouldBeTrue)
			So(fm["commercial_partner_id"], ShouldBeFalse)
			So(ocr.Warning, ShouldBeEmpty)
			So(ocr.Filters, ShouldContainKey, "state_id")
			So(ocr.Filters["state_id"], ShouldResemble, []interface{}{
				[]interface{}{"country_id", operator.Operator("="), belgiumID}})
		})

		Convey("Onchange call on Partner during modification of parent_id", func() {
			var nickID, agrolaitID, belgiumID int64
			So(models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				asusTek := h.Partner().NewSet(env).GetRecord("base_res_partner_1")
				nick := h.Partner().Create(env, h.Partner().NewData().
					SetName("Nicolas").
					SetParent(asusTek))
				nickID = nick.ID()
				agrolait := h.Partner().NewSet(env).GetRecord("base_res_partner_2")
				agrolaitID = agrolait.ID()
				belgiumID = h.Country().NewSet(env).GetRecord("base_be").ID()
			}), ShouldBeNil)
			So(nickID, ShouldNotEqual, 0)
			So(agrolaitID, ShouldNotEqual, 0)
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "Partner",
				Method: "onchange",
				Args: []json.RawMessage{
					json.RawMessage(`[]`),
				},
				KWArgs: map[string]json.RawMessage{
					"values": json.RawMessage(fmt.Sprintf(`{"id":%d,"active":true,"image":"16.51 Kb",
"is_company":false,"commercial_partner_id":%d,"company_type":"person","name":"Nicolas","parent_id":%d,
"company_name":false,"type":"contact","street":false,"street2":false,"city":false,"state_id":false,"zip":false,
"country_id":false,"website":false,"category_ids":[],"function":false,"phone":false,"mobile":false,
"user_ids":[],"email":false,"title_id":false,"lang":"en_US","children_ids":[],"comment":false,
"user_id":false,"ref":false,"company_id":1}`, nickID, nickID, agrolaitID)),
					"field_name": json.RawMessage(`["parent_id"]`),
					"field_onchange": json.RawMessage(`{"active":"1","image":"","is_company":"1",
"commercial_partner_id":"1","company_type":"1","name":"1","parent_id":"1","company_name":"1","type":"1","street":"",
"street2":"","city":"","state_id":"","zip":"","country_id":"1","website":"","category_ids":"","function":"",
"phone":"","mobile":"","user_ids":"","user_ids.login_date":"","user_ids.lang":"","user_ids.name":"",
"user_ids.login":"1","email":"1","title_id":"","lang":"","children_ids":"1","children_ids.comment":"",
"children_ids.function":"","children_ids.color":"","children_ids.image":"","children_ids.street":"",
"children_ids.city":"","children_ids.display_name":"","children_ids.zip":"","children_ids.title_id":"",
"children_ids.country_id":"1","children_ids.parent_id":"1","children_ids.email":"1",
"children_ids.is_company":"1","children_ids.street2":"",
"children_ids.lang":"","children_ids.name":"1","children_ids.phone":"","children_ids.mobile":"","children_ids.type":"1",
"children_ids.state_id":"","comment":"","user_id":"","ref":"","company_id":""}`),
					"context": json.RawMessage(`{"company_id":1,"lang":"en_US","tz":"Europe/Brussels","uid":1,
"params":{"action":"base_action_partner_form"}}`),
				},
			})
			So(err, ShouldBeNil)
			ocr, ok := res.(webtypes.OnChangeResult)
			So(ok, ShouldBeTrue)
			fm := ocr.Value.Underlying().FieldMap
			So(fm, ShouldHaveLength, 5)
			So(fm, ShouldContainKey, "city")
			So(fm, ShouldContainKey, "street")
			So(fm, ShouldContainKey, "country_id")
			So(fm, ShouldContainKey, "zip")
			So(fm, ShouldContainKey, "commercial_partner_id")
			So(fm["city"], ShouldEqual, "Wavre")
			So(fm["street"], ShouldEqual, "69 rue de Namur")
			So(fm["country_id"], ShouldHaveSameTypeAs, webtypes.RecordIDWithName{})
			So(fm["country_id"].(webtypes.RecordIDWithName).ID, ShouldEqual, belgiumID)
			So(fm["country_id"].(webtypes.RecordIDWithName).Name, ShouldEqual, "Belgium")
			So(fm["zip"], ShouldEqual, "1300")
			So(fm["commercial_partner_id"], ShouldHaveSameTypeAs, webtypes.RecordIDWithName{})
			So(fm["commercial_partner_id"].(webtypes.RecordIDWithName).ID, ShouldEqual, agrolaitID)
			So(fm["commercial_partner_id"].(webtypes.RecordIDWithName).Name, ShouldEqual, "Agrolait")
			So(ocr.Warning, ShouldEqual, `Changing the company of a contact should only be done if it
was never correctly set. If an existing contact starts working for a new
company then a new contact should be created under that new
company. You can use the "Discard" button to abandon this change.`)
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
			rin, ok := res.([]webtypes.RecordIDWithName)
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

		Convey("CreateOrUpdate filter on users", func() {
			action := actions.Registry.MustGetByXMLID("base_action_res_users")
			res, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "ir.filters",
				Method: "create_or_replace",
				Args: []json.RawMessage{
					json.RawMessage(fmt.Sprintf(`{"name":"Users","context":"{\"group_by\":[\"lang\",\"name\"]}","domain":"[]","is_default":true,"user_id":1,"model_id":"res.users","action_id":%d,"sort":"[]"}`, action.ID)),
				},
			})
			So(err, ShouldBeNil)
			id, ok := res.(int64)
			So(ok, ShouldBeTrue)
			So(id, ShouldEqual, 1)

		})

		Convey("GetFilters on users", func() {
			action := actions.Registry.MustGetByXMLID("base_action_res_users")
			res2, err := controllers.Execute(security.SuperUserID, controllers.CallParams{
				Model:  "ir.filters",
				Method: "get_filters",
				Args: []json.RawMessage{
					json.RawMessage(`"User"`),
					json.RawMessage(fmt.Sprintf("%d", action.ID)),
				},
			})
			So(err, ShouldBeNil)
			filters, ok := res2.([]m.FilterData)
			So(ok, ShouldBeTrue)
			So(filters, ShouldHaveLength, 1)
			f := filters[0]
			So(f.Name(), ShouldEqual, "Users")
			So(f.IsDefault(), ShouldBeTrue)
			So(f.Action(), ShouldEqual, action.ID)
		})
	})
}
