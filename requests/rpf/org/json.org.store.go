// cSpell:ignore goginrpf, gonic, paulo ferreira
package org

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

type BasicRegOrgStoreToJSON struct {
	Registry *orm.OrgStoreRegistry
}

func (o *BasicRegOrgStoreToJSON) MarshalJSON() ([]byte, error) {
	if !o.Registry.IsValid() {
		return nil, errors.New("Missing Required Structore Value [Registry]")
	}

	return json.Marshal(&struct {
		Org   string `json:"organization"`
		Store string `json:"store"`
		Alias string `json:"alias"`
	}{
		Org:   fmt.Sprintf(":%x", o.Registry.Organization()),
		Store: fmt.Sprintf(":%x", o.Registry.Store()),
		Alias: o.Registry.StoreAlias(),
	})
}

type FullRegOrgStoreToJSON struct {
	Registry *orm.OrgStoreRegistry
}

func (o *FullRegOrgStoreToJSON) MarshalJSON() ([]byte, error) {
	if !o.Registry.IsValid() {
		return nil, errors.New("Missing Required Structore Value [Registry]")
	}

	return json.Marshal(&struct {
		Org   string `json:"organization"`
		Store string `json:"store"`
		Alias string `json:"alias"`
		State uint16 `json:"state"`
	}{
		Org:   fmt.Sprintf(":%x", o.Registry.Organization()),
		Store: fmt.Sprintf(":%x", o.Registry.Store()),
		Alias: o.Registry.StoreAlias(),
		State: o.Registry.State(),
	})
}
