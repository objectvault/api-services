package query

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
	"github.com/pjacferreira/sqlf"
)

type OrderBy struct {
	Field      string
	Descending bool
}

type FilterBy struct {
	Field    string
	Operator string
	Value    string
}

type TQueryConditions interface {
	Filter() *SQLFWhere
	Sort() []OrderBy
	Offset() *uint64
	Limit() *uint64
}

type TQueryResults interface {
	Items() []interface{}
	Sorted() []OrderBy
	Offset() uint64
	Limit() uint64
	MaxLimit() uint64
	Count() uint64
	MaxCount() uint64
}

func ApplyFilterConditions(s *sqlf.Stmt, q TQueryConditions) error {
	// Do we have Sort Conditions
	if q != nil && q.Filter() != nil {
		f := q.Filter()
		m := f.Transpile()

		// TODO: Improve Error Handling
		if m == "" && f.IsValid() {
			w := f.Where()
			a := f.Args()
			s.WhereArgsArray(w, a)
		}
	}

	return nil
}
