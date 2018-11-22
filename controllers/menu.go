// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"

	"github.com/hexya-erp/hexya/src/server"
)

// MenuLoadNeedaction serves the number of objects that need an action
// in the given menu IDs
func MenuLoadNeedaction(c *server.Context) {
	type lnaParams struct {
		MenuIds []string `json:"menu_ids"`
	}
	var params lnaParams
	c.BindRPCParams(&params)

	// TODO: update with real needaction support
	type lnaResponse struct {
		NeedactionEnabled bool `json:"needaction_enabled"`
		NeedactionCounter int  `json:"needaction_counter"`
	}
	res := make(map[string]lnaResponse)
	for _, menu := range params.MenuIds {
		res[menu] = lnaResponse{}
	}
	c.RPC(http.StatusOK, res)
}
