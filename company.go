// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/pool/h"
)

func init() {
	h.Company().AddFields(map[string]models.FieldDefinition{
		"DashboardBackground": models.BinaryField{},
	})
}
