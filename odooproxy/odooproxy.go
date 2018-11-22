// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package odooproxy

import "strings"

// ConvertModelName converts an Odoo dotted style model name (e.g. res.partner) into
// a Hexya Pascal cased style (e.g. Partner).
func ConvertModelName(val string) string {
	var res string
	switch val {
	case "res.users":
		res = "User"
	case "res.partner":
		res = "Partner"
	case "res.groups":
		res = "Group"
	case "res.company":
		res = "Company"
	case "ir.filters":
		res = "Filter"
	case "ir.attachment":
		res = "Attachment"
	case "ir.translation":
		res = "Translation"
	case "res.currency":
		res = "Currency"
	case "res.currency.rate":
		res = "CurrencyRate"
	default:
		tokens := strings.Split(val, ".")
		for _, token := range tokens {
			res += strings.Title(token)
		}
	}
	return res
}

// ConvertMethodName converts an Odoo snake style method name (e.g. search_read) into
// a Hexya Pascal cased style (e.g. SearchRead).
func ConvertMethodName(val string) string {
	var res string
	tokens := strings.Split(val, "_")
	for _, token := range tokens {
		res += strings.Title(token)
	}
	return res
}
