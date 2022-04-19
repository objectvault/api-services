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
	"github.com/gin-contrib/sessions"
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/shared"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func ExportRegistryUserList(r rpf.GINProcessor, c *gin.Context) {
	// Get User Registry Entries
	users := r.MustGet("users").(orm.TQueryResults)

	ures := &shared.ExportList{
		List: users,
		ValueMapper: func(v interface{}) interface{} {
			return &FullRegUserToJSON{
				Registry: v.(*orm.UserRegistry),
			}
		},
		FieldMapper: func(f string) string {
			switch f {
			case "id_user":
				return "id"
			case "username":
				return "alias"
			case "email":
				return "email"
			case "name":
				return "name"
			case "state":
				return "state"
			default:
				return ""
			}
		},
	}

	r.SetResponseDataValue("users", ures)
}

func ExportUserMe(r rpf.GINProcessor, c *gin.Context) {
	// Get User Information
	registry := r.MustGet("registry-user").(*orm.UserRegistry)

	// Transform for Export
	d := &BasicRegUserToJSON{
		Registry: registry,
	}

	r.SetResponseDataValue("user", d)
}

func ExportUserSystem(r rpf.GINProcessor, c *gin.Context) {
	// Get User Information
	registry := r.MustGet("registry-user").(*orm.UserRegistry)
	user := r.MustGet("user").(*orm.User)

	// Transform for Export
	d := &FullUserToJSON{
		Registry: registry,
		User:     user,
	}

	r.SetResponseDataValue("user", d)
}

func ExportUserSession(r rpf.GINProcessor, c *gin.Context) {
	// Get the User From the Context
	registry := r.MustGet("registry-user").(*orm.UserRegistry)

	// Get Session Store
	session := sessions.Default(c)
	hash := session.Get("user-hash")

	// Convert User to Response Object
	data := &FullRegUserToJSON{
		Registry:   registry,
		Registered: hash != nil,
	}

	r.SetResponseDataValue("user", data)
}
