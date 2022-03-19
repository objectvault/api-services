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

type RegUserObjectToJSON struct {
	Registry *orm.UserObjectRegistry
}

func (o *RegUserObjectToJSON) MarshalJSON() ([]byte, error) {
	if o.Registry == nil {
		return nil, errors.New("Missing Required Structure Value [Registry]")
	}

	return json.Marshal(&struct {
		User     string `json:"user"`
		Type     uint16 `json:"type"`
		Object   string `json:"object"`
		Alias    string `json:"alias"`
		Favorite bool   `json:"favorite"`
	}{
		User:     fmt.Sprintf(":%x", o.Registry.User()),
		Type:     o.Registry.Type(),
		Object:   fmt.Sprintf(":%x", o.Registry.Object()),
		Alias:    o.Registry.Alias(),
		Favorite: o.Registry.Favorite(),
	})
}
