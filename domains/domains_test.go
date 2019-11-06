// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package domains

import (
	"fmt"
	"testing"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/tests"
	_ "github.com/hexya-erp/hexya/src/tests/testllmodule"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMain(m *testing.M) {
	tests.RunTests(m, "domains", nil)
}

func TestDomains(t *testing.T) {
	email := models.NewFieldName("Email", "email")
	isStaff := models.NewFieldName("IsStaff", "is_staff")
	nums := models.NewFieldName("Nums", "nums")
	age := models.NewFieldName("Age", "age")
	profile := models.NewFieldName("Profile", "profile_id")
	Convey("Testing Domains", t, func() {
		_ = models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				userModel := models.Registry.MustGet("User")
				profileModel := models.Registry.MustGet("Profile")
				Convey("Creating users", func() {
					userJohnData := models.NewModelData(userModel).
						Set(models.Name, "John Smith").
						Set(email, "jsmith@example.com").
						Set(isStaff, true).
						Set(nums, 1)
					env.Pool("User").Call("Create", userJohnData)

					userJaneProfileData := models.NewModelData(profileModel).
						Set(age, 24).
						Set(models.NewFieldName("Money", "money"), 12345).
						Set(models.NewFieldName("Street", "street"), "165 5th Avenue").
						Set(models.NewFieldName("City", "city"), "New York").
						Set(models.NewFieldName("Zip", "zip"), "0305").
						Set(models.NewFieldName("Country", "country"), "USA")
					janeProfile := env.Pool("Profile").Call("Create", userJaneProfileData).(models.RecordSet).Collection()

					userJaneData := models.NewModelData(userModel).
						Set(models.Name, "Jane Smith").
						Set(email, "jane.smith@example.com").
						Set(profile, janeProfile).
						Set(nums, 2)
					env.Pool("User").Call("Create", userJaneData)

					userWillData := models.NewModelData(userModel).
						Set(models.Name, "Will Smith").
						Set(email, "will.smith@example.com").
						Set(isStaff, true).
						Set(nums, 3)

					env.Pool("User").Call("Create", userWillData)

					martinProfile := env.Pool("Profile").Call("Create",
						models.NewModelData(profileModel).Set(age, 45))
					userData := models.NewModelData(userModel).
						Set(models.Name, "Martin Weston").
						Set(email, "mweston@example.com").
						Set(profile, martinProfile)
					user := env.Pool("User").Call("Create", userData).(models.RecordSet).Collection()
					So(user.Get(profile).(models.RecordSet).Collection().Get(age), ShouldEqual, 45)
				})
				Convey("Testing simple [(A), (B)] domain", func() {
					dom1 := []interface{}{
						0: []interface{}{"Name", "like", "Smith"},
						1: []interface{}{"Age", "=", 24},
					}
					cond := ParseDomain(dom1, userModel)
					So(fmt.Sprintf("%v", cond.Serialize()), ShouldEqual, "[& [age = 24] [name like Smith]]")
					dom1Users := env.Pool("User").Search(cond)
					So(dom1Users.Len(), ShouldEqual, 1)
					So(dom1Users.Get(models.Name), ShouldEqual, "Jane Smith")
				})
				Convey("Testing ['|', (A), (B)] domain", func() {
					dom2 := []interface{}{
						0: "|",
						1: []interface{}{"Name", "like", "Will"},
						2: []interface{}{"Email", "ilike", "Jane.Smith"},
					}
					cond := ParseDomain(dom2, userModel)
					So(fmt.Sprintf("%v", cond.Serialize()), ShouldEqual, "[| [name like Will] [email ilike Jane.Smith]]")
					dom2Users := env.Pool("User").Search(cond).OrderBy("Name")
					So(dom2Users.Len(), ShouldEqual, 2)
					userRecs := dom2Users.Records()
					So(userRecs[0].Get(models.Name), ShouldEqual, "Jane Smith")
					So(userRecs[1].Get(models.Name), ShouldEqual, "Will Smith")
				})
				Convey("Testing ['|', (A), '&' , (B), (C), (D)] domain", func() {
					dom3 := []interface{}{
						0: "|",
						1: []interface{}{"Name", "like", "Will"},
						2: "&",
						3: []interface{}{"Age", ">", 0},
						4: []interface{}{"Age", "<", 25},
						5: []interface{}{"Email", "not like", "will.smith"},
					}
					cond := ParseDomain(dom3, userModel)
					So(fmt.Sprintf("%v", cond.Serialize()), ShouldEqual, "[& | & [age < 25] [age > 0] [name like Will] [email not like will.smith]]")
					dom3Users := env.Pool("User").Search(cond).OrderBy("Name")
					So(dom3Users.Len(), ShouldEqual, 1)
					So(dom3Users.Get(models.Name), ShouldEqual, "Jane Smith")
				})
				Convey("Testing ['|', '|', (A), (B), (C)] domain", func() {
					dom4 := []interface{}{
						0: "|",
						1: "|",
						2: []interface{}{"Name", "ilike", "john"},
						3: []interface{}{"Name", "ilike", "jane"},
						4: []interface{}{"Name", "ilike", "will"},
					}
					cond := ParseDomain(dom4, userModel)
					So(fmt.Sprintf("%v", cond.Serialize()), ShouldEqual, "[| | [name ilike john] [name ilike jane] [name ilike will]]")
					dom1Users := env.Pool("User").Search(cond)
					So(dom1Users.Len(), ShouldEqual, 3)
				})
				Convey("Testing ParseString", func() {
					dom5Str := `[('Name', "ilike", 'john')]`
					dom5 := Domain{
						0: []interface{}{"Name", "ilike", "john"},
					}
					dom6Str := `[("Val", "<", 123.5), ("Val", ">", -10)]`
					dom6 := Domain{
						0: []interface{}{"Val", "<", 123.5},
						1: []interface{}{"Val", ">", -10},
					}
					dom7Str := `['|', ("Name", "ilike", 'john'), '&', ["Val", "<", 123.5], ("Val", ">", -10)]`
					dom7 := Domain{
						0: "|",
						1: []interface{}{"Name", "ilike", "john"},
						2: "&",
						3: []interface{}{"Val", "<", 123.5},
						4: []interface{}{"Val", ">", -10},
					}
					So(ParseString(dom5Str).String(), ShouldEqual, dom5.String())
					So(ParseString(dom6Str).String(), ShouldEqual, dom6.String())
					So(ParseString(dom7Str).String(), ShouldEqual, dom7.String())
				})
			})
		})
	})
}
