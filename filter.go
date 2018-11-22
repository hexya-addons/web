// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"fmt"
	"strings"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/tools/strutils"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
)

func init() {
	filterModel := h.Filter().DeclareModel()
	filterModel.AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{String: "Filter Name", Required: true},
		"User": models.Many2OneField{RelationModel: h.User(), OnDelete: models.Cascade,
			Default: func(env models.Environment) interface{} {
				return h.User().Search(env, q.User().ID().Equals(env.Uid()))
			}, Help: `The user this filter is private to. When left empty the filter is public and available to all users.`},
		"Domain":    models.TextField{Required: true, Default: models.DefaultValue("[]")},
		"Context":   models.TextField{Required: true, Default: models.DefaultValue("{}")},
		"Sort":      models.TextField{Required: true, Default: models.DefaultValue("[]")},
		"ResModel":  models.CharField{String: "Model", Required: true, JSON: "model_id"},
		"IsDefault": models.BooleanField{String: "Default filter"},
		"Action": models.CharField{
			Help: `The menu action this filter applies to. When left empty the filter applies to all menus for this model.`,
			JSON: "action_id"},
		"Active": models.BooleanField{Default: models.DefaultValue(true), Required: true},
	})

	filterModel.AddSQLConstraint("name_model_uid_unique", "unique (name, model_id, user_id, action_id)", "Filter names must be unique")

	filterModel.Methods().GetFilters().DeclareMethod(
		`GetFilters returns the filters for the given model and actionID for the current user`,
		func(rs h.FilterSet, modelName, actionID string) []*h.FilterData {
			actionCondition := rs.GetActionCondition(actionID)
			filters := h.Filter().Search(rs.Env(), q.Filter().ResModel().Equals(modelName).
				AndCond(actionCondition).
				AndCond(q.Filter().UserFilteredOn(q.User().ID().Equals(rs.Env().Uid())).
					Or().User().IsNull()))
			userContext := h.User().Browse(rs.Env(), []int64{rs.Env().Uid()}).ContextGet()
			res := filters.WithNewContext(userContext).All()
			return res
		})

	filterModel.Methods().Copy().Extend("",
		func(rs h.FilterSet, overrides *h.FilterData, fieldsToUnset ...models.FieldNamer) h.FilterSet {
			rs.EnsureOne()
			vals, fieldsToUnset := rs.DataStruct(overrides.FieldMap(fieldsToUnset...))
			vals.Name = fmt.Sprintf("%s (copy)", rs.Name())
			return rs.Super().Copy(overrides, fieldsToUnset...)
		})

	filterModel.Methods().CreateOrReplace().DeclareMethod(
		`CreateOrReplace creates or updates the filter with the given parameters.
		Filter is considered the same if it has the same name (case insensitive) and the same user (if it has one).`,
		func(rs h.FilterSet, vals models.FieldMapper) h.FilterSet {
			fMap := vals.FieldMap()
			if fDomain, exists := fMap["domain"]; exists {
				fMap["domain"] = strutils.MarshalToJSONString(fDomain)
				fMap["domain"] = strings.Replace(fDomain.(string), "false", "False", -1)
				fMap["domain"] = strings.Replace(fDomain.(string), "true", "True", -1)
			}
			if fContext, exists := fMap["context"]; exists {
				fMap["context"] = strutils.MarshalToJSONString(fContext)
			}
			values, _ := rs.DataStruct(fMap)
			currentFilters := rs.GetFilters(values.ResModel, values.Action)
			var matchingFilters []*h.FilterData
			for _, filter := range currentFilters {
				if strings.ToLower(filter.Name) != strings.ToLower(values.Name) {
					continue
				}
				if !filter.User.Equals(values.User) {
					continue
				}
				matchingFilters = append(matchingFilters, filter)
			}

			if values.IsDefault {
				if !values.User.IsEmpty() {
					// Setting new default: any other default that belongs to the user
					// should be turned off
					actionCondition := rs.GetActionCondition(values.Action)
					defaults := h.Filter().Search(rs.Env(), actionCondition.
						And().ResModel().Equals(values.ResModel).
						And().User().Equals(values.User).
						And().IsDefault().Equals(true))
					if !defaults.IsEmpty() {
						defaults.SetIsDefault(false)
					}
				} else {
					rs.CheckGlobalDefault(values, matchingFilters)
				}
			}
			if len(matchingFilters) > 0 {
				// When a filter exists for the same (name, model, user) triple, we simply
				// replace its definition (considering action_id irrelevant here)
				matchingFilter := h.Filter().Browse(rs.Env(), []int64{matchingFilters[0].ID})
				matchingFilter.Write(values)
				return matchingFilter
			}
			return rs.Create(values)
		})

	filterModel.Methods().CheckGlobalDefault().DeclareMethod(
		`CheckGlobalDefault checks if there is a global default for the ResModel requested.

	       If there is, and the default is different than the record being written
	       (-> we're not updating the current global default), raise an error
	       to avoid users unknowingly overwriting existing global defaults (they
	       have to explicitly remove the current default before setting a new one)

	       This method should only be called if 'vals' is trying to set 'IsDefault'`,
		func(rs h.FilterSet, vals *h.FilterData, matchingFilters []*h.FilterData) {
			values, _ := rs.DataStruct(vals.FieldMap())
			actionCondition := rs.GetActionCondition(values.Action)
			defaults := h.Filter().Search(rs.Env(), actionCondition.
				And().ResModel().Equals(values.ResModel).
				And().User().IsNull().
				And().IsDefault().Equals(true))
			if defaults.IsEmpty() {
				return
			}
			if len(matchingFilters) > 0 && matchingFilters[0].ID == defaults.ID() {
				return
			}
			log.Panic("There is already a shared filter set as default for this model, delete or change it before setting a new default", "model", values.ResModel)
		})

	filterModel.Methods().GetActionCondition().DeclareMethod(
		`GetActionCondition returns a condition for matching filters that are visible in the
		same context (menu/view) as the given action.`,
		func(rs h.FilterSet, action string) q.FilterCondition {
			if action != "" {
				// filters specific to this menu + global ones
				return q.Filter().Action().Equals(action).Or().Action().IsNull()
			}
			return q.Filter().Action().IsNull()
		})
}
