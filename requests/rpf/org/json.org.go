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

// ORGANIZATION //
type BasicOrgToJSON struct {
	ID           uint64           // Organization Shard ID
	Organization orm.Organization // Organization Object
}

// JSON Marshaller
func (o *BasicOrgToJSON) MarshalJSON() ([]byte, error) {
	if !o.Organization.IsValid() {
		return nil, errors.New("Missing or Invalid Value [Organization]")
	}

	return json.Marshal(&struct {
		ID    string `json:"id"`
		Alias string `json:"alias"`
		Name  string `json:"name"`
	}{
		ID:    fmt.Sprintf(":%x", o.ID),
		Alias: o.Organization.Alias(),
		Name:  o.Organization.Name(),
	})
}

type FullOrgToJSON struct {
	Registry     *orm.OrgRegistry  // Organization Registry
	Organization *orm.Organization // Organization Object
}

func (o *FullOrgToJSON) MarshalJSON() ([]byte, error) {
	if !o.Registry.IsValid() || !o.Organization.IsValid() {
		return nil, errors.New("Missing or Invalid Value [Registry, Organization]")
	}

	return json.Marshal(&struct {
		ID     string `json:"id"`
		Alias  string `json:"alias"`
		Name   string `json:"name"`
		State  uint16 `json:"state"`
		System bool   `json:"is_system"`
	}{
		ID:     fmt.Sprintf(":%x", o.Registry.ID()),
		Alias:  o.Organization.Alias(),
		Name:   o.Organization.Name(),
		State:  o.Registry.State(),
		System: o.Organization.IsSystem(),
	})
}

// ORGANIZATION REGISTRY //
type BasicRegOrgToJSON struct {
	Registry *orm.OrgRegistry
}

func (o *BasicRegOrgToJSON) MarshalJSON() ([]byte, error) {
	if !o.Registry.IsValid() {
		return nil, errors.New("Missing Required Structore Value [Entry]")
	}

	return json.Marshal(&struct {
		ID     string `json:"id"`
		Alias  string `json:"alias"`
		Name   string `json:"name"`
		System bool   `json:"is_system"`
	}{
		ID:     fmt.Sprintf(":%x", o.Registry.ID()),
		Alias:  o.Registry.Alias(),
		Name:   o.Registry.Name(),
		System: o.Registry.IsSystem(),
	})
}

type FullRegOrgToJSON struct {
	Registry *orm.OrgRegistry
}

func (o *FullRegOrgToJSON) MarshalJSON() ([]byte, error) {
	if !o.Registry.IsValid() {
		return nil, errors.New("Missing Required Structore Value [Entry]")
	}

	return json.Marshal(&struct {
		ID    string `json:"id"`
		Alias string `json:"alias"`
		Name  string `json:"name"`
		State uint16 `json:"state"`
	}{
		ID:    fmt.Sprintf(":%x", o.Registry.ID()),
		Alias: o.Registry.Alias(),
		Name:  o.Registry.Name(),
		State: o.Registry.State(),
	})
}
