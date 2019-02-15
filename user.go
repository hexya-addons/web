// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/m"
)

func init() {
	h.User().AddFields(map[string]models.FieldDefinition{
		"SidebarVisible": models.BooleanField{
			String: "Show App Sidebar", Default: models.DefaultValue(true),
		},
	})

	h.User().Methods().SelfWritableFields().Extend("",
		func(rs m.UserSet) map[string]bool {
			res := rs.Super().SelfWritableFields()
			res["SidebarVisible"] = true
			return res
		})

	h.User().Methods().SelfReadableFields().Extend("",
		func(rs m.UserSet) map[string]bool {
			res := rs.Super().SelfReadableFields()
			res["SidebarVisible"] = true
			return res
		})
}
