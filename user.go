// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/fields"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/m"
)

var fields_User = map[string]models.FieldDefinition{
	"ChatterPosition": fields.Selection{
		Selection: types.Selection{"normal": "Normal", "sided": "Sided"},
		String:    "Chatter Position", Default: models.DefaultValue("sided"),
	},
	"SidebarVisible": fields.Boolean{String: "Show App Sidebar", Default: models.DefaultValue(true)},
}

func user_SelfWritableFields(rs m.UserSet) map[string]bool {
	res := rs.Super().SelfWritableFields()
	res["ChatterPosition"] = true
	res["SidebarVisible"] = true
	return res
}

func user_SelfReadableFields(rs m.UserSet) map[string]bool {
	res := rs.Super().SelfReadableFields()
	res["ChatterPosition"] = true
	res["SidebarVisible"] = true
	return res
}

func user_ContextGet(rs m.UserSet) *types.Context {
	res := rs.Super().ContextGet()
	res = res.WithKey("chatter_position", rs.ChatterPosition())
	return res
}

func init() {
	h.User().AddFields(fields_User)
	h.User().Methods().SelfWritableFields().Extend(user_SelfWritableFields)
	h.User().Methods().SelfReadableFields().Extend(user_SelfReadableFields)
	h.User().Methods().ContextGet().Extend(user_ContextGet)
}
