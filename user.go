// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/fields"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/m"
)

var fields_User = map[string]models.FieldDefinition{
	"SidebarVisible": fields.Boolean{
		String: "Show App Sidebar", Default: models.DefaultValue(true),
	},
}

func user_SelfWritableFields(rs m.UserSet) map[string]bool {
	res := rs.Super().SelfWritableFields()
	res["SidebarVisible"] = true
	return res
}

func user_SelfReadableFields(rs m.UserSet) map[string]bool {
	res := rs.Super().SelfReadableFields()
	res["SidebarVisible"] = true
	return res
}

func init() {
	h.User().AddFields(fields_User)
	h.User().Methods().SelfWritableFields().Extend(user_SelfWritableFields)
	h.User().Methods().SelfReadableFields().Extend(user_SelfReadableFields)
}
