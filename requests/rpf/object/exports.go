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
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/shared"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

// OBJECT <--> USER //

func ExportRegistryObjUsersList(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entries
	users := r.Get("registry-object-users").(orm.TQueryResults)

	list := &shared.ExportList{
		List: users,
		ValueMapper: func(v interface{}) interface{} {
			return &BasicRegObjectUserToJSON{
				Registry: v.(*orm.ObjectUserRegistry),
			}
		},
		FieldMapper: func(f string) string {
			// MAP ORM FIELDS to JSON
			switch f {
			case "id_object":
				return "object"
			case "id_user":
				return "user"
			case "username":
				return f
			case "state":
				return f
			case "roles":
				return f
			default:
				return ""
			}
		},
	}

	r.SetResponseDataValue("users", list)
}

func ExportRegistryObjUserFull(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization Information
	registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
	user := r.MustGet("user").(*orm.User)

	// Transform for Export
	d := &FullRegObjectUserToJSON{
		Registry: registry,
		User:     user,
	}

	r.SetResponseDataValue("user", d)
}

func ExportRegistryObjUserBasic(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization Information
	registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)

	// Transform for Export
	d := &BasicRegObjectUserToJSON{
		Registry: registry,
	}

	r.SetResponseDataValue("user", d)
}

// USER <--> OBJECT //

func ExportRegistryUserObjsList(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entries
	objects := r.Get("registry-user-objects").(orm.TQueryResults)

	list := &shared.ExportList{
		List: objects,
		ValueMapper: func(v interface{}) interface{} {
			return &RegUserObjectToJSON{
				Registry: v.(*orm.UserObjectRegistry),
			}
		},
		FieldMapper: func(f string) string {
			// MAP ORM FIELDS to JSON
			switch f {
			case "id_user":
				return "user"
			case "type":
				return f
			case "id_object":
				return "object"
			case "alias":
				return f
			case "favorite":
				return f
			default:
				return ""
			}
		},
	}

	r.SetResponseDataValue("objects", list)
}

func ExportRegistryUserObj(r rpf.GINProcessor, c *gin.Context) {
	// Get Required Objects
	registry := r.MustGet("registry-user-object").(*orm.UserObjectRegistry)

	// Transform for Export
	d := &RegUserObjectToJSON{
		Registry: registry,
	}

	r.SetResponseDataValue("object", d)
}
