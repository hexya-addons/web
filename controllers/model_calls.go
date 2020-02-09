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

package controllers

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hexya-addons/web/domains"
	"github.com/hexya-addons/web/odooproxy"
	"github.com/hexya-addons/web/webtypes"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/tools/logging"
)

var (
	log logging.Logger
	// TypeSubstitutions maps interface types to appropriate concrete types to unmarshal into.
	//
	// All types that implement the interface type will use the given concrete type.
	TypeSubstitutions = map[reflect.Type]reflect.Type{
		reflect.TypeOf((*models.RecordData)(nil)).Elem(): reflect.TypeOf(models.FieldMap{}),
	}
	// TypePostProcess maps interface types to a function to apply after unmarshalling.
	TypePostProcess = map[reflect.Type]interface{}{
		reflect.TypeOf((*models.RecordData)(nil)).Elem(): func(rs models.RecordSet, arg models.FieldMap) interface{} {
			return models.NewModelData(rs.Collection().Model(), arg).Wrap()
		},
	}
)

// CallParams is the arguments' struct for the Execute function.
// It defines a method to call on a model with the given args and keyword args.
type CallParams struct {
	Model  string                     `json:"model"`
	Method string                     `json:"method"`
	Args   []json.RawMessage          `json:"args"`
	KWArgs map[string]json.RawMessage `json:"kwargs"`
}

// Execute executes a method on an object
func Execute(uid int64, params CallParams) (res interface{}, rError error) {
	checkUser(uid)

	// Create new Environment with new transaction
	rError = models.ExecuteInNewEnvironment(uid, func(env models.Environment) {

		// Create RecordSet from Environment
		rs, parms, _ := createRecordCollection(env, params)
		ctx := extractContext(params)
		rs = rs.WithNewContext(&ctx)

		methodName := odooproxy.ConvertMethodName(params.Method)

		// Parse Args and KWArgs using the following logic:
		// - If 2nd argument of the function is a struct, then:
		//     * Parse remaining Args in the struct fields
		//     * Parse KWArgs in the struct fields, possibly overwriting Args
		// - Else:
		//     * Parse Args as the function args
		//     * Ignore KWArgs
		var fnArgs []interface{}
		if rs.MethodType(methodName).NumIn() > 1 {
			fnSecondArgType := rs.MethodType(methodName).In(1)
			if fnSecondArgType.Kind() == reflect.Struct {
				// 2nd argument is a struct,
				fnArgs = make([]interface{}, 1)
				argStructValue := reflect.New(fnSecondArgType).Elem()
				putParamsValuesInStruct(&argStructValue, rs, parms)
				putKWValuesInStruct(&argStructValue, rs, params.KWArgs)
				fnArgs[0] = argStructValue.Interface()
			} else {
				// Second argument is not a struct, so we parse directly in the function args
				fnArgs = make([]interface{}, len(parms))
				err := putParamsValuesInArgs(&fnArgs, rs, methodName, parms)
				if err != nil {
					log.Panic(err.Error(), "method", methodName, "args", parms)
				}
			}
		}

		adapter, ok := MethodAdapters[methodName]
		if ok {
			res = adapter(rs, methodName, fnArgs)
		} else {
			res = rs.Call(methodName, fnArgs...)
		}

		// resVal := reflect.ValueOf(res)
		// if single && resVal.Kind() == reflect.Slice {
		// 	// Return only the first element of the slice if called with only one id.
		// 	newRes := reflect.New(resVal.Type().Elem()).Elem()
		// 	if resVal.Len() > 0 {
		// 		newRes.Set(resVal.Index(0))
		// 	}
		// 	res = newRes.Interface()
		// }
		res = convertReturnedValue(rs, res)
	})

	return
}

// putParamsValuesInStruct decodes parms and sets the fields of the structValue
// with the values of parms, in order.
func putParamsValuesInStruct(structValue *reflect.Value, rs models.RecordSet, parms []json.RawMessage) {
	argStructValue := *structValue
	for i := 0; i < argStructValue.NumField(); i++ {
		if len(parms) <= i {
			// We have less arguments than the size of the struct
			break
		}
		argsValue := reflect.ValueOf(parms[i])
		fieldPtrValue := reflect.New(argStructValue.Type().Field(i).Type)
		if err := unmarshalJSONValue(argsValue, fieldPtrValue, rs); err != nil {
			// We deliberately continue here to have default value if there is an error
			// This is to manage cases where the given data type is inconsistent (such
			// false instead of [] or object{}).
			log.Debug("Unable to unmarshal argument", "param", string(parms[i]), "error", err)
			continue
		}
		argStructValue.Field(i).Set(fieldPtrValue.Elem())
	}
}

// putKWValuesInStruct decodes kwArgs and sets the fields of the structValue
// with the values of kwArgs, mapping each field with its entry in kwArgs.
func putKWValuesInStruct(structValue *reflect.Value, rs models.RecordSet, kwArgs map[string]json.RawMessage) {
	argStructValue := *structValue
	for k, v := range kwArgs {
		field := getStructFieldByJSONTag(argStructValue, k)
		if field.IsValid() {
			dest := reflect.New(field.Elem().Type())
			if err := unmarshalJSONValue(reflect.ValueOf(v), dest, rs); err != nil {
				// We deliberately continue here to have default value if there is an error
				// This is to manage cases where the given data type is inconsistent (such
				// false instead of [] or object{}).
				log.Debug("Unable to unmarshal argument", "param", string(v), "error", err)
				continue
			}
			field.Elem().Set(dest.Elem())
		}
	}
}

// putParamsValuesInArgs decodes parms and sets fnArgs with the types of methodType arguments
// and the values of parms, in order.
func putParamsValuesInArgs(fnArgs *[]interface{}, rs models.RecordSet, methodName string, parms []json.RawMessage) error {
	methodType := rs.Collection().MethodType(methodName)
	numArgs := methodType.NumIn() - 1
	for i := 0; i < len(parms); i++ {
		if (methodType.IsVariadic() && len(parms) < numArgs-1) ||
			(!methodType.IsVariadic() && len(parms) < numArgs) {
			// We have less arguments than the arguments of the method
			return fmt.Errorf("wrong number of args in non-struct function args (%d instead of %d)", len(parms), numArgs)
		}
		methInType := methodType.In(i + 1)
		argsValue := reflect.ValueOf(parms[i])
		resValue := reflect.New(methInType)
		if err := unmarshalJSONValue(argsValue, resValue, rs); err != nil {
			// Same remark as above
			log.Debug("Unable to unmarshal argument", "param", string(parms[i]), "error", err)
			continue
		}
		res := resValue.Elem().Interface()
		(*fnArgs)[i] = res
	}
	return nil
}

// createRecordCollection creates a RecordCollection instance from the given environment, based
// on the given params. If the first argument given in params can be parsed as an id or a slice
// of ids, then it is used to populate the RecordCollection. Otherwise, it returns an empty
// RecordCollection. This function also returns the remaining arguments after id(s) have been
// parsed, and a boolean value set to true if the RecordSet has only one ID.
func createRecordCollection(env models.Environment, params CallParams) (*models.RecordCollection, []json.RawMessage, bool) {
	modelName := odooproxy.ConvertModelName(params.Model)
	rc := env.Pool(modelName)

	// Try to parse the first argument of Args as id or ids.
	var (
		single    bool
		idsParsed bool
		id        int64
		ids       []int64
		rin       webtypes.RecordIDWithName
		rins      []webtypes.RecordIDWithName
	)
	if len(params.Args) > 0 {
		switch {
		case json.Unmarshal(params.Args[0], &rins) == nil:
			for _, r := range rins {
				ids = append(ids, r.ID)
			}
			fallthrough
		case json.Unmarshal(params.Args[0], &ids) == nil:
			rc = rc.Search(rc.Model().Field(models.ID).In(ids))
			idsParsed = true
		case json.Unmarshal(params.Args[0], &rin) == nil:
			id = rin.ID
			fallthrough
		case json.Unmarshal(params.Args[0], &id) == nil:
			rc = rc.Search(rc.Model().Field(models.ID).Equals(id))
			single = true
			idsParsed = true
		}
	}

	remainingParams := params.Args
	if idsParsed {
		// We remove ids already parsed from args
		remainingParams = remainingParams[1:]
	}
	return rc, remainingParams, single
}

// extractContext extracts the context from the given params and returns it.
func extractContext(params CallParams) types.Context {
	ctxStr, ok := params.KWArgs["context"]
	var ctx types.Context
	if ok {
		if err := json.Unmarshal(ctxStr, &ctx); err != nil {
			log.Panic("Unable to JSON unmarshal context", "context_string", string(ctxStr), "error", err)
		}
	}
	return ctx
}

// checkUser panics if the given uid is 0 (i.e. no user is logged in).
func checkUser(uid int64) {
	if uid == 0 {
		log.Panic("User must be logged in to call model method")
	}
}

// getFieldValue retrieves the given field of the given model and id.
func getFieldValue(uid, id int64, model, field string) (res interface{}, rError error) {
	checkUser(uid)
	rError = models.ExecuteInNewEnvironment(uid, func(env models.Environment) {
		model = odooproxy.ConvertModelName(model)
		rc := env.Pool(model)
		res = rc.Search(rc.Model().Field(models.ID).Equals(id)).Get(rc.Model().FieldName(field))
	})

	return
}

// convertReturnedValue converts the values returned by the ORM
// in a suitable JSON format for the client
func convertReturnedValue(rs models.RecordSet, res interface{}) interface{} {
	switch val := res.(type) {
	case models.RecordSet:
		// Return ID(s) if res is a *RecordSet
		switch {
		case val.IsEmpty():
			res = []int64{}
		case val.Len() == 1:
			res = val.Ids()[0]
		default:
			res = val.Ids()
		}
	case models.RecordData:
		fInfos := rs.Call("FieldsGet", models.FieldsGetArgs{})
		res = rs.Call("FormatRelationFields", res, fInfos).(models.RecordData)
	}
	return res
}

// SearchReadParams is the args struct for the searchRead function.
type SearchReadParams struct {
	Context types.Context  `json:"context"`
	Domain  domains.Domain `json:"domain"`
	Fields  []string       `json:"fields"`
	Limit   interface{}    `json:"limit"`
	Model   string         `json:"model"`
	Offset  int            `json:"offset"`
	Sort    string         `json:"sort"`
}

// SearchRead retrieves database records according to the filters defined in params.
func SearchRead(uid int64, params SearchReadParams) (res *webtypes.SearchReadResult, rError error) {
	checkUser(uid)
	rError = models.ExecuteInNewEnvironment(uid, func(env models.Environment) {
		model := odooproxy.ConvertModelName(params.Model)
		rs := env.Pool(model).WithNewContext(&params.Context)
		srp := webtypes.SearchParams{
			Domain: params.Domain,
			Fields: params.Fields,
			Offset: params.Offset,
			Limit:  params.Limit,
			Order:  params.Sort,
		}
		data := searchReadAdapter(rs, "SearchRead", []interface{}{srp}).([]models.RecordData)
		length := rs.Call("AddDomainLimitOffset", srp.Domain, 0, srp.Offset, srp.Order).(models.RecordSet).Collection().SearchCount()
		res = &webtypes.SearchReadResult{
			Records: data,
			Length:  length,
		}
	})
	return
}

// unmarshalJSONValue unmarshals the given data as a Value of type []byte into
// the dst Value. dst must be a pointer Value.
//
// If dst is an interface, its is passed through TypesSubstitutions are applied.
func unmarshalJSONValue(data, dst reflect.Value, rs models.RecordSet) error {
	if dst.Type().Kind() != reflect.Ptr {
		log.Panic("dst must be a pointer value", "data", data, "dst", dst)
	}
	dstType := dst.Type().Elem()
	var mapType reflect.Type
	for intType, destType := range TypeSubstitutions {
		if dstType.Implements(intType) {
			dstType = destType
			mapType = intType
			break
		}
	}
	dest := reflect.New(dstType)
	res := reflect.ValueOf(json.Unmarshal).Call([]reflect.Value{data, dest})[0]
	if res.Interface() != nil {
		return res.Interface().(error)
	}

	if f, ok := TypePostProcess[mapType]; ok {
		fnct := reflect.ValueOf(f)
		ppVal := fnct.Call([]reflect.Value{reflect.ValueOf(rs), dest.Elem()})
		dest = ppVal[0]
	}
	dst.Elem().Set(dest.Elem())
	return nil
}

// getStructFieldByJSONTag returns a pointer value of the struct field of the given structValue
// with the given JSON tag. If several are found, it returns the first one. If none are
// found it returns the zero value. If structType is not a Type of Kind struct, then it panics.
func getStructFieldByJSONTag(structValue reflect.Value, tag string) (sf reflect.Value) {
	for i := 0; i < structValue.NumField(); i++ {
		sField := structValue.Field(i)
		sfTag := structValue.Type().Field(i).Tag.Get("json")
		if sfTag == tag {
			sf = sField.Addr()
			return
		}
	}
	return
}

func init() {
	log = logging.GetLogger("web")
}
