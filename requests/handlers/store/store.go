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

// cSpell:ignore skey

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/org"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/store"
	"github.com/objectvault/api-services/requests/rpf/user"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

// TODO IMPLEMENT: READ Store Profile
func GetStore(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.STORE", c, 1000, shared.JSONResponse)

	// Required Roles : Store Access with Read Function
	roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_STORE, orm.FUNCTION_READ)}

	// Base Validation for Store Request
	store.AddinGroupValidateStoreRequest(request, func(o string) interface{} {
		switch o {
		case "check-not-admin":
			return true
		case "check-user-unlocked":
			return true
		case "check-user-roles":
			return true
		case "roles":
			return roles
		}

		return nil
	})

	// Request Processing Chain
	request.Append(
		// GET Store Entry
		store.DBGetStoreByID,
		// GET Store's Organization Entry
		func(r rpf.GINProcessor, c *gin.Context) {
			store := r.MustGet("store").(*orm.Store)
			r.SetLocal("org-id", store.Organization())
		},
		org.DBRegistryOrgStoreFind,
		// Export Results //
		store.ExportStoreFull,
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: UPDATE Store Profile
func PutStorePUT(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.STORE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		/* REQUEST VALIDATION */
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Abort(5999, nil)
		},
	}

	// Start Request Processing
	request.Run()
}

func IsStoreOpen(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.IS.STORE.OPEN", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Basic Request Settings
		session.AssertUserSession,
		store.ExtractGINParameterStore,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Credentials
			sid := r.MustGet("request-store").(uint64)
			skey := session.CreateStoreKey(sid)

			// Get Session Store
			s := sessions.Default(c)

			// Do we have a Store Key Cached?
			key := s.Get(skey)

			// Does Store Key Exist in Session?
			if key == nil { // NO
				r.SetResponseDataValue("open", false)
				return
			}

			// TODO: Make the Store Open Configurable
			ss, e := common.ImportStoreSession(key.(string), 5)
			if e != nil {
				s.Delete(skey)
				r.Abort(5010 /* TODO: Invalid Session Store Key */, nil)
				return
			}

			rif := rpf.NestedIF(r,
				func(r rpf.ProcessorIF, c *gin.Context) {
					if ss.IsExpired() { // NO:
						r.ContinueFalse()
					} else {
						r.ContinueTrue()
					}
				},
				func(r rpf.ProcessorIF, c *gin.Context) {
					request.Append(
						func(r rpf.GINProcessor, c *gin.Context) {
							// Extend Store Session (BY: 1 Minute). This avoids situations in which the next request might fail if we were close to expiration period
							ss.Extend(1)

							// Save Session and Key
							r.SetLocal("store-session", ss)
							r.SetLocal("store-key", ss.Key())

							// Set Response
							r.SetResponseDataValue("open", true)
						},
						session.SessionStoreSave, // Update Store Session
						session.SaveSession,      // Update Session Cookie
					)
				},
				func(r rpf.ProcessorIF, c *gin.Context) {
					request.Append(
						func(r rpf.GINProcessor, c *gin.Context) {
							s.Delete(skey)
							r.SetResponseDataValue("open", false)
						},
						session.SaveSession, // Update Session Cookie
					)
				})

			rif.Run()
		},
	}

	// Start Request Processing
	request.Run()
}

func OpenStore(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("POST.STORE.OPEN", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		user.ExtractFormParameterCredentials,
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Required Roles : At least Organization Role to Read Store
			roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_STORE, orm.FUNCTION_READ)}

			// Initialize Request
			store.GroupOrgStoreRequestInitialize(r, roles, true, true).
				Run()
		},
		/*
			// Validate Basic Store Permissions
			func(r rpf.GINProcessor, c *gin.Context) {
				// Required Roles : At least Allow Read of Store
				roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_STORE_ROLES, orm.FUNCTION_READONLY)}

				// Get Session User
				userID := r.MustGet("user-id").(uint64)

				// Get Store ID
				storeID := r.MustGet("store-id").(uint64)

				// Initialize Request
				store.GroupAssertUserStorePermissions(r, userID, storeID, roles, true).
					Run()
			},
		*/
		session.AssertNotSystemAdmin,
		store.DBGetRegistryStoreUser,
		// REQUEST Validation - POST Parameters //
		/* TODO: PROBLEM: Currently if User Changes Password
		 * All of the Stores Keys Have to be unsealed and re-sealed with
		 * the new password, since it's used as the decryption key for store
		 * passwords
		 */
		session.SessionStoreOpen,
		session.SessionStoreSave,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Organization Information
			rus := r.MustGet("registry-store-user").(*orm.ObjectUserRegistry)
			r.SetResponseDataValue("store", rus.Object())
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func DeleteCloseStore(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.STORE.CLOSE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Basic Request Settings
		store.ExtractGINParameterStore,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Credentials
			id := r.MustGet("request-store").(uint64)
			sid := fmt.Sprintf("_s:%d", id)

			// Get Session Store
			session := sessions.Default(c)

			// Do we have a Store Key Cached?
			key := session.Get(sid)
			if key != nil { // YES: Remove it
				session.Delete(sid)
			}
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}
