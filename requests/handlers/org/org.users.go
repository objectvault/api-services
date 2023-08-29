// cSpell:ignore ginrpf, gonic, paulo ferreira
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
	"strings"

	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/object"
	"github.com/objectvault/api-services/requests/rpf/org"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/utils"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

// Service Handlers //

func GetOrgUsers(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.USERS", c, 1000, shared.JSONResponse)

	// Required Roles : Organization User Access with List Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_USER, orm.FUNCTION_LIST)}

	// Basic Request Validate
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		} else if o == "assert-org-user-readonly" {
			return false
		}

		return nil
	})

	// Request Process //
	request.Append(
		// Extract Query Parameters //
		func(r rpf.GINProcessor, c *gin.Context) {
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
		// Query Organization for List //
		object.DBRegistryOrgUsersList,
		// Export Results //
		object.ExportRegistryObjUsersList,
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: MASS DELETE List of Users
func DeleteOrgUsers(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.ORG.USERS", c, 1000, shared.JSONResponse)

	// Required Roles : Organization User Access with Delete Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_USER, orm.FUNCTION_DELETE)}

	// Basic Request Validate
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
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
func PutOrgUsersLock(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.USERS.LOCK", c, 1000, shared.JSONResponse)

	// Required Roles : Organization User Access with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	org.AddinGroupValidateOrgUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		func(r rpf.GINProcessor, c *gin.Context) {
			// TODO Implement
			r.Abort(5999, nil)
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: MASS UPDATE Blocked Status of Users List
func PutOrgUsersBlock(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.USERS.BLOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{}

	// Required Roles : Organization User Access with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	org.AddinGroupValidateOrgUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		func(r rpf.GINProcessor, c *gin.Context) {
			// TODO Implement
			r.Abort(5999, nil)
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: MASS UPDATE Users Roles in Organization
func PutOrgUsersRoles(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.USERS.ROLES", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{}

	// Required Roles : Organization Roles Access with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ROLES, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	org.AddinGroupValidateOrgUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		func(r rpf.GINProcessor, c *gin.Context) {
			// TODO Implement
			r.Abort(5999, nil)
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: READ Organization's User Profile
func GetOrgUser(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.USER", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{}

	// Required Roles : Organization User Access with Read Function (or SELF)
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_USER, orm.FUNCTION_READ)}

	// Basic Request Validate
	org.AddinGroupValidateOrgUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		} else if o == "skip-roles-if-self" {
			return true
		}

		return nil
	})

	// Request Process //
	request.Append(
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get ID from Registry
			id := r.MustGet("request-user").(uint64)
			r.SetLocal("user-id", id)
		},
		object.DBOrgUserFind,
		// Request Response //
		object.ExportRegistryObjUserBasic,
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: DELETE User from Organization
func DeleteOrgUser(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.ORG.USER", c, 1000, shared.JSONResponse)

	// Required Roles : Organization User Access with Delete Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_USER, orm.FUNCTION_DELETE)}

	// Basic Request Validate
	org.AddinGroupValidateOrgUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		func(r rpf.GINProcessor, c *gin.Context) {
			// TODO Implement
			r.Abort(5999, nil)
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: READ User's Locked Status in Organization
func GetOrgUserLock(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.USER.LOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{}

	// Required Roles : Organization User Access with Read Function (or Self)
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_USER, orm.FUNCTION_READ)}

	// Basic Request Validate
	org.AddinGroupValidateOrgUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		// Load Request User Org Registration
		func(r rpf.GINProcessor, c *gin.Context) {
			r.SetLocal("user-id", r.MustGet("request-user"))
			object.DBOrgUserFind(r, c)
		},
		// CALCULATE RESPONSE //
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

func PutOrgUserLock(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.USER.LOCK", c, 1000, shared.JSONResponse)

	// Required Roles : Organization User Access with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	org.AddinGroupValidateOrgUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		} else if o == "assert-if-self" {
			return true
		}

		return nil
	})

	// Request Process //
	request.Append(
		// Extract : GIN Parameter 'bool' //
		shared.ExtractGINParameterBooleanValue,
		// Load Session User Org Registration
		func(r rpf.GINProcessor, c *gin.Context) {
			r.SetLocal("user-id", r.MustGet("request-user"))
			object.DBOrgUserFind(r, c)
		},
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
		object.DBObjectUserFlush,
		// CALCULATE RESPONSE //
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

func GetOrgUserBlock(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.USER.BLOCK", c, 1000, shared.JSONResponse)

	// Required Roles : Organization User Access with Read Function (or Self)
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_USER, orm.FUNCTION_READ)}

	// Basic Request Validate
	org.AddinGroupValidateOrgUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		// Load Request User Org Registration
		func(r rpf.GINProcessor, c *gin.Context) {
			r.SetLocal("user-id", r.MustGet("request-user"))
			object.DBOrgUserFind(r, c)
		},
		// CALCULATE RESPONSE //
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

func PutOrgUserBlock(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.USER.BLOCK", c, 1000, shared.JSONResponse)

	// Required Roles : Organization User Access with Update Function (or Self)
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	org.AddinGroupValidateOrgUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		} else if o == "assert-if-self" {
			return true
		}

		return nil
	})

	// Request Process //
	request.Append(
		// Extract : GIN Parameter 'bool' //
		shared.ExtractGINParameterBooleanValue,
		// Load Request User Org Registration
		func(r rpf.GINProcessor, c *gin.Context) {
			r.SetLocal("user-id", r.MustGet("request-user"))
			object.DBOrgUserFind(r, c)
		},
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
		object.DBObjectUserFlush,
		// CALCULATE RESPONSE //
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

func GetOrgUserState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.USER.STATE", c, 1000, shared.JSONResponse)

	// Required Roles : Organization User Access with Read Function (or Self)
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_USER, orm.FUNCTION_READ)}

	// Basic Request Validate
	org.AddinGroupValidateOrgUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		// FIND Store by Searching Org Store Registry
		object.DBOrgUserFind,
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

func PutOrgUserState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.USER.STATE", c, 1000, shared.JSONResponse)

	// Required Roles : Organization User Access with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	org.AddinGroupValidateOrgUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		// Extract : GIN Parameter 'uint' //
		shared.ExtractGINParameterUINTValue,
		// SEARCH Registry for Entry
		object.DBOrgUserFind,
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			states := r.MustGet("request-value").(uint64)
			states = states & orm.STATE_MASK_FUNCTIONS

			registry.SetStates(uint16(states))
		},
		object.DBObjectUserFlush,
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

// TODO IMPLEMENT: CLEAR User's State in Organization
func DeleteOrgUserState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.ORG.USER.STATE", c, 1000, shared.JSONResponse)

	// Required Roles : Organization User Access with Update Function (or Self)
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	org.AddinGroupValidateOrgUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		// Extract : GIN Parameter 'uint' //
		shared.ExtractGINParameterUINTValue,
		// SEARCH Registry for Entry
		object.DBOrgUserFind,
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			states := r.MustGet("request-value").(uint64)
			states = states & orm.STATE_MASK_FUNCTIONS

			registry.ClearStates(uint16(states))
		},
		object.DBObjectUserFlush,
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

// TODO IMPLEMENT: Toggle User's Admin State (Make or Clear State)
func ToggleOrgUserAdmin(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.USER.ADMIN", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Initialize Request
			org.GroupOrgUserAdminRequestInitialize(r, nil).
				Run()
		},
		// Extract : GIN Parameter 'uint' //
		shared.ExtractGINParameterUINTValue,
		// SEARCH Registry for Entry
		object.DBOrgUserFind,
		// TODO: Can't Toggle SELF
		// TODO: Last Organization Admin Can't be Removed
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)

			if registry.HasAllStates(orm.STATE_SYSTEM) { // Clear Admin
				registry.ClearStates(orm.STATE_SYSTEM)

				if registry.IsSystemOrganization() { // Remove Administration Roles
					// All System Organization Roles and Roles and Invitation Management
					roles := []uint32{0x301FFFF, 0x302FFFF, 0x303FFFF, 0x304FFFF}
					registry.RemoveRoles(roles)
				} else {
					// Remove Roles and Invitation Management
					roles := []uint32{0x204FFFF, 0x205FFFF}
					registry.AddRoles(roles)
				}
			} else { // Make Admin
				registry.SetStates(orm.STATE_SYSTEM)

				if registry.IsSystemOrganization() {
					// Add All System Organization Administration Roles
					roles := []uint32{0x101FFFF, 0x102FFFF, 0x103FFFF, 0x201FFFF, 0x202FFFF, 0x203FFFF, 0x204FFFF, 0x205FFFF}
					registry.AddRoles(roles)
				} else {
					// Add All Normal Organization Administration Roles
					roles := []uint32{0x201FFFF, 0x202FFFF, 0x203FFFF, 0x204FFFF, 0x205FFFF}
					registry.AddRoles(roles)
				}
			}
		},
		object.DBObjectUserFlush,
		// Request Response //
		object.ExportRegistryObjUserFull,
	}

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: UPDATE User's Roles in Organization
func PutOrgUserRoles(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.USER.ROLES", c, 1000, shared.JSONResponse)

	// Required Roles : Organization Roles Access with Update Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_ROLES, orm.FUNCTION_UPDATE)}

	// Basic Request Validate
	org.AddinGroupValidateOrgUserRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Request Process //
	request.Append(
		// Verify User Credentials //
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
		// SEARCH Registry for Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			id := r.MustGet("request-user")
			r.SetLocal("user-id", id)
		},
		object.DBOrgUserFind,
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)
			csv := r.MustGet("roles-csv").(string)
			registry.RolesFromCSV(csv)
		},
		object.DBObjectUserFlush,
		// Request Response //
		object.ExportRegistryObjUserBasic,
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}
