// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package webdata

import (
	"encoding/json"

	"github.com/hexya-addons/web/domains"
	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/operator"
	"github.com/hexya-erp/hexya/src/views"
)

// FieldsViewGetParams is the args struct for the FieldsViewGet function
type FieldsViewGetParams struct {
	ViewID   string `json:"view_id"`
	ViewType string `json:"view_type"`
	Toolbar  bool   `json:"toolbar"`
}

// FieldsViewData is the return type string for the FieldsViewGet function
type FieldsViewData struct {
	Name        string                       `json:"name"`
	Arch        string                       `json:"arch"`
	ViewID      string                       `json:"view_id"`
	Model       string                       `json:"model"`
	Type        views.ViewType               `json:"type"`
	Fields      map[string]*models.FieldInfo `json:"fields"`
	Toolbar     Toolbar                      `json:"toolbar"`
	FieldParent string                       `json:"field_parent"`
}

// SubViewData is the type expected for Views in FieldsViewData
type SubViewData struct {
	Arch   string                       `json:"arch"`
	Fields map[string]*models.FieldInfo `json:"fields"`
}

// SearchParams is the args struct for the SearchRead method
type SearchParams struct {
	Domain domains.Domain `json:"domain"`
	Fields []string       `json:"fields"`
	Offset int            `json:"offset"`
	Limit  interface{}    `json:"limit"`
	Order  string         `json:"order"`
}

// A Toolbar holds the actions in the toolbar of the action manager
type Toolbar struct {
	Print  []*actions.Action `json:"print"`
	Action []*actions.Action `json:"action"`
	Relate []*actions.Action `json:"relate"`
}

// ReadGroupParams is the args struct for the ReadGroup method
type ReadGroupParams struct {
	//domain, fields, groupby, offset=0, limit=None, context=None, orderby=False, lazy=True
	Domain  domains.Domain `json:"domain"`
	Fields  []string       `json:"fields"`
	GroupBy []string       `json:"groupby"`
	Offset  int            `json:"offset"`
	Limit   interface{}    `json:"limit"`
	Order   string         `json:"orderby"`
	Lazy    bool           `json:"lazy"`
}

// NameSearchParams is the args struct for the NameSearch function
type NameSearchParams struct {
	Args     domains.Domain    `json:"args"`
	Name     string            `json:"name"`
	Operator operator.Operator `json:"operator"`
	Limit    interface{}       `json:"limit"`
}

// RecordIDWithName is a tuple with an ID and the display name of a record
type RecordIDWithName struct {
	ID   int64
	Name string
}

// MarshalJSON for RecordIDWithName type
func (rf RecordIDWithName) MarshalJSON() ([]byte, error) {
	arr := [2]interface{}{
		0: rf.ID,
		1: rf.Name,
	}
	res, err := json.Marshal(arr)
	if err != nil {
		return nil, err
	}
	return []byte(res), nil
}

// UnmarshalJSON for RecordIDWithName type
func (rf *RecordIDWithName) UnmarshalJSON(data []byte) error {
	var arr [2]interface{}
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	}
	rf.ID = arr[0].(int64)
	rf.Name = arr[1].(string)
	return nil
}

// SearchReadResult is the result struct for the searchRead function.
type SearchReadResult struct {
	Records []models.RecordData `json:"records"`
	Length  int                 `json:"length"`
}

// LoadViewsArgs is the argument struct for the LoadViews method.
type LoadViewsArgs struct {
	Views   []views.ViewTuple `json:"views"`
	Options LoadViewsOptions  `json:"options"`
}

// LoadViewsOptions are options that can be passed to LoadViews method
type LoadViewsOptions struct {
	Toolbar     bool   `json:"toolbar"`
	LoadFilters bool   `json:"load_filters"`
	ActionID    string `json:"action_id"`
	LoadFields  bool   `json:"load_fields"`
}

// LoadViewsData is the result struct of the LoadViews method
type LoadViewsData struct {
	FieldsViews map[views.ViewType]*FieldsViewData `json:"fields_views"`
	Filters     []models.FieldMap                  `json:"filters"`
	Fields      map[string]*models.FieldInfo       `json:"fields"`
}

// CheckAccessRightsArgs is the args struct for the CheckAccessRights method
type CheckAccessRightsArgs struct {
	Operation      string `json:"operation"`
	RaiseException bool   `json:"raise_exception"`
}
