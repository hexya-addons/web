// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"fmt"

	"github.com/hexya-addons/web/domains"
	"github.com/hexya-addons/web/odooproxy"
	"github.com/hexya-addons/web/webtypes"
	"github.com/hexya-erp/hexya/src/models"
)

// MethodAdapters is a map giving the adapter to call for each method
var MethodAdapters = map[string]methodAdapter{
	"Create":     createAdapter,
	"Write":      writeAdapter,
	"Onchange":   onchangeAdapter,
	"Read":       readAdapter,
	"SearchRead": searchReadAdapter,
	"FieldsGet":  fieldsGetAdapter,
	"NameGet":    nameGetAdapter,
}

// A methodAdapter can modify calls made by the odoo client
// to match the expected arguments of the Hexya ORM. Similarly
// it can modify the returned values so that they are understood by the client
type methodAdapter func(*models.RecordCollection, string, []interface{}) interface{}

// checkMethods panics if the given method is different from expected or if args does not have a length of numArgs.
func checkMethod(method, expected string, args []interface{}, numArgs int) {
	if odooproxy.ConvertMethodName(method) != expected {
		log.Panic(fmt.Sprintf("%s adapter should only be called on %s methods", expected, expected), "method", method, "args", args)
	}
	if len(args) != numArgs {
		log.Panic("Wrong number of arguments for method adapter", "method", method, "expected", numArgs, "args", args)
	}
}

// createAdapter adapts json object received from client to Create's FieldMap argument.
func createAdapter(rc *models.RecordCollection, method string, args []interface{}) interface{} {
	checkMethod(method, "Create", args, 1)
	data, ok := args[0].(models.RecordData)
	if !ok {
		log.Panic("Expected arg for Create method to be RecordData", "argType", fmt.Sprintf("%T", args[0]))
	}
	pcv := rc.CallMulti("ProcessCreateValues", data)
	cMap := pcv[0].(models.RecordData)
	dMap := pcv[1].(models.RecordData)
	res := rc.WithContext("hexya_skip_check_constraints", true).Call("Create", cMap).(models.RecordSet).Collection()
	res.Call("PostProcessCreateValues", dMap)
	res.WithContext("hexya_skip_check_constraints", false).CheckConstraints(rc.Model().FieldNames())
	return res
}

// writeAdapter adapts json object received from client to Write's FieldMap and []FieldNamer argument.
func writeAdapter(rc *models.RecordCollection, method string, args []interface{}) interface{} {
	checkMethod(method, "Write", args, 1)
	data, ok := args[0].(models.RecordData)
	if !ok {
		log.Panic("Expected arg for Write method to be models.FieldMap", "argType", fmt.Sprintf("%T", args[0]))
	}
	data = rc.Call("ProcessWriteValues", data).(models.RecordData)
	res := rc.Call("Write", data)
	return res
}

// onchangeAdapter adapts json object received from client and add names to relation to the result.
func onchangeAdapter(rc *models.RecordCollection, method string, args []interface{}) interface{} {
	checkMethod(method, "Onchange", args, 1)
	params, ok := args[0].(models.OnchangeParams)
	if !ok {
		log.Panic("Expected arg for Onchange method to be OnchangeParams", "argType", fmt.Sprintf("%T", args[0]))
	}
	params.Values = rc.Call("ProcessWriteValues", params.Values).(models.RecordData)
	mRes := rc.Call("Onchange", params).(models.OnchangeResult)
	fInfos := rc.Call("FieldsGet", models.FieldsGetArgs{})
	var res webtypes.OnChangeResult
	res.Value = rc.Call("AddNamesToRelations", mRes.Value, fInfos).(models.RecordData)
	res.Filters = rc.Call("PostProcessFilters", mRes.Filters).(map[string][]interface{})
	res.Warning = mRes.Warning
	return res

}

// readAdapter add names to relation of the result.
func readAdapter(rc *models.RecordCollection, method string, args []interface{}) interface{} {
	checkMethod(method, "Read", args, 1)
	params, ok := args[0].(models.FieldNames)
	if !ok {
		log.Panic("Expected arg for Read method to be models.FieldNames", "argType", fmt.Sprintf("%T", args[0]))
	}
	res := rc.Call("Read", params).([]models.RecordData)
	for i, data := range res {
		// Getting rec, which is this RecordSet but with its real type (not CommonMixinSet)
		id := data.Underlying().Get(models.ID).(int64)
		rec := rc.Env().Pool(rc.ModelName()).Search(rc.Model().Field(models.ID).Equals(id))
		fInfos := rec.Call("FieldsGet", models.FieldsGetArgs{})
		res[i] = rec.Call("AddNamesToRelations", data, fInfos).(models.RecordData)
	}
	return res
}

// searchReadAdapter add names to relation of the result.
func searchReadAdapter(rc *models.RecordCollection, method string, args []interface{}) interface{} {
	checkMethod(method, "SearchRead", args, 1)
	params, ok := args[0].(webtypes.SearchParams)
	if !ok {
		log.Panic("Expected arg for SearchRead method to be webdata.SearchParams", "argType", fmt.Sprintf("%T", args[0]))
	}
	res := rc.Call("SearchRead", params).([]models.RecordData)
	for i, data := range res {
		// Getting rec, which is this RecordSet but with its real type (not CommonMixinSet)
		id := data.Underlying().Get(models.ID).(int64)
		rec := rc.Env().Pool(rc.ModelName()).Search(rc.Model().Field(models.ID).Equals(id))
		fInfos := rec.Call("FieldsGet", models.FieldsGetArgs{})
		res[i] = rec.Call("AddNamesToRelations", data, fInfos).(models.RecordData)
	}
	if res == nil {
		res = []models.RecordData{}
	}
	return res
}

// fieldsGetAdapter stringifies the domain of each field in the returned value.
func fieldsGetAdapter(rc *models.RecordCollection, method string, args []interface{}) interface{} {
	checkMethod(method, "FieldsGet", args, 1)
	params, ok := args[0].(models.FieldsGetArgs)
	if !ok {
		log.Panic("Expected arg for FieldsGet method to be FieldsGetArgs", "args", args)
	}
	res := rc.Call("FieldsGet", params).(map[string]*models.FieldInfo)
	for f, fInfo := range res {
		dom, _ := fInfo.Domain.([]interface{})
		res[f].Domain = domains.Domain(dom).String()
	}
	return res
}

// nameGetAdapter handles calls with multiple ids.
func nameGetAdapter(rc *models.RecordCollection, method string, args []interface{}) interface{} {
	checkMethod(method, "NameGet", args, 0)
	// We make the slice to be sure not to have nil returned
	res := make([][2]interface{}, 0)
	for _, rec := range rc.Records() {
		res = append(res, [2]interface{}{
			rec.Ids()[0],
			rec.Call("NameGet").(string),
		})
	}
	return res
}
