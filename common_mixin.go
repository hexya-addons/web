// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/hexya-addons/web/controllers"
	"github.com/hexya-addons/web/domains"
	"github.com/hexya-addons/web/webtypes"
	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
	"github.com/hexya-erp/hexya/src/tools/strutils"
	"github.com/hexya-erp/hexya/src/tools/typesutils"
	"github.com/hexya-erp/hexya/src/views"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/m"
	"github.com/hexya-erp/pool/q"
)

// AddNameToRelations returns the given RecordData after getting the name of all 2one relation ids
func commonMixin_AddNameToRelations(rs m.CommonMixinSet, data models.RecordData, fInfos map[string]*models.FieldInfo) models.RecordData {
	for _, fName := range data.Underlying().FieldNames() {
		fi := fInfos[fName.JSON()]
		value := data.Underlying().Get(fName)
		switch v := value.(type) {
		case models.RecordSet:
			relRS := v.Collection().WithEnv(rs.Env())
			switch {
			case fi.Type.Is2OneRelationType():
				if rcId := relRS.Get(models.ID); rcId != int64(0) {
					value = webtypes.RecordIDWithName{
						ID:   rcId.(int64),
						Name: relRS.Call("NameGet").(string),
					}
				} else {
					value = false
				}
			case fi.Type.Is2ManyRelationType():
				if v.Len() > 0 {
					value = v.Ids()
				} else {
					value = []int64{}
				}
			}
		case int64:
			if fi.Type.Is2OneRelationType() {
				if v != 0 {
					rSet := rs.Env().Pool(fi.Relation).Search(rs.Collection().Model().Field(models.ID).Equals(v))
					value = webtypes.RecordIDWithName{
						ID:   v,
						Name: rSet.Call("NameGet").(string),
					}
				} else {
					value = false
				}
			}
		}
		data.Underlying().Set(fName, value)
	}
	return data
}

// FormatRelationFields returns the given data with all relation fields converted to int64 or []int64
func commonMixin_FormatRelationFields(rs m.CommonMixinSet, data models.RecordData, fInfos map[string]*models.FieldInfo) models.RecordData {
	for _, fName := range data.Underlying().FieldNames() {
		fi := fInfos[fName.JSON()]
		value := data.Underlying().Get(fName)
		switch v := value.(type) {
		case models.RecordSet:
			relRS := v.Collection().WithEnv(rs.Env())
			switch {
			case fi.Type.Is2OneRelationType():
				if rcId := relRS.Get(models.ID); rcId != int64(0) {
					value = rcId.(int64)
				} else {
					value = false
				}
			case fi.Type.Is2ManyRelationType():
				if v.Len() > 0 {
					value = v.Ids()
				} else {
					value = []int64{}
				}
			}
		case int64:
			if fi.Type.Is2OneRelationType() {
				if v != 0 {
					value = v
				} else {
					value = false
				}
			}
		}
		data.Underlying().Set(fName, value)
	}
	return data
}

// NameSearch searches for records that have a display name matching the given
// "name" pattern when compared with the given "operator", while also
// matching the optional search domain ("args").
//
// This is used for example to provide suggestions based on a partial
// value for a relational field. Sometimes be seen as the inverse
// function of NameGet but it is not guaranteed to be.
func commonMixin_NameSearch(rs m.CommonMixinSet, params webtypes.NameSearchParams) []webtypes.RecordIDWithName {
	displayNameField := rs.Collection().Model().FieldName("DisplayName")
	searchRs := rs.SearchByName(
		params.Name,
		params.Operator,
		q.CommonMixinCondition{
			Condition: domains.ParseDomain(params.Args, rs.Collection().Model()),
		},
		models.ConvertLimitToInt(params.Limit))
	searchRs.Load(models.ID, displayNameField)

	res := make([]webtypes.RecordIDWithName, searchRs.Len())
	for i, rec := range searchRs.Records() {
		res[i].ID = rec.ID()
		name := rec.Collection().String()
		if _, ok := models.Registry.MustGet(rec.Collection().ModelName()).Fields().Get("DisplayName"); ok {
			name = rec.Get(displayNameField).(string)
		}
		res[i].Name = name
	}
	return res
}

// ProcessWriteValues updates the given data values for Write method to be
// compatible with the ORM, in particular for relation fields
func commonMixin_ProcessWriteValues(rs m.CommonMixinSet, data models.RecordData) models.RecordData {
	fInfos := rs.FieldsGet(models.FieldsGetArgs{})
	for _, f := range data.Underlying().FieldNames() {
		v := data.Underlying().Get(f)
		if _, exists := fInfos[f.JSON()]; !exists {
			log.Panic("Unable to find field", "model", rs.ModelName(), "field", f)
		}
		switch fInfos[f.JSON()].Type {
		case fieldtype.Many2One, fieldtype.One2One:
			if _, isRs := v.(models.RecordSet); isRs {
				continue
			}
			if tuple, ok := v.([]interface{}); ok {
				v = tuple[0]
			}
			id, err := nbutils.CastToInteger(v)
			if err != nil {
				log.Panic("Unable to cast field value", "error", err, "model", rs.ModelName(), "field", f, "value", fInfos[f.JSON()])
			}
			if id == 0 {
				data.Underlying().Set(f, nil)
				continue
			}
			resSet := rs.Env().Pool(fInfos[f.JSON()].Relation).Call("BrowseOne", id).(models.RecordSet).Collection()
			data.Underlying().Set(f, resSet)
		case fieldtype.Many2Many:
			data.Underlying().Set(f, rs.NormalizeM2MData(f, fInfos[f.JSON()], v))
		case fieldtype.One2Many:
			data.Underlying().Set(f, rs.ExecuteO2MActions(f, fInfos[f.JSON()], v))
		case fieldtype.Integer:
			val := reflect.New(fInfos[f.JSON()].GoType).Interface()
			typesutils.Convert(v, val, false)
			data.Underlying().Set(f, reflect.ValueOf(val).Elem().Interface())
		}
	}
	return data
}

// ProcessCreateValues updates the given data values for Create method to be
// compatible with the ORM, in particular for relation fields.
//
// It returns a first FieldMap to be used as argument to the Create method, and
// a second map to be used with a subsequent call to PostProcessCreateValues (for
// updating FKs pointing to the newly created record).
func commonMixin_ProcessCreateValues(rs m.CommonMixinSet, data models.RecordData) (models.RecordData, models.RecordData) {
	createMap := models.NewModelData(data.Underlying().Model)
	deferredMap := models.NewModelData(data.Underlying().Model)
	fInfos := rs.FieldsGet(models.FieldsGetArgs{})
	for _, f := range data.Underlying().FieldNames() {
		v := data.Underlying().Get(f)
		if _, exists := fInfos[f.JSON()]; !exists {
			log.Panic("Unable to find field", "model", rs.ModelName(), "field", f)
		}
		switch fInfos[f.JSON()].Type {
		case fieldtype.Many2One, fieldtype.One2One:
			if vRs, isRs := v.(models.RecordSet); isRs {
				createMap.Set(f, vRs)
				continue
			}
			id, err := nbutils.CastToInteger(v)
			if err != nil {
				log.Panic("Unable to cast field value", "error", err, "model", rs.ModelName(), "field", f, "value", fInfos[f.JSON()])
			}
			if id == 0 {
				createMap.Set(f, nil)
				continue
			}
			resSet := rs.Env().Pool(fInfos[f.JSON()].Relation).Call("BrowseOne", id).(models.RecordSet).Collection()
			createMap.Set(f, resSet)
		case fieldtype.One2Many, fieldtype.Many2Many:
			deferredMap.Set(f, v)
		case fieldtype.Integer:
			val := reflect.New(fInfos[f.JSON()].GoType).Interface()
			typesutils.Convert(v, val, false)
			createMap.Set(f, reflect.ValueOf(val).Elem().Interface())
		default:
			createMap.Set(f, v)
		}
	}
	return createMap, deferredMap
}

// PostProcessCreateValues updates FK of related records created at the same time.
//
// This method is meant to be called with the second returned value of ProcessCreateValues
// after record creation.
func commonMixin_PostProcessCreateValues(rs m.CommonMixinSet, data models.RecordData) {
	if len(data.Underlying().Keys()) == 0 {
		return
	}
	fInfos := rs.FieldsGet(models.FieldsGetArgs{})
	for _, f := range data.Underlying().FieldNames() {
		v := data.Underlying().Get(f)
		if _, exists := fInfos[f.JSON()]; !exists {
			log.Panic("Unable to find field", "model", rs.ModelName(), "field", f)
		}
		switch fInfos[f.JSON()].Type {
		case fieldtype.Many2Many:
			data.Underlying().Set(f, rs.NormalizeM2MData(f, fInfos[f.JSON()], v))
		case fieldtype.One2Many:
			data.Underlying().Set(f, rs.ExecuteO2MActions(f, fInfos[f.JSON()], v))
		}
	}
	rs.Call("Write", data)
}

// ExecuteO2MActions executes the actions on one2many fields given by
// the list of triplets received from the client
func commonMixin_ExecuteO2MActions(rs m.CommonMixinSet, fieldName models.FieldName, info *models.FieldInfo, value interface{}) interface{} {
	switch v := value.(type) {
	case []interface{}:
		relSet := rs.Env().Pool(info.Relation)
		recs := rs.Get(fieldName).(models.RecordSet).Collection()
		if len(v) == 0 {
			return []int64{}
		}
		// We assume we have a list of triplets from client
		for _, triplet := range v {
			action := int(triplet.([]interface{})[0].(float64))
			var values models.RecordData
			switch val := triplet.([]interface{})[2].(type) {
			case bool:
			case map[string]interface{}:
				values = models.NewModelData(relSet.Model(), val)
			case models.FieldMap:
				values = models.NewModelData(relSet.Model(), val)
			}
			switch action {
			case 0:
				// Add reverse FK to point to this RecordSet if this is not the case
				values.Underlying().Set(values.Underlying().Model.FieldName(info.ReverseFK), rs.ID())
				// Create a new record with values
				res := relSet.CallMulti("ProcessCreateValues", values)
				cMap := res[0].(models.RecordData)
				dMap := res[1].(models.RecordData)
				newRec := relSet.Call("Create", cMap).(models.RecordSet).Collection()
				newRec.Call("PostProcessCreateValues", dMap)
				recs = recs.Union(newRec)
			case 1:
				// Update the id record with the given values
				id := int(triplet.([]interface{})[1].(float64))
				rec := relSet.Search(relSet.Model().Field(models.ID).Equals(id))
				values = relSet.Call("ProcessWriteValues", values).(models.RecordData)
				rec.Call("Write", values)
				// add rec to recs in case we are in create
				recs = recs.Union(rec)
			case 2:
				// Remove and delete the id record
				id := int(triplet.([]interface{})[1].(float64))
				rec := relSet.Search(relSet.Model().Field(models.ID).Equals(id))
				recs = recs.Subtract(rec)
				rec.Call("Unlink")
			case 3:
				// Detach the id record
				id := int(triplet.([]interface{})[1].(float64))
				rec := relSet.Search(relSet.Model().Field(models.ID).Equals(id))
				recs = recs.Subtract(rec)
			}
		}
		return recs
	}
	return value
}

// NormalizeM2MData converts the list of triplets received from the client into the final list of ids
// to keep in the Many2Many relationship of this model through the given field.
func commonMixin_NormalizeM2MData(rs m.CommonMixinSet, fieldName models.FieldName, info *models.FieldInfo, value interface{}) interface{} {
	resSet := rs.Env().Pool(info.Relation)
	switch v := value.(type) {
	case []interface{}:
		if len(v) == 0 {
			return resSet
		}
		// We assume we have a list of triplets from client
		for _, triplet := range v {
			// TODO manage effectively multi-tuple input
			action := int(triplet.([]interface{})[0].(float64))
			switch action {
			case 0:
			case 1:
			case 2:
			case 3:
			case 4:
			case 5:
			case 6:
				idList := triplet.([]interface{})[2].([]interface{})
				ids := make([]int64, len(idList))
				for i, id := range idList {
					ids[i] = int64(id.(float64))
				}
				return resSet.Call("Browse", ids).(models.RecordSet).Collection()
			}
		}
	}
	return value
}

// GetFormviewID returns an view id to open the document with.
// This method is meant to be overridden in addons that want
// to give specific view ids for example.
func commonMixin_GetFormviewID(rs m.CommonMixinSet) string {
	return ""
}

// GetFormviewAction returns an action to open the document.
// This method is meant to be overridden in addons that want
// to give specific view ids for example.`,
func commonMixin_GetFormviewAction(rs m.CommonMixinSet) *actions.Action {
	viewID := rs.GetFormviewId()
	return &actions.Action{
		Type:        actions.ActionActWindow,
		Model:       rs.ModelName(),
		ActViewType: actions.ActionViewTypeForm,
		ViewMode:    "form",
		Views:       []views.ViewTuple{{ID: viewID, Type: views.ViewTypeForm}},
		Target:      "current",
		ResID:       rs.ID(),
		Context:     rs.Env().Context(),
	}
}

// FieldsViewGet is the base implementation of the 'FieldsViewGet' method which
// gets the detailed composition of the requested view like fields, mixin,
// view architecture.
func commonMixin_FieldsViewGet(rs m.CommonMixinSet, args webtypes.FieldsViewGetParams) *webtypes.FieldsViewData {
	lang := rs.Env().Context().GetString("lang")
	view := views.Registry.GetByID(args.ViewID)
	if view == nil {
		view = views.Registry.GetFirstViewForModel(rs.ModelName(), views.ViewType(args.ViewType))
	}
	cols := make([]models.FieldName, len(view.Fields))
	for i, f := range view.Fields {
		cols[i] = rs.Collection().Model().FieldName(f)
	}
	fInfos := rs.FieldsGet(models.FieldsGetArgs{Fields: cols})
	arch := rs.ProcessView(view.Arch(lang), fInfos)
	toolbar := rs.GetToolbar()
	res := webtypes.FieldsViewData{
		Name:    view.Name,
		Arch:    arch,
		ViewID:  args.ViewID,
		Model:   view.Model,
		Type:    view.Type,
		Toolbar: toolbar,
		Fields:  fInfos,
	}
	// Sub views
	for field, sViews := range view.SubViews {
		fJSON := rs.Collection().Model().JSONizeFieldName(field)
		relRS := rs.Env().Pool(fInfos[fJSON].Relation)
		if res.Fields[fJSON].Views == nil {
			res.Fields[fJSON].Views = make(map[string]interface{})
		}
		for svType, sv := range sViews {
			sCols := make([]models.FieldName, len(sv.Fields))
			for i, f := range sv.Fields {
				sCols[i] = relRS.Model().FieldName(f)
			}
			svFields := relRS.Call("FieldsGet", models.FieldsGetArgs{Fields: sCols}).(map[string]*models.FieldInfo)
			res.Fields[fJSON].Views[string(svType)] = &webtypes.SubViewData{
				Fields: svFields,
				Arch:   relRS.Call("ProcessView", sv.Arch(lang), svFields).(string),
			}
		}
	}
	return &res
}

// LoadViews returns the data for all the views and filters required in the parameters.
func commonMixin_LoadViews(rs m.CommonMixinSet, args webtypes.LoadViewsArgs) *webtypes.LoadViewsData {
	var res webtypes.LoadViewsData
	res.FieldsViews = make(map[views.ViewType]*webtypes.FieldsViewData)
	for _, viewTuple := range args.Views {
		vType := viewTuple.Type
		if vType == views.ViewTypeList {
			vType = views.ViewTypeTree
		}
		toolbar := args.Options.Toolbar
		if vType == views.ViewTypeSearch {
			toolbar = false
		}
		res.FieldsViews[viewTuple.Type] = rs.FieldsViewGet(webtypes.FieldsViewGetParams{
			Toolbar:  toolbar,
			ViewType: string(vType),
			ViewID:   viewTuple.ID,
		})
	}
	if args.Options.LoadFilters {
		// TODO: doesn't work
		res.Filters = controllers.MethodAdapters["GetFilters"](h.Filter().NewSet(rs.Env()).Collection(),
			"GetFilters",
			[]interface{}{rs.ModelName(), args.Options.ActionID}).([]models.FieldMap)
	}
	res.Fields = rs.FieldsGet(models.FieldsGetArgs{})
	return &res
}

// GetToolbar returns a toolbar populated with the actions linked to this model
func commonMixin_GetToolbar(rs m.CommonMixinSet) webtypes.Toolbar {
	var res webtypes.Toolbar
	for _, a := range actions.Registry.GetActionLinksForModel(rs.ModelName()) {
		switch a.Type {
		case actions.ActionActWindow, actions.ActionServer:
			res.Action = append(res.Action, a)
		}
	}
	return res
}

// ProcessView makes all the necessary modifications to the view
// arch and returns the new xml string.`,
func commonMixin_ProcessView(rs m.CommonMixinSet, arch *etree.Document, fieldInfos map[string]*models.FieldInfo) string {
	// Copy arch into a new document
	doc := arch.Copy()
	// Apply changes
	rs.ManageGroupsOnFields(doc, fieldInfos)
	rs.AddModifiers(doc, fieldInfos)
	// Dump xml to string and return
	res, err := doc.WriteToString()
	if err != nil {
		log.Panic("Unable to render XML", "error", err)
	}
	return res
}

// ManageGroupsOnFields adds the invisible attribute to fields nodes if the current
// user does not belong to one of the groups of the 'groups' attribute
func commonMixin_ManageGroupsOnFields(rs m.CommonMixinSet, doc *etree.Document, fieldInfos map[string]*models.FieldInfo) {
	for _, fieldTag := range doc.FindElements("//field") {
		groupsString := fieldTag.SelectAttrValue("groups", "")
		if groupsString == "" {
			continue
		}
		groups := strings.Split(groupsString, ",")
		var hasGroup bool
		for _, g := range groups {
			group := security.Registry.GetGroup(g)
			if security.Registry.HasMembership(rs.Env().Uid(), group) {
				hasGroup = true
				break
			}
		}
		if !hasGroup {
			fieldTag.CreateAttr("invisible", "1")
		}
	}
}

// AddModifiers adds the modifiers attribute nodes to given xml doc.
func commonMixin_AddModifiers(rs m.CommonMixinSet, doc *etree.Document, fieldInfos map[string]*models.FieldInfo) {
	allModifiers := make(map[*etree.Element]map[string]interface{})
	// Process attrs on all nodes
	for _, attrsTag := range doc.FindElements("[@attrs]") {
		allModifiers[attrsTag] = rs.ProcessElementAttrs(attrsTag, fieldInfos)
	}
	// Process field nodes
	for _, fieldTag := range doc.FindElements("//field") {
		mods, exists := allModifiers[fieldTag]
		if !exists {
			mods = map[string]interface{}{"readonly": false, "required": false, "invisible": false,
				"tree_invisible": false}
		}
		allModifiers[fieldTag] = rs.ProcessFieldElementModifiers(fieldTag, fieldInfos, mods)
	}
	// Set modifier attributes on elements
	for element, modifiers := range allModifiers {
		// Remove false or not applicable keys
		for mod, val := range modifiers {
			v, ok := val.(bool)
			if ok && !v {
				delete(modifiers, mod)
				continue
			}
			viewType := doc.Root().Tag
			toks := strings.Split(mod, "_")
			if len(toks) > 1 && toks[0] != viewType {
				delete(modifiers, mod)
				continue
			}
		}
		// Remove required if field is invisible or readonly
		if req, ok := modifiers["required"].(bool); ok && req {
			inv, ok2 := modifiers["invisible"].(bool)
			ro, ok3 := modifiers["readonly"].(bool)
			if ok2 && inv || ok3 && ro {
				delete(modifiers, "required")
			}
		}

		modJSON, _ := json.Marshal(modifiers)
		element.CreateAttr("modifiers", string(modJSON))
	}
}

// ProcessFieldElementModifiers modifies the given modifiers map by taking into account:
// - 'invisible', 'readonly' and 'required' attributes in field tags
// - 'ReadOnly' and 'Required' parameters of the model's field'
// It returns the modified map.
func commonMixin_ProcessFieldElementModifiers(rs m.CommonMixinSet, element *etree.Element, fieldInfos map[string]*models.FieldInfo, modifiers map[string]interface{}) map[string]interface{} {
	fieldName := element.SelectAttr("name").Value
	// Check if we have the modifier as attribute in the field node
	for modifier := range modifiers {
		modView := modifier
		toks := strings.Split(modifier, "_")
		if len(toks) > 1 {
			modView = toks[1]
		}
		modTag := element.SelectAttrValue(modView, "")
		if modTag == "" {
			continue
		}
		modVal, err := strconv.ParseBool(modTag)
		if modVal || err != nil {
			// If we have an error, we assume it is true
			modifiers[modView] = true
			modifiers[modifier] = true
			continue
		}
		modifiers[modView] = false
		modifiers[modifier] = false
	}
	// Force modifiers if defined in the model
	if fieldInfos[fieldName].ReadOnlyFunc != nil {
		req, cond := fieldInfos[fieldName].ReadOnlyFunc(rs.Env())
		modifiers["readonly"] = req
		if cond != nil {
			modifiers["readonly"] = domains.Domain(cond.Underlying().Serialize()).String()
		}
	}
	if fieldInfos[fieldName].ReadOnly {
		modifiers["readonly"] = true
	}

	if fieldInfos[fieldName].RequiredFunc != nil {
		req, cond := fieldInfos[fieldName].RequiredFunc(rs.Env())
		modifiers["required"] = req
		if cond != nil {
			modifiers["required"] = domains.Domain(cond.Underlying().Serialize()).String()
		}
	}
	if fieldInfos[fieldName].Required {
		modifiers["required"] = true
	}

	if fieldInfos[fieldName].InvisibleFunc != nil {
		req, cond := fieldInfos[fieldName].InvisibleFunc(rs.Env())
		modifiers["invisible"] = req
		if cond != nil {
			modifiers["invisible"] = domains.Domain(cond.Underlying().Serialize()).String()
		}
	}

	return modifiers
}

// ProcessElementAttrs returns a modifiers map according to the domain
// in attrs of the given element
func commonMixin_ProcessElementAttrs(rc *models.RecordCollection, element *etree.Element, fieldInfos map[string]*models.FieldInfo) map[string]interface{} {
	modifiers := map[string]interface{}{"readonly": false, "required": false, "invisible": false}
	attrStr := element.SelectAttrValue("attrs", "")
	if attrStr == "" {
		return modifiers
	}
	var attrs map[string]domains.Domain
	attrStr = strutils.DictToJSON(attrStr)
	err := json.Unmarshal([]byte(attrStr), &attrs)
	if err != nil {
		log.Panic("Invalid attrs definition", "model", rc.ModelName(), "error", err, "attrs", attrStr)
	}
	for modifier := range modifiers {
		cond := domains.ParseDomain(attrs[modifier], rc.Model())
		if cond.IsEmpty() {
			continue
		}
		modifiers[modifier] = attrs[modifier]
	}

	return modifiers
}

// SearchRead retrieves database records according to the filters defined in params.
func commonMixin_SearchRead(rs m.CommonMixinSet, params webtypes.SearchParams) []models.RecordData {
	rSet := rs.AddDomainLimitOffset(params.Domain, models.ConvertLimitToInt(params.Limit), params.Offset, params.Order)
	fields := make([]models.FieldName, len(params.Fields))
	for i, v := range params.Fields {
		fields[i] = rSet.Collection().Model().FieldName(v)
	}
	records := rSet.Read(fields)
	return records
}

// AddDomainLimitOffset adds the given domain, limit, offset
// and order to the current RecordSet query.
func commonMixin_AddDomainLimitOffset(rc *models.RecordCollection, domain domains.Domain, limit int, offset int, order string) *models.RecordCollection {
	rSet := rc
	if searchCond := domains.ParseDomain(domain, rSet.Model()); !searchCond.IsEmpty() {
		rSet = rSet.Call("Search", searchCond).(models.RecordSet).Collection()
	} else {
		rSet = rSet.Call("SearchAll").(models.RecordSet).Collection()
	}
	// Limit
	rSet = rSet.Limit(limit)

	// Offset
	if offset != 0 {
		rSet = rSet.Offset(offset)
	}

	// Order
	if order != "" {
		rSet = rSet.OrderBy(strings.Split(order, ",")...)
	}
	return rSet
}

// PostProcessFilters transforms a map[models.FieldName]models.Conditioner
// in a map[string][]interface{} which acts as a map of domains.
func commonMixin_PostProcessFilters(rs m.CommonMixinSet, in map[models.FieldName]models.Conditioner) map[string][]interface{} {
	res := make(map[string][]interface{})
	for k, v := range in {
		res[k.JSON()] = v.Underlying().Serialize()
	}
	return res
}

// WebReadGroup returns the result of a read_group (and optionally search for and read records inside each
// group), and the total number of groups matching the search domain.
func commonMixin_WebReadGroup(rs m.CommonMixinSet, params webtypes.WebReadGroupParams) webtypes.WebReadGroupResult {
	groups := rs.WebReadGroupPrivate(params)
	limit := models.ConvertLimitToInt(params.Limit)
	var length int
	switch {
	case len(groups) == 0:
	case limit > 0 && len(groups) == limit:
		allGroups := rs.ReadGroup(webtypes.ReadGroupParams{
			Domain:  params.Domain,
			Fields:  []string{"display_name"},
			GroupBy: params.GroupBy,
			Lazy:    true,
		})
		length = len(allGroups)
	default:
		length = len(groups) + params.Offset
	}
	return webtypes.WebReadGroupResult{
		Groups: groups,
		Length: length,
	}
}

// WebReadGroupPrivate performs a read_group and optionally a web_search_read for each group.
func commonMixin_WebReadGroupPrivate(rs m.CommonMixinSet, params webtypes.WebReadGroupParams) []models.FieldMap {
	groups := rs.ReadGroup(webtypes.ReadGroupParams{
		Domain:  params.Domain,
		Fields:  params.Fields,
		GroupBy: params.GroupBy,
		Offset:  params.Offset,
		Limit:   params.Limit,
		Order:   params.Order,
		Lazy:    params.Lazy,
	})
	if params.Expand && len(params.GroupBy) == 1 {
		for i, group := range groups {
			groups[i]["__data"] = rs.WebSearchRead(webtypes.SearchParams{
				Domain: group["__domain"].(domains.Domain),
				Fields: params.Fields,
				Offset: 0,
				Limit:  params.ExpandLimit,
				Order:  params.ExpandOrder,
			})
		}
	}
	return groups
}

// WebSearchRead performs a search_read and a search_count.
func commonMixin_WebSearchRead(rs m.CommonMixinSet, params webtypes.SearchParams) webtypes.SearchReadResult {
	records := rs.SearchRead(params)
	if len(records) == 0 {
		return webtypes.SearchReadResult{}
	}
	limit := models.ConvertLimitToInt(params.Limit)
	length := len(records) + params.Offset
	if limit > 0 && len(records) == limit {
		length = rs.AddDomainLimitOffset(params.Domain, -1, 0, params.Order).SearchCount()
	}
	return webtypes.SearchReadResult{
		Length:  length,
		Records: records,
	}
}

// ReadGroup gets a list of record aggregates according to the given parameters.
func commonMixin_ReadGroup(rs m.CommonMixinSet, params webtypes.ReadGroupParams) []models.FieldMap {
	rSet := rs.AddDomainLimitOffset(params.Domain, models.ConvertLimitToInt(params.Limit), params.Offset, params.Order)
	gb := make(models.FieldNames, len(params.GroupBy))
	for i, v := range params.GroupBy {
		gb[i] = rSet.Collection().Model().FieldName(v)
	}
	effGB := gb
	countFieldPrefix := "_"
	if params.Lazy {
		effGB = gb[:1]
		countFieldPrefix = gb[0].JSON()
	}
	countFieldName := countFieldPrefix + "_count"
	fields := make([]models.FieldName, len(params.Fields))
	for i, v := range params.Fields {
		fields[i] = rSet.Collection().Model().FieldName(v)
	}
	rSet = rSet.GroupBy(effGB[0])
	// We don't want aggregates as CommonMixin Aggregate, so we switch to RecordCollection
	aggregates := rSet.Call("Aggregates", fields).([]models.GroupAggregateRow)
	res := make([]models.FieldMap, len(aggregates))
	fInfos := rSet.FieldsGet(models.FieldsGetArgs{})
	for i, ag := range aggregates {
		line := rs.AddNamesToRelations(ag.Values, fInfos)
		line.Underlying().Set(models.NewFieldName(countFieldName, countFieldName), ag.Count)
		line.Underlying().Set(models.NewFieldName("__domain", "__domain"), ag.Condition.Serialize())
		if len(gb) > 1 {
			line.Underlying().Set(models.NewFieldName("__context", "__context"), models.FieldMap{"group_by": gb[1:].JSON()})
		}
		res[i] = line.Underlying().FieldMap
	}
	return res
}

// SearchDomain execute a search on the given domain.
func commonMixin_SearchDomain(rs m.CommonMixinSet, domain domains.Domain) m.CommonMixinSet {
	cond := q.CommonMixinCondition{
		Condition: domains.ParseDomain(domain, rs.Collection().Model()),
	}
	return rs.Search(cond)
}

// CheckAccessRights verifies that the operation given by "operation" is allowed for
// the current user according to the access rights.
//
// operation must be one of "read", "create", "unlink", "write".
func commonMixin_CheckAccessRights(rs m.CommonMixinSet, args webtypes.CheckAccessRightsArgs) bool {
	switch args.Operation {
	case "read":
		return rs.CheckExecutionPermission(h.CommonMixin().Methods().Read().Underlying(), !args.RaiseException)
	case "write":
		return rs.CheckExecutionPermission(h.CommonMixin().Methods().Write().Underlying(), !args.RaiseException)
	case "unlink":
		return rs.CheckExecutionPermission(h.CommonMixin().Methods().Unlink().Underlying(), !args.RaiseException)
	case "create":
		return rs.CheckExecutionPermission(h.CommonMixin().Methods().Create().Underlying(), !args.RaiseException)
	}
	return false
}

func init() {
	h.CommonMixin().NewMethod("AddNamesToRelations", commonMixin_AddNameToRelations)
	h.CommonMixin().NewMethod("FormatRelationFields", commonMixin_FormatRelationFields)
	h.CommonMixin().NewMethod("NameSearch", commonMixin_NameSearch)
	h.CommonMixin().NewMethod("ProcessWriteValues", commonMixin_ProcessWriteValues)
	h.CommonMixin().NewMethod("ProcessCreateValues", commonMixin_ProcessCreateValues)
	h.CommonMixin().NewMethod("PostProcessCreateValues", commonMixin_PostProcessCreateValues)
	h.CommonMixin().NewMethod("ExecuteO2MActions", commonMixin_ExecuteO2MActions)
	h.CommonMixin().NewMethod("NormalizeM2MData", commonMixin_NormalizeM2MData)
	h.CommonMixin().NewMethod("GetFormviewId", commonMixin_GetFormviewID)
	h.CommonMixin().NewMethod("GetFormviewAction", commonMixin_GetFormviewAction)
	h.CommonMixin().NewMethod("FieldsViewGet", commonMixin_FieldsViewGet)
	h.CommonMixin().NewMethod("LoadViews", commonMixin_LoadViews)
	h.CommonMixin().NewMethod("GetToolbar", commonMixin_GetToolbar)
	h.CommonMixin().NewMethod("ProcessView", commonMixin_ProcessView)
	h.CommonMixin().NewMethod("ManageGroupsOnFields", commonMixin_ManageGroupsOnFields)
	h.CommonMixin().NewMethod("AddModifiers", commonMixin_AddModifiers)
	h.CommonMixin().NewMethod("ProcessFieldElementModifiers", commonMixin_ProcessFieldElementModifiers)
	h.CommonMixin().NewMethod("ProcessElementAttrs", commonMixin_ProcessElementAttrs)
	h.CommonMixin().NewMethod("SearchRead", commonMixin_SearchRead)
	h.CommonMixin().NewMethod("AddDomainLimitOffset", commonMixin_AddDomainLimitOffset)
	h.CommonMixin().NewMethod("PostProcessFilters", commonMixin_PostProcessFilters)
	h.CommonMixin().NewMethod("WebReadGroup", commonMixin_WebReadGroup)
	h.CommonMixin().NewMethod("WebReadGroupPrivate", commonMixin_WebReadGroupPrivate)
	h.CommonMixin().NewMethod("WebSearchRead", commonMixin_WebSearchRead)
	h.CommonMixin().NewMethod("ReadGroup", commonMixin_ReadGroup)
	h.CommonMixin().NewMethod("SearchDomain", commonMixin_SearchDomain)
	h.CommonMixin().NewMethod("CheckAccessRights", commonMixin_CheckAccessRights)
}
