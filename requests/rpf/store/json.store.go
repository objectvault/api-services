// cSpell:ignore goginrpf, gonic, paulo ferreira
package store

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
	"errors"
	"fmt"

	"github.com/objectvault/api-services/orm"
)

type BasicStoreToJSON struct {
	ID    uint64
	Store *orm.Store // Store Object
}

// JSON Marshaller
func (o *BasicStoreToJSON) MarshalJSON() ([]byte, error) {
	if !o.Store.IsValid() {
		return nil, errors.New("Missing or Invalid Value [Store]")
	}

	return json.Marshal(&struct {
		ID    string `json:"id"`
		Org   string `json:"organization"`
		Alias string `json:"alias"`
		Name  string `json:"name"`
	}{
		ID:    fmt.Sprintf(":%x", o.ID),
		Org:   fmt.Sprintf(":%x", o.Store.Organization()),
		Alias: o.Store.Alias(),
		Name:  o.Store.Name(),
	})
}

type FullStoreToJSON struct {
	Registry *orm.OrgStoreRegistry // Organization Store Registry
	Store    *orm.Store            // Store Object
}

// JSON Marshaller
func (o *FullStoreToJSON) MarshalJSON() ([]byte, error) {
	if !o.Registry.IsValid() || !o.Store.IsValid() {
		return nil, errors.New("Missing or Invalid Value [Registry, Store]")
	}

	return json.Marshal(&struct {
		ID    string `json:"id"`
		Org   string `json:"organization"`
		Alias string `json:"alias"`
		Name  string `json:"name"`
		State uint16 `json:"state"`
	}{
		ID:    fmt.Sprintf(":%x", o.Registry.Store()),
		Org:   fmt.Sprintf(":%x", o.Registry.Organization()),
		Alias: o.Store.Alias(),
		Name:  o.Store.Name(),
		State: o.Registry.State(),
	})
}
