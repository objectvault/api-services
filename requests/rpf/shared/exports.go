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
	"encoding/json"
	"strings"

	"github.com/objectvault/api-services/orm/query"
)

type ExportList struct {
	List        query.TQueryResults
	ValueMapper TMapListEntryORMtoExport
	FieldMapper TMapFieldNameORMToExternal
}

func orderByToString(asort []query.OrderBy, mapper TMapFieldNameORMToExternal) string {
	fl := []string{}

	if len(asort) > 0 {
		var f string
		for _, isort := range asort {
			if mapper == nil {
				f = isort.Field
			} else {
				f = mapper(isort.Field)
			}

			if isort.Descending {
				f = "!" + f
			}
			fl = append(fl, f)
		}
	}

	if len(fl) == 0 {
		return ""
	} else {
		return strings.Join(fl, ",")
	}
}

func (o *ExportList) MarshalJSON() ([]byte, error) {
	// MAP Values for Export
	items := o.List.Items()
	eitems := make([]interface{}, len(items))

	// Loop Through Array Mapping Entries
	count := 0
	for i, v := range items {
		item := o.ValueMapper(v)
		if item != nil {
			eitems[i] = item
			count += 1
		}
	}

	// List Contains other types of objects?
	if count < len(eitems) { // YES: Shrink Array
		eitems = eitems[:count]
	}

	// MAP Order By for Export
	sortBy := orderByToString(o.List.Sorted(), o.FieldMapper)

	return json.Marshal(&struct {
		Items []interface{} `json:"items"`
		Pager interface{}   `json:"pager"`
		Query interface{}   `json:"query"`
	}{
		Items: eitems,
		Query: struct {
			SortBy string `json:"sortby,omitempty"`
			Filter string `json:"filter,omitempty"`
		}{
			SortBy: sortBy,
		},
		Pager: struct {
			Offset   uint64 `json:"offset"`
			Count    uint64 `json:"count"`
			CountAll uint64 `json:"countAll"`
			Limit    uint64 `json:"pageSize"`
			MaxLimit uint64 `json:"maxPageSize"`
		}{
			Offset:   o.List.Offset(),
			Count:    o.List.Count(),
			CountAll: o.List.MaxCount(),
			Limit:    o.List.Limit(),
			MaxLimit: o.List.MaxLimit(),
		},
	})
}
