// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/server"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
)

// ActionLoad returns the action with the given id
func ActionLoad(c *server.Context) {
	var lang string
	if c.Session().Get("uid") != nil {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			user := h.User().Search(env, q.User().ID().Equals(c.Session().Get("uid").(int64)))
			lang = user.ContextGet().GetString("lang")
		})
	}

	params := struct {
		ActionID          interface{}    `json:"action_id"`
		AdditionalContext *types.Context `json:"additional_context"`
	}{}
	c.BindRPCParams(&params)
	var action actions.Action
	switch actionID := params.ActionID.(type) {
	case string:
		action = *actions.Registry.MustGetByXMLID(actionID)
	case float64:
		action = *actions.Registry.MustGetById(int64(actionID))
	}
	action.Name = action.TranslatedName(lang)
	c.RPC(http.StatusOK, action)
}

// ActionRun runs the given server action
func ActionRun(c *server.Context) {
	params := struct {
		ActionID string         `json:"action_id"`
		Context  *types.Context `json:"context"`
	}{}
	c.BindRPCParams(&params)
	action := actions.Registry.MustGetByXMLID(params.ActionID)

	// Process context ids into args
	var ids []int64
	if params.Context.Get("active_ids") != nil {
		ids = params.Context.Get("active_ids").([]int64)
	} else if params.Context.Get("active_id") != nil {
		ids = []int64{params.Context.Get("active_id").(int64)}
	}
	idsJSON, err := json.Marshal(ids)
	if err != nil {
		log.Panic("Unable to marshal ids")
	}

	// Process context into kwargs
	contextJSON, _ := json.Marshal(params.Context)
	kwargs := make(map[string]json.RawMessage)
	kwargs["context"] = contextJSON

	// Execute the function
	resAction, _ := Execute(c.Session().Get("uid").(int64), CallParams{
		Model:  action.Model,
		Method: action.Method,
		Args:   []json.RawMessage{idsJSON},
		KWArgs: kwargs,
	})

	if _, ok := resAction.(*actions.Action); ok {
		c.RPC(http.StatusOK, resAction)
	} else {
		c.RPC(http.StatusOK, false)
	}
}
