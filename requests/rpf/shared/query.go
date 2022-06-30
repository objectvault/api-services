// cSpell:ignore goginrpf, gonic, paulo ferreira
package shared

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/objectvault/api-services/orm/query"
	"github.com/objectvault/api-services/requests/rpf/utils"
	"github.com/objectvault/filter-parser/ast"
	"github.com/objectvault/filter-parser/lexer"
	"github.com/objectvault/filter-parser/parser"
	"github.com/objectvault/filter-parser/syntax"
	"github.com/objectvault/filter-parser/token"

	rpf "github.com/objectvault/goginrpf"
)

// Map External Field Name to ORM Field Name
type TFilterPostProcessor = func(n ast.Node) ast.Node

// SINGLE/MULTI FIELD VALIDATION RPF HANDLERS //
func GroupExtractQueryConditions(parent rpf.GINProcessor, post TFilterPostProcessor, fmapper query.TMapFieldNameExternalToORM, vmapper query.TMapFieldValueExternalToORM) *rpf.ProcessorGroup {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = parent

	group.Chain = rpf.ProcessChain{
		// Validate Query Parameters //
		utils.RPFReadyVFields,
		InitializeQueryConditions,
		ExtractQueryFilter,
		func(r rpf.GINProcessor, c *gin.Context) {
			var a ast.Node

			// Get Filter AST
			n := r.Get("filter-ast")
			if n != nil {
				a = n.(ast.Node)
			}

			// Do we have a Post Filter Processor?
			if post != nil { // YES
				a = post(a)
			}

			// Extract Query Conditions
			q := r.MustGet("query-conditions").(*query.QueryConditions)

			// Set Query Filter
			fmapper := r.MustGet("mapper-field-external-to-orm").(query.TMapFieldNameExternalToORM)
			vmapper := r.MustGet("mapper-value-external-to-orm").(query.TMapFieldValueExternalToORM)
			q.SetFilter(query.NewQueryFilterTOWhere(a, fmapper, vmapper))
		},
		ExtractQueryOrderBy,
		ExtractQueryOffset,
		ExtractQueryLimit,
		func(r rpf.GINProcessor, c *gin.Context) {
			utils.RPFTestVFields(3300, r, c)
		},
	}

	// Save Field Mapper
	group.SetLocal("mapper-field-external-to-orm", fmapper)
	group.SetLocal("mapper-value-external-to-orm", vmapper)

	return group
}

// RPFCreateQueryOptions Generate Query Options (if any)
func InitializeQueryConditions(r rpf.GINProcessor, c *gin.Context) {
	q := query.QueryConditions{}
	r.SetLocal("query-conditions", &q)
}

// ExtractQueryFilter See if Request Contains Filter Paramater
func ExtractQueryFilter(r rpf.GINProcessor, c *gin.Context) {
	// Filter Node
	var a ast.Node
	var ok bool

	// Get Fields Error Message Map
	fields := r.Get("v_fields").(map[string]string)

	// FILTER //
	value, message := utils.ValidateURLParameter(c, "filter", false, true, true)
	if message != "" {
		fields["filter"] = message
		return
	}

	// Do we have a Filter Value?
	if value != "" { // YES: Parse it
		// Create New Lexer (for Input)
		l := lexer.NewLexer(value)

		for tok := l.NextToken(); ; tok = l.NextToken() {
			fmt.Printf("TOKEN [%q] - [%q]\n", tok.Type, tok.Literal)
			if tok.Type == token.EOL {
				break
			}
		}

		// Parse Input
		p := parser.NewParser(l.Reset())
		rp := p.ParseFilter()
		a, ok = rp.(ast.Node)
		if !ok {
			fields["filter"] = "FILTER Has Parse Error"
			return
		}

		// Check Resultant AST
		s := syntax.NewSyntaxChecker(a)
		e := s.Verify()
		if e != nil {
			fields["filter"] = fmt.Sprintf("SYNTAX ERROR: %s\n", e.ToString())
			return
		}

		r.SetLocal("filter-ast", a)
	}
}

// RPFExtractQueryOrderBy See if Request Contains Order By Paramater
func ExtractQueryOrderBy(r rpf.GINProcessor, c *gin.Context) {
	// Get Fields Error Message Map
	fields := r.Get("v_fields").(map[string]string)

	// ORDER BY //
	value, message := utils.ValidateURLParameter(c, "sortby", false, true, true)
	if message != "" {
		fields["sortby"] = message
		return
	}

	// Have Value?
	if value == "" || value == "," || value == "!" { // NO
		return
	}

	// Extract Query Conditions
	q := r.MustGet("query-conditions").(*query.QueryConditions)
	mapper := r.MustGet("mapper-field-external-to-orm").(query.TMapFieldNameExternalToORM)

	sortby := strings.Split(value, ",")
	var field string
	var descending bool
	for _, s := range sortby {
		descending = false
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		if s[0:1] == "!" {
			if len(s) == 1 {
				continue
			}

			descending = true
			s = s[1:]
		}

		if mapper == nil {
			field = s
		} else {
			field = mapper(s)
		}

		if field == "" {
			continue
		}

		q.AppendSort(field, descending)
	}
}

// RPFExtractQueryOffset See if Request Contains Offser Parameter
func ExtractQueryOffset(r rpf.GINProcessor, c *gin.Context) {
	// Get Fields Error Message Map
	fields := r.Get("v_fields").(map[string]string)

	value, message := utils.ValidateURLParameter(c, "offset", false, true, true)
	if message != "" {
		fields["offset"] = message
		return
	}

	// Have Value?
	if value == "" { // NO
		return
	}

	// String to Integer
	u, message := utils.ValidateUintParameter("offset", value, true)
	if message != "" {
		fields["offset"] = message
		return
	}

	// Have Value?
	if u != nil { // YES
		// Set Query Offset
		q := r.MustGet("query-conditions").(*query.QueryConditions)
		q.SetOffset(*u)
	}
}

// RPFExtractQueryLimit See if Request Contains Limit Paramater
func ExtractQueryLimit(r rpf.GINProcessor, c *gin.Context) {
	// Get Fields Error Message Map
	fields := r.Get("v_fields").(map[string]string)

	value, message := utils.ValidateURLParameter(c, "limit", false, true, true)
	if message != "" {
		fields["limit"] = message
		return
	}

	// Have Value?
	if value == "" { // NO
		return
	}

	// String to Integer
	u, message := utils.ValidateUintParameter("limit", value, true)
	if message != "" {
		fields["limit"] = message
		return
	}

	// Have Value?
	if u != nil { // YES
		// Set Query Limit
		q := r.MustGet("query-conditions").(*query.QueryConditions)
		q.SetLimit(*u)
	}
}
