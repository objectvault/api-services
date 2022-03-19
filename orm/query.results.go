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

type QueryResults struct {
	TQueryResults
	items    []interface{}
	sortBy   []OrderBy
	offset   uint64
	limit    uint64
	maxLimit uint64
	maxCount uint64
}

func (o QueryResults) Sorted() []OrderBy {
	return o.sortBy
}

func (o *QueryResults) AppendSort(f string, d bool) {
	by := OrderBy{Field: f, Descending: d}
	o.sortBy = append(o.sortBy, by)
}

func (o QueryResults) Items() []interface{} {
	return o.items
}

func (o *QueryResults) AppendValue(v interface{}) {
	o.items = append(o.items, v)
	/*
	  if o.items == nil {
	    o.items = make([]interface{}, 1)
	    o.items[0] = v
	  } else {
	    o.items = append(orgs, o)
	  }
	*/
}

func (o QueryResults) Offset() uint64 {
	if o.limit == 0 {
		return 0
	}
	return o.offset
}

func (o *QueryResults) SetOffset(l uint64) {
	o.offset = l
}

func (o QueryResults) Limit() uint64 {
	// Do we have a Limit Set?
	if o.limit == 0 { // NO: Return Maximum Limit
		return o.MaxLimit()
	}

	return o.limit
}

func (o *QueryResults) SetLimit(l uint64) {
	// Is Limit Greater than Max Limit (if set)?
	if (o.maxLimit == 0) || (o.maxLimit != 0 && l < o.maxLimit) { // NO: Okay
		o.limit = l
		return
	}
	// ELSE: Use Max Limit
	o.limit = o.maxLimit
}

func (o QueryResults) MaxLimit() uint64 {
	// Do we have a limit set?
	if o.limit == 0 { // NO: Then Max Limit === Max Records
		return o.Count()
	}

	// Do we have a Maximum Limit Set?
	if o.maxLimit == 0 { // NO: Then Max Limit === Offset + Limit
		return o.limit + o.offset
	}

	return o.maxLimit
}

func (o *QueryResults) SetMaxLimit(l uint64) {
	if o.limit > l {
		o.limit = l
	}

	o.maxLimit = l
}

func (o QueryResults) Count() uint64 {
	/*
	  if o.items == nil {
	    o.items = make([]interface{}, 1)
	    o.items[0] = v
	  } else {
	    o.items = append(orgs, o)
	  }
	*/

	return uint64(len(o.items))
}

func (o QueryResults) MaxCount() uint64 {
	// Do we have Maximum Count Set?
	if o.maxCount == 0 { // NO: Then Max Count === Record Count
		return o.Count()
	}
	return o.maxCount
}

func (o *QueryResults) SetMaxCount(c uint64) {
	o.maxCount = c
}
