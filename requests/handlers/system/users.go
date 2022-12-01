// cSpell:ignore ginrpf, gonic, paulo ferreira
package system

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
	"github.com/objectvault/api-services/requests/rpf/org"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/user"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.SYSTEM.USERS", c, 1000, shared.JSONResponse)

	// Required Roles : Organization 0 Access with Roles Users List
	roles := []uint32{orm.Role(orm.CATEGORY_SYSTEM|orm.SUBCATEGORY_USER, orm.FUNCTION_LIST)}

	// TODO Session in System Mode? Should we force mode uses?
	// Do Basic ORG Request Validation
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		switch o {
		case "system-organization":
			return true
		case "roles":
			return roles
		}

		return nil
	})

	// Request Processing
	request.Append(
		func(r rpf.GINProcessor, c *gin.Context) {
			gQuery := shared.GroupExtractQueryConditions(r, nil, func(f string) string {
				switch f {
				case "id":
					return "id_user"
				case "alias":
					return "username"
				case "email":
					return "email"
				case "name":
					return "name"
				case "state": // Can not Sort, but can Filter
					return "state"
				default: // Invalid Field
					return ""
				}
			}, nil)
			gQuery.Run()
			if !r.IsFinished() { // YES
				// Save Query Settings as Global
				gQuery.LocalToGlobal("query-conditions")
			}
		},
		// Query System for List //
		user.DBRegistryUserList,
		// Export Results //
		user.ExportRegistryUserList,
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: MASS DELETE List of Users
func DeleteUsers(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.SYSTEM.USERS", c, 1000, shared.JSONResponse)

	// Required Roles : Organization 0 Access with Roles User Delete
	roles := []uint32{orm.Role(orm.CATEGORY_SYSTEM|orm.SUBCATEGORY_USER, orm.FUNCTION_DELETE)}

	// Do Basic ORG Request Validation
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		switch o {
		case "system-organization":
			return true
		case "roles":
			return roles
		}

		return nil
	})

	// Request Processing
	request.Append(
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Abort(5999, nil)
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: MASS UPDATE Locked Status of Users List
func PutUsersLock(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.SYSTEM.USERS.LOCK", c, 1000, shared.JSONResponse)

	// Required Roles : Organization 0 Access with Roles User Update
	roles := []uint32{orm.Role(orm.CATEGORY_SYSTEM|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

	// Do Basic ORG Request Validation
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		switch o {
		case "system-organization":
			return true
		case "roles":
			return roles
		}

		return nil
	})

	// Request Processing
	request.Append(
		// Extract : GIN Parameter 'bool' //
		shared.ExtractGINParameterBooleanValue,
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Abort(5999, nil)
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: MASS UPDATE Blocked Status of Users List
func PutUsersBlock(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.SYSTEM.USERS.BLOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	// Required Roles : Organization 0 Access with Roles User Update
	roles := []uint32{orm.Role(orm.CATEGORY_SYSTEM|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

	// Do Basic ORG Request Validation
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		switch o {
		case "system-organization":
			return true
		case "roles":
			return roles
		}

		return nil
	})

	// Request Processing
	request.Append(
		// Extract : GIN Parameter 'bool' //
		shared.ExtractGINParameterBooleanValue,
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Abort(5999, nil)
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}
