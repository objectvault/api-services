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
	"errors"
	"strings"

	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/org"
	sharedorg "github.com/objectvault/api-services/requests/rpf/org"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/store"
	"github.com/objectvault/api-services/requests/rpf/user"
	"github.com/objectvault/api-services/requests/rpf/utils"
	"github.com/objectvault/api-services/xjson"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

// TODO IMPLEMENT: LIST Organization Stores
func GetOrgStores(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.STORES", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameters
		sharedorg.ExtractGINParameterOrg,
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Organization ID
			oid := r.MustGet("request-org").(uint64)

			// Required Roles : Organization Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_STORE, orm.FUNCTION_LIST)}

			// Initialize Request
			g := sharedorg.GroupOrgRequestInitialize(r, oid, roles, false)
			g.Run()
		},
		// Extract Query Parameters //
		func(r rpf.GINProcessor, c *gin.Context) {
			gQuery := shared.GroupExtractQueryConditions(r, nil, func(f string) string {
				switch f {
				case "id":
					return "id_store"
				case "store":
					return "storename"
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
		org.DBRegistryOrgStoreList,
		// Export Results //
		org.ExportRegistryOrgStoreList,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func PostCreateStore(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("POST.ORG.STORE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameters
		sharedorg.ExtractGINParameterOrg,
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Organization ID
			oid := r.MustGet("request-org").(uint64)

			// Required Roles : Organization Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_STORE, orm.FUNCTION_CREATE)}

			// Initialize Request
			org.GroupOrgRequestInitialize(r, oid, roles, false).
				Run()
		},
		// TODO ASSERT NOT SYSTEM ADMIN (System Admin Cannot Create Stores)
		shared.RequestExtractJSON,
		// Verify User Credentials //
		func(r rpf.GINProcessor, c *gin.Context) {
			// Extract and Validat Post Parameters
			m := r.MustGet("request-json").(xjson.T_xMap)
			vmap := xjson.S_xJSONMap{Source: m}

			// Store ALIAS
			vmap.Required("credentials", nil, func(v interface{}) (interface{}, error) {
				v, e := xjson.F_xToTrimmedString(v)
				if e != nil {
					return nil, e
				}

				s := strings.ToLower(v.(string))
				if !utils.IsValidPasswordHash(s) {
					return nil, errors.New("Value does not contains a valid password hash")
				}
				return s, nil
			}, func(v interface{}) error {
				r.Set("hash", v.(string))
				return nil
			})

			// Did we have an Error Processing the Map?
			if vmap.Error != nil {
				r.Abort(5202, nil)
				return
			}
		},
		user.DBGetUserByID,
		user.AssertCredentials,
		// Create Store From Post //
		store.CreateFromJSON,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Parent Organization ID
			orgID := r.MustGet("org-id").(uint64)

			// Get Store
			store := r.MustGet("store").(*orm.Store)

			// Set Parent Organization
			store.SetOrganization(orgID)
		},
		store.DBInsertStore,
		org.DBRegisterStoreWithOrg,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Set Default Store Administration Roles
			roles := []uint32{0x301FFFF, 0x302FFFF, 0x303FFFF, 0x304FFFF, 0x306FFFF}
			r.SetLocal("register-roles", roles)
			r.SetLocal("register-as-admin", true)
		},
		store.DBRegisterUserWithNewStore,
		store.DBRegisterStoreWithUser,
		// Request Response //
		store.ExportStoreFull,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func GetStoreProfile(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.STORE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Required Roles : Organization Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_STORE, orm.FUNCTION_READONLY)}

			// Initialize Request
			store.GroupOrgStoreRequestInitialize(r, roles, true, false).
				Run()
		},
		// Get Store
		store.DBGetStoreByID,
		sharedorg.DBRegistryOrgStoreFind,
		// Request Response //
		store.ExportStoreFull,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func PutStoreProfile(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.STORE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Required Roles : Organization Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_STORE, orm.FUNCTION_MODIFY)}

			// Initialize Request
			store.GroupOrgStoreRequestInitialize(r, roles, true, true).
				Run()
		},
		// Update Store From JSON //
		store.DBGetStoreByID,
		store.UpdateFromJSON,
		store.DBUpdateStore,
		org.DBRegistryUpdateFromStore,
		// Request Response //
		store.ExportStoreFull,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func DeleteStore(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.ORG.STORE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Required Roles : Organization Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_STORE, orm.FUNCTION_DELETE)}

			// Initialize Request
			store.GroupOrgStoreRequestInitialize(r, roles, true, false).
				Run()
		},
		// Delete Store Information
		org.DBRegistryDeleteStore,
		store.DBDeleteStoreByID,
		// Request Response //
		func(r rpf.GINProcessor, c *gin.Context) {
			// TODO What Value to Return?
			r.SetReponseDataValue("ok", true)
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func GetOrgStoreLockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.STORE.LOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Required Roles : Organization Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_STORE, orm.FUNCTION_READONLY)}

			// Initialize Request
			store.GroupOrgStoreRequestInitialize(r, roles, true, false).
				Run()
		},
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)
			r.SetReponseDataValue("locked", registry.HasAnyStates(orm.STATE_READONLY))
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func PutOrgStoreLockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.STORE.LOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Required Roles : Organization Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_STORE, orm.FUNCTION_MODIFY)}

			// Initialize Request
			store.GroupOrgStoreRequestInitialize(r, roles, true, false).
				Run()
		},
		// Extract : GIN Parameter 'bool' //
		shared.ExtractGINParameterBooleanValue,
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)
			lock := r.MustGet("request-value").(bool)

			if lock {
				registry.SetStates(orm.STATE_READONLY)
			} else {
				registry.ClearStates(orm.STATE_READONLY)
			}
		},
		org.DBRegistryOrgStoreUpdate,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)
			r.SetReponseDataValue("locked", registry.HasAnyStates(orm.STATE_READONLY))
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func GetOrgStoreBlockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.STORE.BLOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Required Roles : Organization Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_STORE, orm.FUNCTION_READONLY)}

			// Initialize Request
			store.GroupOrgStoreRequestInitialize(r, roles, true, false).
				Run()
		},
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)
			r.SetReponseDataValue("blocked", registry.HasAnyStates(orm.STATE_BLOCKED))
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func PutOrgStoreBlockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.STORE.BLOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Required Roles : Organization Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_STORE, orm.FUNCTION_MODIFY)}

			// Initialize Request
			store.GroupOrgStoreRequestInitialize(r, roles, true, false).
				Run()
		},
		// Extract : GIN Parameter 'bool' //
		shared.ExtractGINParameterBooleanValue,
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)
			lock := r.MustGet("request-value").(bool)

			if lock {
				registry.SetStates(orm.STATE_BLOCKED)
			} else {
				registry.ClearStates(orm.STATE_BLOCKED)
			}
		},
		org.DBRegistryOrgStoreUpdate,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)
			r.SetReponseDataValue("blocked", registry.HasAnyStates(orm.STATE_READONLY))
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func GetOrgStoreState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ORG.STORE.STATE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Required Roles : Organization Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_STORE, orm.FUNCTION_READONLY)}

			// Initialize Request
			store.GroupOrgStoreRequestInitialize(r, roles, true, false).
				Run()
		},
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)
			r.SetReponseDataValue("state", registry.State()&orm.STATE_MASK_FUNCTIONS)
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func PutOrgStoreState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ORG.STORE.STATE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Required Roles : Organization Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_STORE, orm.FUNCTION_MODIFY)}

			// Initialize Request
			store.GroupOrgStoreRequestInitialize(r, roles, true, false).
				Run()
		},
		// Extract : GIN Parameter 'uint' //
		shared.ExtractGINParameterUINTValue,
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)
			states := r.MustGet("request-value").(uint64)
			states = states & orm.STATE_MASK_FUNCTIONS

			registry.SetStates(uint16(states))
		},
		org.DBRegistryOrgStoreUpdate,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)
			r.SetReponseDataValue("state", registry.State()&orm.STATE_MASK_FUNCTIONS)
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func DeleteOrgStoreState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PDELETEUT.ORG.STORE.STATE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Required Roles : Organization Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_STORE, orm.FUNCTION_MODIFY)}

			// Initialize Request
			store.GroupOrgStoreRequestInitialize(r, roles, true, false).
				Run()
		},
		// Extract : GIN Parameter 'uint' //
		shared.ExtractGINParameterUINTValue,
		// UPDATE Registry Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)
			states := r.MustGet("request-value").(uint64)
			states = states & orm.STATE_MASK_FUNCTIONS

			registry.ClearStates(uint16(states))
		},
		org.DBRegistryOrgStoreUpdate,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)
			r.SetReponseDataValue("state", registry.State()&orm.STATE_MASK_FUNCTIONS)
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}
