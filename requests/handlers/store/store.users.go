// cSpell:ignore ginrpf, gonic, paulo ferreira
package store

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
	"strings"

	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/object"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/store"
	"github.com/objectvault/api-services/requests/rpf/user"
	"github.com/objectvault/api-services/requests/rpf/utils"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func GetStoreUsers(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.STORE.USERS", c, 1000, shared.JSONResponse)

	// Required Roles : Store User Access with List Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_USER, orm.FUNCTION_LIST)}

	// Basic Request Validate
	store.AddinGroupValidateStoreRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		func(r rpf.GINProcessor, c *gin.Context) { // Extract Query Parameters //
			gQuery := shared.GroupExtractQueryConditions(r, nil, func(f string) string {
				switch f {
				case "id":
					return "id_user"
				case "username":
					return "username"
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
		// Query Store for List //
		store.DBRegistryStoreUserList,
		// Export Results //
		object.ExportRegistryObjUsersList,
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: MASS DELETE List of Users
func DeleteStoreUsers(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.STORE.USERS", c, 1000, shared.JSONResponse)

	// Required Roles : Store User Access with Delete Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_USER, orm.FUNCTION_DELETE)}

	// Basic Request Validate
	store.AddinGroupValidateStoreRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		/* REQUEST VALIDATION */
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Abort(5999, nil)
		},
	)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: MASS UPDATE Locked Status of Users List
func PutStoreUsersLock(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.STORE.USERS.LOCK", c, 1000, shared.JSONResponse)

	// Required Roles : Store User Access with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	store.AddinGroupValidateStoreRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		/* REQUEST VALIDATION */
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Abort(5999, nil)
		},
	)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: MASS UPDATE Blocked Status of Users List
func PutStoreUsersBlock(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.STORE.USERS.BLOCK", c, 1000, shared.JSONResponse)

	// Required Roles : Store User Access with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	store.AddinGroupValidateStoreRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		/* REQUEST VALIDATION */
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Abort(5999, nil)
		},
	)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: MASS UPDATE Users Roles in Store
func PutStoreUsersRoles(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.STORE.USERS.ROLES", c, 1000, shared.JSONResponse)

	// Required Roles : Store Roles Access with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_ROLES, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	store.AddinGroupValidateStoreRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		/* REQUEST VALIDATION */
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Abort(5999, nil)
		},
	)

	// Start Request Processing
	request.Run()
}

func GetStoreUser(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.STORE.USER", c, 1000, shared.JSONResponse)

	// Required Roles : Store User Access with Read Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_USER, orm.FUNCTION_READ)}

	// Basic Request Validate
	store.AddinGroupValidateStoreUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		} else if o == "skip-roles-if-self" {
			return true
		}

		return nil
	})

	// Get Object Registry Entry for Request Store / User
	store.AddinRequestStoreUserRegistry(request)

	// Request Processing
	request.Append(
		user.DBGetUserByID,               // Get Record
		object.ExportRegistryObjUserFull, // Create Response
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: DELETE User from Store
func DeleteStoreUser(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.STORE.USER", c, 1000, shared.JSONResponse)

	// Required Roles : Store User Access with Delete Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_USER, orm.FUNCTION_DELETE)}

	// Basic Request Validate
	store.AddinGroupValidateStoreUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Get Object Registry Entry for Request Store / User
	store.AddinRequestStoreUserRegistry(request)

	// Request Processing
	request.Append(
		/* REQUEST VALIDATION */
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Abort(5999, nil)
		},
	)

	// Start Request Processing
	request.Run()
}

func GetStoreUserLock(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.STORE.USER.LOCK", c, 1000, shared.JSONResponse)

	// Required Roles : Store User Access with Read Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_USER, orm.FUNCTION_READ)}

	// Basic Request Validate
	store.AddinGroupValidateStoreUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Get Object Registry Entry for Request Store / User
	store.AddinRequestStoreUserRegistry(request)

	// Request Processing
	request.Append(
		// FIND User by Searching Store User Registry
		store.DBGetRegistryStoreUser,
		// Request Response //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			r.SetResponseDataValue("locked", registry.HasAnyStates(orm.STATE_READONLY))
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

func PutStoreUserLock(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.STORE.USER.LOCK", c, 1000, shared.JSONResponse)

	// Required Roles : Store User Access with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	store.AddinGroupValidateStoreUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Extract : GIN Parameter 'bool' //
	request.Append(
		shared.ExtractGINParameterBooleanValue,
	)

	// Get Object Registry Entry for Request Store / User
	store.AddinRequestStoreUserRegistry(request)

	request.Append(
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			lock := r.MustGet("request-value").(bool)

			if lock {
				registry.SetStates(orm.STATE_READONLY)
			} else {
				registry.ClearStates(orm.STATE_READONLY)
			}
		},
		store.DBRegistryStoreUserUpdate,
		// Request Response //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			r.SetResponseDataValue("locked", registry.HasAnyStates(orm.STATE_READONLY))
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

func GetStoreUserBlock(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.STORE.USER.BLOCK", c, 1000, shared.JSONResponse)

	// Required Roles : Store User Access with Read Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_USER, orm.FUNCTION_READ)}

	// Basic Request Validate
	store.AddinGroupValidateStoreUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Get Object Registry Entry for Request Store / User
	store.AddinRequestStoreUserRegistry(request)

	// Request Processing
	request.Append(
		// Request Response //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			r.SetResponseDataValue("blocked", registry.HasAnyStates(orm.STATE_BLOCKED))
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

func PutStoreUserBlock(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.STORE.USER.BLOCK", c, 1000, shared.JSONResponse)

	// Required Roles : Store User Access with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	store.AddinGroupValidateStoreUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Extract : GIN Parameter 'bool' //
	request.Append(
		shared.ExtractGINParameterBooleanValue,
	)

	// Get Object Registry Entry for Request Store / User
	store.AddinRequestStoreUserRegistry(request)

	request.Append(
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			lock := r.MustGet("request-value").(bool)

			if lock {
				registry.SetStates(orm.STATE_BLOCKED)
			} else {
				registry.ClearStates(orm.STATE_BLOCKED)
			}
		},
		object.DBRegistryObjectUserFlush,
		// Request Response //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			r.SetResponseDataValue("blocked", registry.HasAnyStates(orm.STATE_BLOCKED))
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

func GetStoreUserState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.STORE.USER.STATE", c, 1000, shared.JSONResponse)

	// Required Roles : Store User Access with Read Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_USER, orm.FUNCTION_READ)}

	// Basic Request Validate
	store.AddinGroupValidateStoreUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Get Object Registry Entry for Request Store / User
	store.AddinRequestStoreUserRegistry(request)

	// Request Processing
	request.Append(
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			r.SetResponseDataValue("state", registry.State()&orm.STATE_MASK_FUNCTIONS)
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

func PutStoreUserState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.STORE.USER.STATE", c, 1000, shared.JSONResponse)

	// Required Roles : Store User Access with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	store.AddinGroupValidateStoreUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Extract : GIN Parameter 'uint' //
	request.Append(
		shared.ExtractGINParameterUINTValue,
	)

	// Get Object Registry Entry for Request Store / User
	store.AddinRequestStoreUserRegistry(request)

	request.Append(
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			states := r.MustGet("request-value").(uint64)
			states = states & orm.STATE_MASK_FUNCTIONS

			registry.SetStates(uint16(states))
		},
		object.DBRegistryObjectUserFlush,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			r.SetResponseDataValue("state", registry.State()&orm.STATE_MASK_FUNCTIONS)
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

func DeleteStoreUserState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.STORE.USER.STATE", c, 1000, shared.JSONResponse)

	// Required Roles : Store User Access with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	store.AddinGroupValidateStoreUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Extract : GIN Parameter 'uint' //
	request.Append(
		shared.ExtractGINParameterUINTValue,
	)

	// Get Object Registry Entry for Request Store / User
	store.AddinRequestStoreUserRegistry(request)

	request.Append(
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			states := r.MustGet("request-value").(uint64)
			states = states & orm.STATE_MASK_FUNCTIONS

			registry.ClearStates(uint16(states))
		},
		object.DBRegistryObjectUserFlush,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			r.SetResponseDataValue("state", registry.State()&orm.STATE_MASK_FUNCTIONS)
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

func ToggleStoreUserAdmin(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.STORE.USER.ADMIN", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Initialize Request
			store.GroupStoreUserAdminRequestInitialize(r, nil).
				Run()
		},
		// Extract : GIN Parameter 'uint' //
		shared.ExtractGINParameterUINTValue,
		// FIND User by Searching Store User Registry
		store.DBGetRegistryStoreUser,
		// TODO: Can't Toggle SELF
		// TODO: Last Store Admin Can't be Removed
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-store-user").(*orm.ObjectUserRegistry)

			// Store Administration Roles
			roles := []uint32{0x301FFFF, 0x302FFFF, 0x303FFFF, 0x304FFFF}
			if registry.HasAllStates(orm.STATE_SYSTEM) { // Clear Admin
				registry.ClearStates(orm.STATE_SYSTEM)
				registry.RemoveRoles(roles)
			} else { // Make Admin
				registry.SetState(orm.STATE_SYSTEM)
				registry.AddRoles(roles)
			}
		},
		object.DBRegistryObjectUserFlush,
		// Request Response //
		object.ExportRegistryObjUserFull,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: UPDATE User's Roles in Store
func PutStoreUserRoles(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.STORE.USER.ROLES", c, 1000, shared.JSONResponse)

	// Required Roles : Store Roles Access with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_ROLES, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	store.AddinGroupValidateStoreUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		} else if o == "assert-if-self" {
			return true
		}

		return nil
	})

	// Request Process //
	request.Append(
		// Verify Roles Request Parameter //
		func(r rpf.GINProcessor, c *gin.Context) {
			roles, exists := c.GetPostForm("roles")
			if !exists {
				// TODO: errors.New("Missing Roles Value")
				r.Abort(5202, nil)
				return
			}

			if roles != "" {
				roles = strings.TrimSpace(roles)
			}

			if roles != "" && !utils.IsValidRolesCSV(roles) {
				// TODO: errors.New("Value does not contains a valid Roles CSV List")
				r.Abort(5202, nil)
				return
			}

			r.Set("roles-csv", roles)
		},
	)

	// Get Object Registry Entry for Request Store / User
	store.AddinRequestStoreUserRegistry(request)

	request.Append(
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			csv := r.MustGet("roles-csv").(string)
			registry.RolesFromCSV(csv)
		},
		object.DBRegistryObjectUserFlush,
		// Request Response //
		object.ExportRegistryObjUserBasic,
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}
