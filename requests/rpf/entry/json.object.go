// cSpell:ignore goginrpf, gonic, paulo ferreira
package entry

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

type BasicStoreObjectToJSON struct {
	Store  uint64 // Store SHARD ID
	Object *orm.StoreObject
}

// Store Object JSON Export
func (o *BasicStoreObjectToJSON) MarshalJSON() ([]byte, error) {
	if o.Object == nil {
		return nil, errors.New("Missing or Invalid Value [Object]")
	}

	return json.Marshal(&struct {
		Store  string `json:"store"`
		ID     uint32 `json:"id"`
		Parent uint32 `json:"parent"`
		Title  string `json:"title"`
		Type   uint8  `json:"type"`
	}{
		Store:  fmt.Sprintf(":%x", o.Store),
		ID:     o.Object.ID(),
		Parent: o.Object.Parent(),
		Title:  o.Object.Title(),
		Type:   o.Object.Type(),
	})
}

type FullStoreObjectToJSON struct {
	Store    uint64 // Store SHARD ID
	Object   *orm.StoreObject
	Template *orm.StoreTemplateObject
}

// Store Object JSON Export
func (o *FullStoreObjectToJSON) MarshalJSON() ([]byte, error) {
	if o.Object == nil {
		return nil, errors.New("Missing or Invalid Value [Object]")
	}

	// Template Information
	template := &struct {
		Name    string `json:"name"`
		Version uint16 `json:"version"`
	}{
		Name:    o.Template.Template(),
		Version: o.Template.Version(),
	}

	/*
		// Convert to String
		s, e := t.ToString()
		if e != nil {
			return nil, e
		}
	*/
	return json.Marshal(&struct {
		Store    string      `json:"store"`
		Parent   uint32      `json:"parent"`
		ID       uint32      `json:"id"`
		Type     uint8       `json:"type"`
		Title    string      `json:"title"`
		Template interface{} `json:"template"`
		Values   interface{} `json:"values"`
	}{
		Store:    fmt.Sprintf(":%x", o.Store),
		Parent:   o.Object.Parent(),
		ID:       o.Object.ID(),
		Type:     o.Object.Type(),
		Title:    o.Object.Title(),
		Template: template,
		Values:   o.Template.Values(),
	})
}

type StoreFolderObjectToJSON struct {
	Store  uint64 // Store SHARD ID
	Object *orm.StoreObject
}

// JSON Marshaller
func (o *StoreFolderObjectToJSON) MarshalJSON() ([]byte, error) {
	if o.Object == nil {
		return nil, errors.New("Missing or Invalid Value [Object]")
	}

	if o.Object.Type() != orm.OBJECT_TYPE_FOLDER {
		return nil, errors.New("Object is not of Folder Type")
	}

	return json.Marshal(&struct {
		Store  string `json:"store"`
		ID     uint32 `json:"id"`
		Parent uint32 `json:"parent"`
		Title  string `json:"title"`
	}{
		Store:  fmt.Sprintf(":%x", o.Store),
		ID:     o.Object.ID(),
		Parent: o.Object.Parent(),
		Title:  o.Object.Title(),
	})
}
