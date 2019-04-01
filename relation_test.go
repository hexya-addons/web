// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hexya-addons/web/controllers"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
	. "github.com/smartystreets/goconvey/convey"
)

func Test2ManyRelations(t *testing.T) {
	Convey("Testing 2many relations modification with client triplets", t, func() {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Testing many2many '6' triplet", func() {
				adminGroup := h.Group().Search(env, q.Group().GroupID().Equals(security.GroupAdminID))
				jsonGroupsData := fmt.Sprintf("[[6, 0, [%d]]]", adminGroup.ID())
				var groupsData interface{}
				json.Unmarshal([]byte(jsonGroupsData), &groupsData)

				rc := env.Pool("User")
				controllers.MethodAdapters["Create"](rc, "Create", []interface{}{
					models.NewModelData(rc.Model(), models.FieldMap{
						"Name":   "Test User",
						"Login":  "test_user",
						"Groups": groupsData,
					}),
				})
				user := h.User().Search(env, q.User().Login().Equals("test_user"))
				So(user.Len(), ShouldEqual, 1)
				So(user.Groups().Len(), ShouldEqual, 1)
				So(user.Groups().ID(), ShouldEqual, adminGroup.ID())
			})
		})
	})
}
