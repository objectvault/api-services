// cSpell:ignore goginrpf, gonic, paulo ferreira
package object

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

// OBJECT USER //
type BasicRegObjectUserToJSON struct {
	Registry *orm.ObjectUserRegistry
}

func (o *BasicRegObjectUserToJSON) MarshalJSON() ([]byte, error) {
	if !o.Registry.IsValid() {
		return nil, errors.New("Missing Required Structore Value [Entry]")
	}

	return json.Marshal(&struct {
		Object string `json:"object"`
		User   string `json:"user"`
		Alias  string `json:"username"`
		State  uint16 `json:"state"`
		Roles  string `json:"roles"`
	}{
		Object: fmt.Sprintf(":%x", o.Registry.Object()),
		User:   fmt.Sprintf(":%x", o.Registry.User()),
		Alias:  o.Registry.UserName(),
		State:  o.Registry.State(),
		Roles:  o.Registry.RolesToCSV(),
	})
}

type NoRolesRegObjectUserToJSON struct {
	Registry *orm.ObjectUserRegistry
}

func (o *NoRolesRegObjectUserToJSON) MarshalJSON() ([]byte, error) {
	if !o.Registry.IsValid() {
		return nil, errors.New("Missing Required Structore Value [Entry]")
	}

	return json.Marshal(&struct {
		Object string `json:"object"`
		User   string `json:"user"`
		Alias  string `json:"username"`
		State  uint16 `json:"state"`
	}{
		Object: fmt.Sprintf(":%x", o.Registry.Object()),
		User:   fmt.Sprintf(":%x", o.Registry.User()),
		Alias:  o.Registry.UserName(),
		State:  o.Registry.State(),
	})
}

type FullRegObjectUserToJSON struct {
	Registry *orm.ObjectUserRegistry
	User     *orm.User // User Object
}

func (o *FullRegObjectUserToJSON) MarshalJSON() ([]byte, error) {
	if !o.Registry.IsValid() {
		return nil, errors.New("Missing Required Structore Value [Entry]")
	}

	return json.Marshal(&struct {
		Object string `json:"object"`
		User   string `json:"user"`
		Alias  string `json:"username"`
		Email  string `json:"email"`
		Name   string `json:"name"`
		State  uint16 `json:"state"`
		Roles  string `json:"roles"`
	}{
		Object: fmt.Sprintf(":%x", o.Registry.Object()),
		User:   fmt.Sprintf(":%x", o.Registry.User()),
		Alias:  o.User.UserName(),
		Email:  o.User.Email(),
		Name:   o.User.Name(),
		State:  o.Registry.State(),
		Roles:  o.Registry.RolesToCSV(),
	})
}
