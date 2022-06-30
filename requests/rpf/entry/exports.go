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
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/orm/query"
	"github.com/objectvault/api-services/requests/rpf/shared"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

// OBJECTS //
func ExportStoreObjectList(r rpf.GINProcessor, c *gin.Context) {
	// Get List
	sid := r.MustGet("store-id").(uint64)
	objs := r.Get("store-objects").(query.TQueryResults)

	ores := &shared.ExportList{
		List: objs,
		ValueMapper: func(v interface{}) interface{} {
			return &BasicStoreObjectToJSON{
				Store:  sid,
				Object: v.(*orm.StoreObject),
			}
		},
		FieldMapper: func(f string) string {
			switch f {
			case "id_store":
				return "store"
			case "id":
				return "id"
			case "parent":
				return "parent"
			case "title":
				return "title"
			case "type":
				return "type"
			default:
				return ""
			}
		},
	}

	r.SetResponseDataValue("objects", ores)
}

func ExportStoreObjectFolder(r rpf.GINProcessor, c *gin.Context) {
	// Get Required Information
	sid := r.MustGet("store-id").(uint64)
	obj := r.MustGet("store-object").(*orm.StoreObject)

	// Transform for Export
	d := &StoreFolderObjectToJSON{
		Store:  sid,
		Object: obj,
	}

	r.SetResponseDataValue("object", d)
}

func ExportStoreObjectRegistry(r rpf.GINProcessor, c *gin.Context) {
	// Get Required Information
	sid := r.MustGet("store-id").(uint64)
	obj := r.MustGet("store-object").(*orm.StoreObject)

	// Transform for Export
	d := &BasicStoreObjectToJSON{
		Store:  sid,
		Object: obj,
	}

	r.SetResponseDataValue("object", d)
}

func ExportStoreObjectJSON(r rpf.GINProcessor, c *gin.Context) {
	// Get Required Information
	sid := r.MustGet("store-id").(uint64)
	obj := r.MustGet("store-object").(*orm.StoreObject)
	t := r.MustGet("store-template-object").(*orm.StoreTemplateObject)

	// Transform for Export
	d := &FullStoreObjectToJSON{
		Store:    sid,
		Object:   obj,
		Template: t,
	}

	r.SetResponseDataValue("object", d)
}
