// cSpell:ignore bson, paulo ferreira
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

type QueryConditions struct {
	TQueryConditions
	filter *SQLFWhere
	sortBy []OrderBy
	offset *uint64
	limit  *uint64
}

func (o QueryConditions) Filter() *SQLFWhere {
	return o.filter
}

func (o *QueryConditions) SetFilter(filter *SQLFWhere) {
	o.filter = filter
}

func (o QueryConditions) Sort() []OrderBy {
	return o.sortBy
}

func (o *QueryConditions) AppendSort(f string, d bool) {
	by := OrderBy{Field: f, Descending: d}
	o.sortBy = append(o.sortBy, by)
}

func (o QueryConditions) Offset() *uint64 {
	return o.offset
}

func (o *QueryConditions) SetOffset(l uint64) {
	o.offset = &l
}

func (o QueryConditions) Limit() *uint64 {
	return o.limit
}

func (o *QueryConditions) SetLimit(l uint64) {
	o.limit = &l
}
