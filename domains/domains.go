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

package domains

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/operator"
	"github.com/hexya-erp/hexya/src/tools/logging"
)

// A Domain is a list of search criteria (DomainTerm) in the form of
// a tuplet (field_name, operator, value).
// Domain criteria (DomainTerm) can be combined using logical operators
// in prefix form (DomainPrefixOperator)
type Domain []interface{}

// Value JSON encode our Domain for storing in the database.
func (d Domain) Value() (driver.Value, error) {
	bytes, err := json.Marshal(d)
	return driver.Value(bytes), err
}

// Scan JSON decodes the value of the database into a Domain
func (d *Domain) Scan(src interface{}) error {
	var data []byte
	switch s := src.(type) {
	case string:
		data = []byte(s)
	case []byte:
		data = s
	case []interface{}:
		*d = Domain(s)
		return nil
	default:
		return fmt.Errorf("Invalid type for Domain: %T", src)
	}
	var dom Domain
	err := json.Unmarshal(data, &dom)
	if err != nil {
		return err
	}
	*d = dom
	return nil

}

var _ driver.Valuer = Domain{}
var _ sql.Scanner = &Domain{}

// String method for Domain type. Returns a valid domain for client.
func (d Domain) String() string {
	if len(d) == 0 {
		return "[]"
	}
	var res []string
	for _, term := range d {
		switch t := term.(type) {
		case string:
			res = append(res, t)
		case []interface{}:
			var argStr string
			switch arg := t[2].(type) {
			case nil:
				argStr = `False`
			case string:
				argStr = fmt.Sprintf(`"%s"`, arg)
			default:
				argStr = fmt.Sprintf("%v", arg)
			}
			res = append(res, fmt.Sprintf(`["%s", "%s", %s]`, t[0], t[1], argStr))
		default:
			log.Panic("Unexpected Domain term", "term", term)
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(res, ","))
}

// A DomainTerm is a search criterion in the form of
// a tuplet (field_name, operator, value).
type DomainTerm []interface{}

// A DomainPrefixOperator is used to combine DomainTerms
type DomainPrefixOperator string

// Domain prefix operators
const (
	PREFIX_AND DomainPrefixOperator = "&"
	PREFIX_OR  DomainPrefixOperator = "|"
	PREFIX_NOT DomainPrefixOperator = "!"
)

var log logging.Logger

// ParseDomain gets Domain and parses it into a RecordSet query Condition.
// Returns an empty condition if the domain is []
func ParseDomain(dom Domain) *models.Condition {
	res := parseDomain(&dom)
	if res == nil {
		return &models.Condition{}
	}
	for len(dom) > 0 {
		res = models.Condition{}.AndCond(res).AndCond(parseDomain(&dom))
	}
	return res
}

// parseDomain is the internal recursive function making all the job of
// ParseDomain. The given domain through pointer is deleted during operation.
func parseDomain(dom *Domain) *models.Condition {
	if len(*dom) == 0 {
		return nil
	}

	res := &models.Condition{}
	currentOp := PREFIX_AND

	operatorTerm := (*dom)[0]
	firstTerm := (*dom)[0]
	if ftStr, ok := operatorTerm.(string); ok {
		currentOp = DomainPrefixOperator(ftStr)
		*dom = (*dom)[1:]
		firstTerm = (*dom)[0]
	}

	switch ft := firstTerm.(type) {
	case string:
		// We have a unary operator '|' or '&', so this is an included condition
		// We have AndCond because this is the first term.
		res = res.AndCond(parseDomain(dom))
	case []interface{}:
		// We have a domain leaf ['field', 'op', value]
		term := DomainTerm(ft)
		res = addTerm(res, term, currentOp)
		*dom = (*dom)[1:]
	}

	// dom has been reduced in previous step
	// check if we still have terms to add
	if len(*dom) > 0 {
		secondTerm := (*dom)[0]
		switch secondTerm.(type) {
		case string:
			// We have a unary operator '|' or '&', so this is an included condition
			switch currentOp {
			case PREFIX_OR:
				res = res.OrCond(parseDomain(dom))
			default:
				res = res.AndCond(parseDomain(dom))
			}
		case []interface{}:
			term := DomainTerm(secondTerm.([]interface{}))
			res = addTerm(res, term, currentOp)
			*dom = (*dom)[1:]
		}
	}
	return res
}

// addTerm parses the given DomainTerm and adds it to the given condition with the given
// prefix operator. Returns the new condition.
func addTerm(cond *models.Condition, term DomainTerm, op DomainPrefixOperator) *models.Condition {
	if len(term) != 3 {
		log.Panic("Malformed domain term", "term", term)
	}
	fieldName := term[0].(string)
	optr := operator.Operator(term[1].(string))
	value := term[2]
	newCond := models.Condition{}.And().Field(fieldName).AddOperator(optr, value)
	cond = getConditionMethod(newCond, op)(cond)
	return cond
}

// getConditionMethod returns the condition method to use on the given condition
// for the given prefix operator and negation condition.
func getConditionMethod(cond *models.Condition, op DomainPrefixOperator) func(*models.Condition) *models.Condition {
	switch op {
	case PREFIX_AND:
		return cond.AndCond
	case PREFIX_OR:
		return cond.OrCond
	default:
		log.Panic("Unknown prefix operator", "operator", op)
	}
	return nil
}

func init() {
	log = logging.GetLogger("domains")
}
