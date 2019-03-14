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
	"strconv"
	"strings"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/operator"
	"github.com/hexya-erp/hexya/src/tools/logging"
	"github.com/hexya-erp/hexya/src/tools/strutils"
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

// findNextIndexIgnore returns the index of the first occcurence of b, while ignoring parts of the string that are
// enclosed in the key-value of each entry in ignore map
func findNextIndexIgnore(str []byte, b byte, ignore map[byte]byte) int {
	var index int
	for len(str) > 0 && str[0] != 0 {
		c := str[0]
		if c == b {
			return index
		} else if v, ok := ignore[c]; ok {
			i := findNextIndexIgnore(str[1:], v, ignore) + 2
			index += i
			str = str[i:]
		} else {
			index++
			str = str[1:]
		}
	}
	return index
}

// Returns a slice of strings, splited with the split string, while ignoring parts of the strings that are
// enclosed in the key-value of each entry in ingore map
// May need to be moved elsewhere (in hexya?)
func splitIgnoreParenthesis(str string, split byte, ignore map[byte]byte) []string {
	var out []string
	var cur []byte
	s := []byte(str)

	for len(s) > 0 && s[0] != 0 {
		c := s[0]
		if c == split {
			out = append(out, string(cur))
			cur = []byte{}
			s = s[1:]
		} else if v, ok := ignore[c]; ok {
			index := findNextIndexIgnore(s[1:], v, ignore) + 2
			cur = append(cur, s[:index]...)
			s = s[index:]
		} else {
			cur = append(cur, s[0])
			s = s[1:]
		}
	}
	out = append(out, string(cur))
	return out
}

// ContainsOnly returns true if the given str contains only characters from the given set
func ContainsOnly(str string, set string) bool {
	for _, s := range []byte(str) {
		inSet := false
		for _, b := range []byte(set) {
			if s == b {
				inSet = true
			}
		}
		if !inSet {
			return false
		}
	}
	return true
}

// CleanStringQuotes returns a string without both side string quotes (if it has any)
func CleanStringQuotes(str string) string {
	if strutils.StartsAndEndsWith(str, "\"", "\"") {
		return strings.TrimSuffix(strings.TrimPrefix(str, "\""), "\"")
	} else if strutils.StartsAndEndsWith(str, "'", "'") {
		return strings.TrimSuffix(strings.TrimPrefix(str, "'"), "'")
	}
	return str
}

func parseUnknownBasicVar(str string) interface{} {
	str = strings.TrimSpace(str)
	if strutils.StartsAndEndsWith(str, "\"", "\"") || strutils.StartsAndEndsWith(str, "'", "'") {
		// string
		return CleanStringQuotes(str)
	}
	switch {
	case str == "True" || str == "true":
		// positive boolean
		return true
	case str == "False" || str == "false":
		// negative boolean
		return false
	case ContainsOnly(str, "-0123456789"):
		// integer
		v, err := strconv.Atoi(str)
		if err != nil {
			panic(err)
		}
		return v
	case ContainsOnly(str, "-01232456789."):
		//float
		v, err := strconv.ParseFloat(str, 64)
		if err != nil {
			panic(err)
		}
		return v
	default:
		// unknown
		return nil
	}
}

// ParseString returns a Domain that corresponds to the supposedly well formatted domain given as a string
func ParseString(str string) *Domain {
	var out Domain
	var ignoreMap = map[byte]byte{
		'"':  '"',
		'\'': '\'',
		'(':  ')',
		'{':  '}',
		'[':  ']',
	}
	// remove border brackets
	str = strings.TrimSuffix(strings.TrimPrefix(str, "["), "]")
	// split to get all domain terms
	tuples := splitIgnoreParenthesis(str, ',', ignoreMap)
	for _, tuple := range tuples {
		tuple = strings.TrimSpace(tuple)
		if []byte(tuple)[0] == '(' {
			tuple = strings.TrimSuffix(strings.TrimPrefix(tuple, "("), ")")
		} else if []byte(tuple)[0] == '[' {
			tuple = strings.TrimSuffix(strings.TrimPrefix(tuple, "["), "]")
		}
		terms := splitIgnoreParenthesis(tuple, ',', ignoreMap)
		if len(terms) == 1 {
			// terms is a DomainPrefixOperator
			s := CleanStringQuotes(terms[0])
			out = append(out, s)
		} else if len(terms) == 3 {
			// terms is a DomainTerm
			for i, t := range terms {
				terms[i] = strings.TrimSpace(t)
			}
			var domainTerm []interface{}
			field := CleanStringQuotes(terms[0])
			domainTerm = append(domainTerm, field)
			op := CleanStringQuotes(terms[1])
			domainTerm = append(domainTerm, op)
			value := parseUnknownBasicVar(terms[2])
			domainTerm = append(domainTerm, value)
			out = append(out, domainTerm)
		}
	}
	return &out
}

func init() {
	log = logging.GetLogger("domains")
}
