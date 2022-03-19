// cSpell:ignore goginrpf, gonic, paulo ferreira
package user

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

// USER //
type BasicUserToJSON struct {
	ID   uint64    // User Shard ID
	User *orm.User // User Object
}

func (o *BasicUserToJSON) MarshalJSON() ([]byte, error) {
	if o.User == nil {
		return nil, errors.New("Missing Required Structure Value [User]")
	}

	return json.Marshal(&struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}{
		ID:       fmt.Sprintf(":%x", o.ID),
		Name:     o.User.Name(),
		Username: o.User.UserName(),
		Email:    o.User.Email(),
	})
}

type FullUserToJSON struct {
	Registry *orm.UserRegistry // User Registry
	User     *orm.User         // User Object
}

func (o *FullUserToJSON) MarshalJSON() ([]byte, error) {
	if o.Registry == nil || o.User == nil {
		return nil, errors.New("Missing Required Structure Value [Registry, User]")
	}

	return json.Marshal(&struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Email    string `json:"email"`
		State    uint16 `json:"state"`
	}{
		ID:       fmt.Sprintf(":%x", o.Registry.ID()),
		Name:     o.User.Name(),
		Username: o.User.UserName(),
		Email:    o.User.Email(),
		State:    o.Registry.State(),
	})
}

// USER REGISTRY //
type BasicRegUserToJSON struct {
	Registry   *orm.UserRegistry
	Registered bool
}

func (o *BasicRegUserToJSON) MarshalJSON() ([]byte, error) {
	if o.Registry == nil {
		return nil, errors.New("Missing Required Structure Value [Registry]")
	}

	return json.Marshal(&struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Username   string `json:"username"`
		Email      string `json:"email"`
		Registered bool   `json:"registered"`
	}{
		ID:         fmt.Sprintf(":%x", o.Registry.ID()),
		Name:       o.Registry.Name(),
		Username:   o.Registry.UserName(),
		Email:      o.Registry.Email(),
		Registered: o.Registered,
	})
}

type FullRegUserToJSON struct {
	Registry   *orm.UserRegistry
	Registered bool
}

func (o *FullRegUserToJSON) MarshalJSON() ([]byte, error) {
	if o.Registry == nil {
		return nil, errors.New("Missing Required Structure Value [Registry]")
	}

	return json.Marshal(&struct {
		ID         string `json:"id"`
		Alias      string `json:"alias"`
		Email      string `json:"email"`
		Name       string `json:"name"`
		State      uint16 `json:"state"`
		Registered bool   `json:"registered"`
	}{
		ID:         fmt.Sprintf(":%x", o.Registry.ID()),
		Alias:      o.Registry.UserName(),
		Email:      o.Registry.Email(),
		Name:       o.Registry.Name(),
		State:      o.Registry.State(),
		Registered: o.Registered,
	})
}
