// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"testing"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetFilters(t *testing.T) {
	Convey("Testing GetFilters logic", t, func() {
		So(models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			demoUser := h.User().Search(env, q.User().HexyaExternalID().Equals("base_user_demo"))
			adminUser := h.User().Search(env, q.User().ID().Equals(security.SuperUserID))
			Convey("Testing own filters", func() {
				h.Filter().Create(env, h.Filter().NewData().
					SetName("a").
					SetUser(demoUser).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("b").
					SetUser(demoUser).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("c").
					SetUser(demoUser).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("d").
					SetUser(demoUser).
					SetResModel("Filter"))
				filters := h.Filter().NewSet(env).Sudo(demoUser.ID()).GetFilters("Filter", "")
				So(filters, ShouldHaveLength, 4)
				So(filters[0].Name(), ShouldEqual, "a")
				So(filters[1].Name(), ShouldEqual, "b")
				So(filters[2].Name(), ShouldEqual, "c")
				So(filters[3].Name(), ShouldEqual, "d")
				So(filters[0].IsDefault(), ShouldBeFalse)
				So(filters[0].User().Equals(demoUser), ShouldBeTrue)
				So(filters[0].Domain(), ShouldEqual, "[]")
				So(filters[0].Context(), ShouldEqual, "{}")
				So(filters[0].Sort(), ShouldEqual, "[]")
			})
			Convey("Test Global Filters", func() {
				h.Filter().Create(env, h.Filter().NewData().
					SetName("a").
					SetUser(nil).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("b").
					SetUser(nil).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("c").
					SetUser(nil).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("d").
					SetUser(nil).
					SetResModel("Filter"))
				filters := h.Filter().NewSet(env).Sudo(demoUser.ID()).GetFilters("Filter", "")
				So(filters, ShouldHaveLength, 4)
				So(filters[0].Name(), ShouldEqual, "a")
				So(filters[1].Name(), ShouldEqual, "b")
				So(filters[2].Name(), ShouldEqual, "c")
				So(filters[3].Name(), ShouldEqual, "d")
				So(filters[0].IsDefault(), ShouldBeFalse)
				So(filters[0].User().Len(), ShouldEqual, 0)
				So(filters[0].Domain(), ShouldEqual, "[]")
				So(filters[0].Context(), ShouldEqual, "{}")
				So(filters[0].Sort(), ShouldEqual, "[]")
			})
			Convey("Test No Third Party Filters", func() {
				h.Filter().Create(env, h.Filter().NewData().
					SetName("a").
					SetUser(nil).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("b").
					SetUser(adminUser).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("c").
					SetUser(demoUser).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("d").
					SetUser(adminUser).
					SetResModel("Filter"))
				filters := h.Filter().NewSet(env).Sudo(demoUser.ID()).GetFilters("Filter", "")
				So(filters, ShouldHaveLength, 2)
				So(filters[0].Name(), ShouldEqual, "a")
				So(filters[1].Name(), ShouldEqual, "c")
				So(filters[0].IsDefault(), ShouldBeFalse)
				So(filters[0].User().Len(), ShouldEqual, 0)
				So(filters[0].Domain(), ShouldEqual, "[]")
				So(filters[0].Context(), ShouldEqual, "{}")
				So(filters[0].Sort(), ShouldEqual, "[]")
				So(filters[1].IsDefault(), ShouldBeFalse)
				So(filters[1].User().Equals(demoUser), ShouldBeTrue)
				So(filters[1].Domain(), ShouldEqual, "[]")
				So(filters[1].Context(), ShouldEqual, "{}")
				So(filters[1].Sort(), ShouldEqual, "[]")
			})
		}), ShouldBeNil)
	})
}

func TestOwnDefaults(t *testing.T) {
	Convey("Test Own Defaults", t, func() {
		So(models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			demoUser := h.User().Search(env, q.User().HexyaExternalID().Equals("base_user_demo"))
			Convey("New No Filter", func() {
				h.Filter().NewSet(env).Sudo(demoUser.ID()).CreateOrReplace(h.Filter().NewData().
					SetName("a").
					SetResModel("Filter").
					SetUser(demoUser).
					SetIsDefault(true))
				filters := h.Filter().NewSet(env).Sudo(demoUser.ID()).GetFilters("Filter", "")
				So(filters, ShouldHaveLength, 1)
				So(filters[0].Name(), ShouldEqual, "a")
				So(filters[0].User().Equals(demoUser), ShouldBeTrue)
				So(filters[0].IsDefault(), ShouldBeTrue)
				So(filters[0].Domain(), ShouldEqual, "[]")
				So(filters[0].Context(), ShouldEqual, "{}")
				So(filters[0].Sort(), ShouldEqual, "[]")
			})
			Convey("New Filter Not Default", func() {
				h.Filter().Create(env, h.Filter().NewData().
					SetName("a").
					SetUser(demoUser).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("b").
					SetUser(demoUser).
					SetResModel("Filter"))
				h.Filter().NewSet(env).Sudo(demoUser.ID()).CreateOrReplace(h.Filter().NewData().
					SetName("c").
					SetResModel("Filter").
					SetUser(demoUser).
					SetIsDefault(true))
				filters := h.Filter().NewSet(env).Sudo(demoUser.ID()).GetFilters("Filter", "")
				So(filters, ShouldHaveLength, 3)
				So(filters[0].Name(), ShouldEqual, "a")
				So(filters[0].User().Equals(demoUser), ShouldBeTrue)
				So(filters[0].IsDefault(), ShouldBeFalse)
				So(filters[0].Domain(), ShouldEqual, "[]")
				So(filters[0].Context(), ShouldEqual, "{}")
				So(filters[0].Sort(), ShouldEqual, "[]")
				So(filters[1].Name(), ShouldEqual, "b")
				So(filters[1].User().Equals(demoUser), ShouldBeTrue)
				So(filters[1].IsDefault(), ShouldBeFalse)
				So(filters[2].Name(), ShouldEqual, "c")
				So(filters[2].User().Equals(demoUser), ShouldBeTrue)
				So(filters[2].IsDefault(), ShouldBeTrue)
			})
			Convey("New Filter Existing Default", func() {
				h.Filter().Create(env, h.Filter().NewData().
					SetName("a").
					SetUser(demoUser).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("b").
					SetUser(demoUser).
					SetResModel("Filter").
					SetIsDefault(true))
				h.Filter().NewSet(env).Sudo(demoUser.ID()).CreateOrReplace(h.Filter().NewData().
					SetName("c").
					SetResModel("Filter").
					SetUser(demoUser).
					SetIsDefault(true))
				filters := h.Filter().NewSet(env).Sudo(demoUser.ID()).GetFilters("Filter", "")
				So(filters, ShouldHaveLength, 3)
				So(filters[0].Name(), ShouldEqual, "a")
				So(filters[0].User().Equals(demoUser), ShouldBeTrue)
				So(filters[0].IsDefault(), ShouldBeFalse)
				So(filters[0].Domain(), ShouldEqual, "[]")
				So(filters[0].Context(), ShouldEqual, "{}")
				So(filters[0].Sort(), ShouldEqual, "[]")
				So(filters[1].Name(), ShouldEqual, "b")
				So(filters[1].User().Equals(demoUser), ShouldBeTrue)
				So(filters[1].IsDefault(), ShouldBeFalse)
				So(filters[2].Name(), ShouldEqual, "c")
				So(filters[2].User().Equals(demoUser), ShouldBeTrue)
				So(filters[2].IsDefault(), ShouldBeTrue)
			})
			Convey("Update Filter Set Default", func() {
				h.Filter().Create(env, h.Filter().NewData().
					SetName("a").
					SetUser(demoUser).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("b").
					SetUser(demoUser).
					SetResModel("Filter").
					SetIsDefault(true))
				h.Filter().NewSet(env).Sudo(demoUser.ID()).CreateOrReplace(h.Filter().NewData().
					SetName("a").
					SetResModel("Filter").
					SetUser(demoUser).
					SetIsDefault(true))
				filters := h.Filter().NewSet(env).Sudo(demoUser.ID()).GetFilters("Filter", "")
				So(filters, ShouldHaveLength, 2)
				So(filters[0].Name(), ShouldEqual, "a")
				So(filters[0].User().Equals(demoUser), ShouldBeTrue)
				So(filters[0].IsDefault(), ShouldBeTrue)
				So(filters[0].Domain(), ShouldEqual, "[]")
				So(filters[0].Context(), ShouldEqual, "{}")
				So(filters[0].Sort(), ShouldEqual, "[]")
				So(filters[1].Name(), ShouldEqual, "b")
				So(filters[1].User().Equals(demoUser), ShouldBeTrue)
				So(filters[1].IsDefault(), ShouldBeFalse)
			})
		}), ShouldBeNil)
	})
}

func TestGlobalDefaults(t *testing.T) {
	Convey("Global Defaults", t, func() {
		So(models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			demoUser := h.User().Search(env, q.User().HexyaExternalID().Equals("base_user_demo"))
			Convey("New Filter Not Default", func() {
				h.Filter().Create(env, h.Filter().NewData().
					SetName("a").
					SetUser(h.User().NewSet(env)).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("b").
					SetUser(h.User().NewSet(env)).
					SetResModel("Filter"))
				h.Filter().NewSet(env).Sudo(demoUser.ID()).CreateOrReplace(h.Filter().NewData().
					SetName("c").
					SetUser(h.User().NewSet(env)).
					SetResModel("Filter").
					SetIsDefault(true))
				filters := h.Filter().NewSet(env).Sudo(demoUser.ID()).GetFilters("Filter", "")
				So(filters, ShouldHaveLength, 3)
				So(filters[0].Name(), ShouldEqual, "a")
				So(filters[0].User().IsEmpty(), ShouldBeTrue)
				So(filters[0].IsDefault(), ShouldBeFalse)
				So(filters[0].Domain(), ShouldEqual, "[]")
				So(filters[0].Context(), ShouldEqual, "{}")
				So(filters[0].Sort(), ShouldEqual, "[]")
				So(filters[1].Name(), ShouldEqual, "b")
				So(filters[1].User().IsEmpty(), ShouldBeTrue)
				So(filters[1].IsDefault(), ShouldBeFalse)
				So(filters[2].Name(), ShouldEqual, "c")
				So(filters[2].User().IsEmpty(), ShouldBeTrue)
				So(filters[2].IsDefault(), ShouldBeTrue)
			})
			Convey("Update Filter Set Default", func() {
				h.Filter().Create(env, h.Filter().NewData().
					SetName("a").
					SetUser(nil).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("b").
					SetResModel("Filter").
					SetUser(nil).
					SetIsDefault(true))
				So(func() {
					h.Filter().NewSet(env).Sudo(demoUser.ID()).CreateOrReplace(h.Filter().NewData().
						SetName("c").
						SetResModel("Filter").
						SetIsDefault(true))
				}, ShouldPanic)
			})
			Convey("Update Default Filter", func() {
				h.Filter().Create(env, h.Filter().NewData().
					SetName("a").
					SetUser(h.User().NewSet(env)).
					SetResModel("Filter"))
				h.Filter().Create(env, h.Filter().NewData().
					SetName("b").
					SetResModel("Filter").
					SetUser(h.User().NewSet(env)).
					SetIsDefault(true))
				h.Filter().NewSet(env).Sudo(demoUser.ID()).CreateOrReplace(h.Filter().NewData().
					SetName("b").
					SetResModel("Filter").
					SetUser(h.User().NewSet(env)).
					SetContext("{'some_key': True}").
					SetIsDefault(true))
				filters := h.Filter().NewSet(env).Sudo(demoUser.ID()).GetFilters("Filter", "")
				So(filters, ShouldHaveLength, 2)
				So(filters[0].Name(), ShouldEqual, "a")
				So(filters[0].User().IsEmpty(), ShouldBeTrue)
				So(filters[0].IsDefault(), ShouldBeFalse)
				So(filters[0].Domain(), ShouldEqual, "[]")
				So(filters[0].Context(), ShouldEqual, "{}")
				So(filters[0].Sort(), ShouldEqual, "[]")
				So(filters[1].Name(), ShouldEqual, "b")
				So(filters[1].User().IsEmpty(), ShouldBeTrue)
				So(filters[1].IsDefault(), ShouldBeTrue)
				So(filters[1].Domain(), ShouldEqual, "[]")
				So(filters[1].Context(), ShouldEqual, "{'some_key': True}")
				So(filters[1].Sort(), ShouldEqual, "[]")
			})
		}), ShouldBeNil)
	})
}

func TestReadGroup(t *testing.T) {
	Convey("Test Read Group", t, func() {
		So(models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Filter Read Group", func() {
				h.Filter().Create(env, h.Filter().NewData().
					SetName("Filter A").
					SetResModel("Filter"))
				filterB := h.Filter().Create(env, h.Filter().NewData().
					SetName("Filter B").
					SetResModel("Filter"))
				filterB.SetUser(h.User().NewSet(env))

				res := h.Filter().NewSet(env).SearchAll().
					GroupBy(q.Filter().User()).
					Aggregates(q.Filter().Name(), q.Filter().User())
				var oneFalse bool
				for _, r := range res {
					if r.Values().User().IsEmpty() {
						oneFalse = true
						break
					}
				}
				So(oneFalse, ShouldBeTrue)
			})
		}), ShouldBeNil)
	})
}
