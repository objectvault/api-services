// cSpell:ignore ferreira, paulo
package orm

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

	"github.com/objectvault/filter-parser/ast"
)

type SQLFWhere struct {
	filter    *ast.Filter
	where     strings.Builder
	args      []interface{}
	mapField  TMapFieldNameExternalToORM
	mapValue  TMapFieldValueExternalToORM
	valid     bool
	processed bool
}

func nullFieldNameMapper(n string) string {
	return n
}

func nullFieldValueMapper(f string, v interface{}) interface{} {
	return v
}

func NewQueryFilterTOWhere(n ast.Node, mf TMapFieldNameExternalToORM, mv TMapFieldValueExternalToORM) *SQLFWhere {
	// Do we have Filter?
	var f *ast.Filter
	if n != nil { // MAYBE: Node is not NIL, but is it a Filter?
		pf, ok := n.(*ast.Filter)
		if ok { // YES
			f = pf
		} else {
			fmt.Println("INVALID AST NODE TYPE, Expecting [ast.Filter]")
		}
	}

	a := &SQLFWhere{
		filter:   f,
		mapField: nullFieldNameMapper,
		mapValue: nullFieldValueMapper,
	}

	// Do we have Field Name Map Function
	if mf != nil { // YES: Use it
		a.mapField = mf
	}

	// Do we have Field Value Map Function
	if mv != nil { // YES: Use it
		a.mapValue = mv
	}
	return a
}

func (a *SQLFWhere) Transpile() string {
	// Has Query been Transpiled Already?
	if !a.processed {
		// NO: Do we have a Filter?
		if a.filter != nil { // YES: Convert to Where Conditions
			m := a.sqlfFunctionToWhere(nil, a.filter.F)
			a.valid = m == ""
			a.processed = a.valid
			return m
		}
	}
	// ELSE: No Filter
	return ""
}

func (a *SQLFWhere) Reset() {
	a.where.Reset()
	a.args = nil
	a.valid = false
	a.processed = false
}

func (a *SQLFWhere) IsValid() bool {
	return a.valid
}

func (a *SQLFWhere) Where() string {
	if a.valid {
		return a.where.String()
	}
	return ""
}

func (a *SQLFWhere) Args() []interface{} {
	if a.valid {
		return a.args
	}
	return nil
}

func (a *SQLFWhere) sqlfFunctionToWhere(p *ast.Function, f *ast.Function) string {
	// ASSUMPTION: Filter has been run through Syntax Checker so AST is Correct
	fname := f.Name.Literal
	switch fname {
	case "NOT":
		return a.sqlfLogicalNOT(f)
	case "AND":
		return a.sqlfLogicalAND(f)
	case "OR":
		return a.sqlfLogicalOR(f)
	case "EQ":
		return a.sqlfOperatorEQ(f)
	case "NEQ":
		return a.sqlfOperatorNEQ(f)
	case "GT":
		return a.sqlfOperatorGT(f)
	case "GTE":
		return a.sqlfOperatorGTE(f)
	case "LT":
		return a.sqlfOperatorLT(f)
	case "LTE":
		return a.sqlfOperatorLTE(f)
	case "CONTAINS":
		return a.sqlfOperatorCONTAINS(f)
	case "IN":
		return a.sqlfOperatorIN(f)
	default:
		return fmt.Sprintf("Unsupported Funcion [%s]", fname)
	}
}

// Logical NOT
func (a *SQLFWhere) sqlfLogicalNOT(fnot *ast.Function) string {
	// 1st Parameter to NOT Should be a Logical Function or Operator Function
	pf1 := (fnot.Parameters[0]).(*ast.Function)

	// Start NOT
	a.where.WriteString("NOT(")

	// Converted Function?
	m := a.sqlfFunctionToWhere(fnot, pf1)
	if m != "" { // NO: Abort
		return m
	}

	// End NOT
	a.where.WriteString(")")
	return ""
}

// Logical AND
func (a *SQLFWhere) sqlfLogicalAND(fand *ast.Function) string {
	// There should be 2 Parameter
	// Both Should be ast.Function
	pf1 := (fand.Parameters[0]).(*ast.Function)
	pf2 := (fand.Parameters[1]).(*ast.Function)

	return a.sqlfBinaryLogical(fand, "AND", pf1, pf2)
}

// Logical OR
func (a *SQLFWhere) sqlfLogicalOR(fop *ast.Function) string {
	// There should be 2 Parameter
	// Both Should be ast.Function
	pf1 := (fop.Parameters[0]).(*ast.Function)
	pf2 := (fop.Parameters[1]).(*ast.Function)

	return a.sqlfBinaryLogical(fop, "OR", pf1, pf2)
}

// Operator EQ
func (a *SQLFWhere) sqlfOperatorEQ(f *ast.Function) string {
	// There should be 2 Parameter
	// Both Should be ast.Value
	pv1 := (f.Parameters[0]).(*ast.Value) // Identifier
	pv2 := (f.Parameters[1]).(*ast.Value)

	return a.sqlfBinaryOperator("=", pv1, pv2)
}

func (a *SQLFWhere) sqlfOperatorNEQ(f *ast.Function) string {
	// There should be 2 Parameter
	// Both Should be ast.Value
	pv1 := (f.Parameters[0]).(*ast.Value) // Identifier
	pv2 := (f.Parameters[1]).(*ast.Value)

	return a.sqlfBinaryOperator("!=", pv1, pv2)
}

func (a *SQLFWhere) sqlfOperatorGT(f *ast.Function) string {
	// There should be 2 Parameter
	// Both Should be ast.Value
	pv1 := (f.Parameters[0]).(*ast.Value) // Identifier
	pv2 := (f.Parameters[1]).(*ast.Value)

	return a.sqlfBinaryOperator(">", pv1, pv2)
}

func (a *SQLFWhere) sqlfOperatorGTE(f *ast.Function) string {
	// There should be 2 Parameter
	// Both Should be ast.Value
	pv1 := (f.Parameters[0]).(*ast.Value) // Identifier
	pv2 := (f.Parameters[1]).(*ast.Value)

	return a.sqlfBinaryOperator(">=", pv1, pv2)
}

func (a *SQLFWhere) sqlfOperatorLT(f *ast.Function) string {
	// There should be 2 Parameter
	// Both Should be ast.Value
	pv1 := (f.Parameters[0]).(*ast.Value) // Identifier
	pv2 := (f.Parameters[1]).(*ast.Value)

	return a.sqlfBinaryOperator("<", pv1, pv2)
}

func (a *SQLFWhere) sqlfOperatorLTE(f *ast.Function) string {
	// There should be 2 Parameter
	// Both Should be ast.Value
	pv1 := (f.Parameters[0]).(*ast.Value) // Identifier
	pv2 := (f.Parameters[1]).(*ast.Value)

	return a.sqlfBinaryOperator("<=", pv1, pv2)
}

func (a *SQLFWhere) sqlfOperatorCONTAINS(f *ast.Function) string {
	// There should be 2 Parameter
	// Both Should be ast.Value
	pv1 := (f.Parameters[0]).(*ast.Value) // Identifier
	pv2 := (f.Parameters[1]).(*ast.Value) // STRING

	// Is Valid Field?
	field := a.mapField(pv1.V.Literal)
	if field == "" { // NO
		return fmt.Sprintf("Invalid Field [%s]", pv1.V.Literal)
	}

	// Do we have a Field Value?
	value := a.mapValue(field, pv2.V.Literal)
	if value == nil { // NO: Abort
		return fmt.Sprintf("Invalid Field [%s] Value [%s]", pv1.V.Literal, pv2.V.Literal)
	}

	a.where.WriteString(fmt.Sprintf("%s LIKE ?", field))
	a.args = append(a.args, sqlfEscapeValue(value))
	return ""
}

func (a *SQLFWhere) sqlfOperatorIN(f *ast.Function) string {
	// There should be 2 Parameter
	// Both Should be ast.Value
	pv1 := (f.Parameters[0]).(*ast.Value) // Identifier
	pv2 := (f.Parameters[1]).(*ast.Value) // STRING

	// Is Valid Field?
	field := a.mapField(pv1.V.Literal)
	if field == "" { // NO
		return fmt.Sprintf("Invalid Field [%s]", pv1.V.Literal)
	}

	// Do we have a Field Value?
	value := a.mapValue(field, pv2.V.Literal)
	if value == nil { // NO: Abort
		return fmt.Sprintf("Invalid Field [%s] Value [%s]", pv1.V.Literal, pv2.V.Literal)
	}

	a.where.WriteString(fmt.Sprintf("%s IN ?", field))
	a.args = append(a.args, sqlfEscapeValue(value))
	return ""
}

// HELPERS //
func (a *SQLFWhere) sqlfBinaryLogical(p *ast.Function, op string, f1 *ast.Function, f2 *ast.Function) string {
	// 1st Parameter
	a.where.WriteString("(")

	// Error Transpiling?
	m := a.sqlfFunctionToWhere(p, f1)
	if m != "" { // YES: Abort
		return m
	}
	a.where.WriteString(fmt.Sprintf(") %s (", op))

	// 2nd Parameter

	// Error Transpiling?
	m = a.sqlfFunctionToWhere(p, f2)
	if m != "" { // YES: Abort
		return m
	}

	// Close
	a.where.WriteString(")")
	return ""
}

func (a *SQLFWhere) sqlfBinaryOperator(op string, f *ast.Value, v *ast.Value) string {
	// 1st Parameter should be an Identifier (Field Name)
	field := a.mapField(f.V.Literal)

	// Is Valid Field?
	if field == "" { // NO
		return fmt.Sprintf("Invalid Field [%s]", f.V.Literal)
	}

	// Do we have a Field Value?
	value := a.mapValue(field, v.V.Literal)
	if value == nil { // NO: Abort
		return fmt.Sprintf("Invalid Field [%s] Value [%s]", f.V.Literal, v.V.Literal)
	}

	a.where.WriteString(fmt.Sprintf("%s %s ?", field, op))
	a.args = append(a.args, sqlfEscapeValue(value))
	return ""
}

func sqlfEscapeValue(v interface{}) interface{} {
	// Is String Value?
	vs, ok := v.(string)
	if !ok {
		return v
	}

	// Escape the Escape Character
	s := strings.ReplaceAll(vs, "\\", "\\\\")

	// Make sure Embedded Quotes are Escaped
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, `'`, `\'`)

	// Escape Special Characters
	s = strings.ReplaceAll(s, `%`, `\%`)

	// Convert '\uFFFD' (replacement for '*') to '%'
	s = strings.ReplaceAll(s, "�", "%") // Temporarily Store \* as something else

	return s
}
