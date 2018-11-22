// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"testing"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/tools/xmlutils"
	"github.com/hexya-erp/pool/h"
	. "github.com/smartystreets/goconvey/convey"
)

var viewDef1 = `
<view id="my_id" name="My View" model="ResUSers">
	<form>
		<group>
			<field name="name" required="1" readonly="1"/>
			<field name="tz" invisible="1"/>
		</group>
	</form>
</view>
`

var viewFieldInfos1 = map[string]*models.FieldInfo{
	"name": {},
	"tz":   {},
}

var viewDef2 = `
<view id="my_id" name="My View" model="ResUSers">
	<form>
		<group>
			<field name="name" attrs='{"readonly": [["Function", "ilike", "manager"]], "required": [["ID", "!=", false]]}'/>
			<field name="tz" invisible="1" attrs='{"invisble": [["Login", "!=", "john"]]}'/>
		</group>
	</form>
</view>
`

var viewFieldInfos2 = map[string]*models.FieldInfo{
	"name": {Required: true},
	"tz":   {ReadOnly: true},
}

func TestViewModifiers(t *testing.T) {
	Convey("Testing correct modifiers injection in views", t, func() {
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("'invisible', 'required' and 'readonly' field attributes should be set in modifiers", func() {
				v, _ := xmlutils.XMLToElement(viewDef1)
				view := h.User().NewSet(env).ProcessView(v, viewFieldInfos1)
				So(view, ShouldEqual, `<view id="my_id" name="My View" model="ResUSers">
	<form>
		<group>
			<field name="name" required="1" readonly="1" modifiers="{&quot;readonly&quot;:true}"/>
			<field name="tz" invisible="1" modifiers="{&quot;invisible&quot;:true}"/>
		</group>
	</form>
</view>`)
			})
			Convey("attrs should be set in modifiers", func() {
				v, _ := xmlutils.XMLToElement(viewDef2)
				view := h.User().NewSet(env).ProcessView(v, viewFieldInfos1)
				So(view, ShouldEqual, `<view id="my_id" name="My View" model="ResUSers">
	<form>
		<group>
			<field name="name" attrs="{&quot;readonly&quot;: [[&quot;Function&quot;, &quot;ilike&quot;, &quot;manager&quot;]], &quot;required&quot;: [[&quot;ID&quot;, &quot;!=&quot;, false]]}" modifiers="{&quot;readonly&quot;:[[&quot;Function&quot;,&quot;ilike&quot;,&quot;manager&quot;]],&quot;required&quot;:[[&quot;ID&quot;,&quot;!=&quot;,false]]}"/>
			<field name="tz" invisible="1" attrs="{&quot;invisble&quot;: [[&quot;Login&quot;, &quot;!=&quot;, &quot;john&quot;]]}" modifiers="{&quot;invisible&quot;:true}"/>
		</group>
	</form>
</view>`)
			})
			Convey("'Readonly' and 'Required' field data should be taken into account", func() {
				v, _ := xmlutils.XMLToElement(viewDef2)
				view := h.User().NewSet(env).ProcessView(v, viewFieldInfos2)
				So(view, ShouldEqual, `<view id="my_id" name="My View" model="ResUSers">
	<form>
		<group>
			<field name="name" attrs="{&quot;readonly&quot;: [[&quot;Function&quot;, &quot;ilike&quot;, &quot;manager&quot;]], &quot;required&quot;: [[&quot;ID&quot;, &quot;!=&quot;, false]]}" modifiers="{&quot;readonly&quot;:[[&quot;Function&quot;,&quot;ilike&quot;,&quot;manager&quot;]],&quot;required&quot;:true}"/>
			<field name="tz" invisible="1" attrs="{&quot;invisble&quot;: [[&quot;Login&quot;, &quot;!=&quot;, &quot;john&quot;]]}" modifiers="{&quot;invisible&quot;:true,&quot;readonly&quot;:true}"/>
		</group>
	</form>
</view>`)
			})
		})
	})
}
