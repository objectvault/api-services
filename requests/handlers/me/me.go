// cSpell:ignore ginrpf, gonic, orgs, paulo, ferreira
package me

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

// cSpell:ignore objs, vmap, xjson

import (
	"errors"
	"fmt"

	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/object"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/user"
	"github.com/objectvault/api-services/requests/rpf/utils"
	"github.com/objectvault/api-services/xjson"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func GetMe(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ME", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()
			if !gSessionUser.IsFinished() { // YES
				gSessionUser.LocalToGlobal("registry-user")
				gSessionUser.LocalToGlobal("user-id")
			}
		},
		user.DBGetUserByID,
		// RESPONSE //
		user.ExportUserMe,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: UPDATE Your Profile
func PutMe(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ME", c, 1000, shared.JSONResponse)

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

// TODO IMPLEMENT: DELETE Your Account
func DeleteMe(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.ME", c, 1000, shared.JSONResponse)

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

func ChangePassword(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("POST.ME.PASSWORD", c, 1000, shared.JSONResponse)

	// SESSION: We have an active session for user that is not blocked
	session.AddinActiveUserSession(request, func(o string) interface{} {
		if o == "check-user-unlocked" {
			return true
		}

		return nil
	})

	// Request Process //
	request.Append(
		// GET JSON Body //
		shared.RequestExtractJSON,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Extract and Validate JSON Message
			m := r.MustGet("request-json").(xjson.T_xMap)
			vmap := xjson.S_xJSONMap{Source: m}

			// Current Password Hash
			vmap.Required("current", nil, func(v interface{}) (interface{}, error) {
				v, e := xjson.F_xToTrimmedString(v)
				if e != nil {
					return nil, e
				}

				s := v.(string)
				if !utils.IsValidPasswordHash(s) {
					return nil, errors.New("SYSTEM ERROR: Invalid Current Password")
				}
				return s, nil
			}, func(v interface{}) error {
				r.Set("hash", v.(string))
				return nil
			})

			// Current Password Hash
			vmap.Required("new", nil, func(v interface{}) (interface{}, error) {
				v, e := xjson.F_xToTrimmedString(v)
				if e != nil {
					return nil, e
				}

				s := v.(string)
				if !utils.IsValidPasswordHash(s) {
					return nil, errors.New("SYSTEM ERROR: Invalid New Password")
				}
				return s, nil
			}, func(v interface{}) error {
				r.Set("new-hash", v.(string))
				return nil
			})

			// Did we have an Error Processing the Map?
			if vmap.Error != nil {
				fmt.Println(vmap.Error)
				fmt.Println(vmap.StringSrc())
				r.Abort(5202, nil)
				return
			}
		},
		user.AssertCredentials, // See if User Password Correct
		user.DBGetUserByID,     // GET User Object
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Hashes
			hash := r.MustGet("hash").(string)
			newHash := r.MustGet("new-hash").(string)

			// Update User Password
			u := r.MustGet("user").(*orm.User)
			e := u.UpdateHash(hash, newHash)
			if e != nil {
				r.Abort(5400, nil)
				return
			}

			// Update User Registry Password
			ur := r.MustGet("registry-user").(*orm.UserRegistry)
			e = ur.UpdatePassword(u)
			if e != nil {
				r.Abort(5400, nil)
				return
			}

			// Set Modifier
			e = u.SetModifier(ur.ID())
			if e != nil {
				r.Abort(5900, nil)
			}
		},
		user.DBUserUpdate,
		user.DBRegistryUserUpdate,
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: GET My Links
func GetMyObjects(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ME.OBJECTS", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			g := session.GroupGetSessionUser(r, true, false)
			g.Run()
			if !r.IsFinished() { // YES
				g.LocalToGlobal("registry-user")
				g.LocalToGlobal("user-id")
			}
		},
		// Extract Query Parameters //
		func(r rpf.GINProcessor, c *gin.Context) {
			gQuery := shared.GroupExtractQueryConditions(r, nil, func(f string) string {
				// MAP JSON / URL FIELDS to ORM FIELDS
				switch f {
				case "user":
					return "id_user"
				case "object":
					return "id_object"
				default: // Other Fields
					return f
				}
			}, nil)

			gQuery.Run()
			if !r.IsFinished() { // YES
				// Save Query Settings as Global
				gQuery.LocalToGlobal("query-conditions")
			}
		},
		// Query User for List //
		object.DBRegistryUserObjsList,
		// Export Results //
		object.ExportRegistryUserObjsList,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: GET My Link
func GetMyObject(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ME.LINK.{ID}", c, 1000, shared.JSONResponse)

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

// TODO IMPLEMENT: GET My Favorite Links
func GetMyFavoriteObjects(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ME.FAVORITES", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()
			if !r.IsFinished() { // YES
				gSessionUser.LocalToGlobal("registry-user")
				gSessionUser.LocalToGlobal("user-id")
			}
		},
		// Extract Query Parameters //
		func(r rpf.GINProcessor, c *gin.Context) {
			gQuery := shared.GroupExtractQueryConditions(r, nil, func(f string) string {
				// DIRECT JSON to ORM MAP
				return f
			}, nil)

			gQuery.Run()
			if !r.IsFinished() { // YES
				// Save Query Settings as Global
				gQuery.LocalToGlobal("query-conditions")
			}
		},
		// Query User for List //
		object.DBRegistryUserObjsList,
		// Export Results //
		object.ExportRegistryUserObjsList,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: GET Toggle Favorite Link
func ToggleLinkFavorite(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.ME.FAVORITE.TOGGLE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()
			if !r.IsFinished() { // YES
				gSessionUser.LocalToGlobal("registry-user")
				gSessionUser.LocalToGlobal("user-id")
			}
		},
		// Get URL Parameter
		user.ExtractGINParameterObject,
		// Extract Query Parameters //
		func(r rpf.GINProcessor, c *gin.Context) {
			gQuery := shared.GroupExtractQueryConditions(r, nil, func(f string) string {
				// MAP JSON / URL FIELDS to ORM FIELDS
				switch f {
				case "user":
					return "id_user"
				case "object":
					return "id_object"
				default: // Other Fields
					return f
				}
			}, nil)

			gQuery.Run()
			if !r.IsFinished() { // YES
				// Save Query Settings as Global
				gQuery.LocalToGlobal("query-conditions")
			}
		},
		// Query User for Link //
		func(r rpf.GINProcessor, c *gin.Context) {
			id := r.MustGet("request-object")
			r.SetLocal("object-id", id)
		},
		object.DBRegistryUserObjFindOrNil,
		// Update Registry Object //
		func(r rpf.GINProcessor, c *gin.Context) {
			// Toggle Favorite Flag
			registry := r.MustGet("registry-user-object").(*orm.UserObjectRegistry)
			registry.SetFavorite(!registry.Favorite())
		},
		object.DBRegistryUserObjFlush,
		// Export Results //
		object.ExportRegistryUserObj,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: LIST Organizations you belong to
func GetMyOrgs(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ME.ORGS", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()
			if !r.IsFinished() { // YES
				gSessionUser.LocalToGlobal("registry-user")
				gSessionUser.LocalToGlobal("user-id")
			}
		},
		// Extract Query Parameters //
		func(r rpf.GINProcessor, c *gin.Context) {
			gQuery := shared.GroupExtractQueryConditions(r, nil, func(f string) string {
				// MAP JSON / URL FIELDS to ORM FIELDS
				switch f {
				case "user":
					return "id_user"
				case "object":
					return "id_object"
				default: // Other Fields
					return f
				}
			}, nil)

			gQuery.Run()
			if !r.IsFinished() { // YES
				// Save Query Settings as Global
				gQuery.LocalToGlobal("query-conditions")
			}
		},
		// Query User for List //
		object.DBRegistryUserObjsList,
		// Export Results //
		object.ExportRegistryUserObjsList,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: DELETE Your Access to an Organization
func DeleteMeFromOrg(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.ME.ORG", c, 1000, shared.JSONResponse)

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

// TODO IMPLEMENT: LIST Organization's Stores you belong to
func GetMyOrgStores(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ME.ORG.STORES", c, 1000, shared.JSONResponse)

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

// TODO IMPLEMENT: LIST Stores you belong to
func GetMyStores(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ME.STORES", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()
			if !r.IsFinished() { // YES
				gSessionUser.LocalToGlobal("registry-user")
				gSessionUser.LocalToGlobal("user-id")
			}
		},
		// Extract Query Parameters //
		func(r rpf.GINProcessor, c *gin.Context) {
			gQuery := shared.GroupExtractQueryConditions(r, nil, func(f string) string {
				// MAP JSON / URL FIELDS to ORM FIELDS
				switch f {
				case "user":
					return "id_user"
				case "object":
					return "id_object"
				default: // Other Fields
					return f
				}
			}, nil)

			gQuery.Run()
			if !r.IsFinished() { // YES
				// Save Query Settings as Global
				gQuery.LocalToGlobal("query-conditions")
			}
		},
		// Query User for List //
		user.DBRegistryUserStoreList,
		// Export Results //
		object.ExportRegistryUserObjsList,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: DELETE Your Access to a Store
func DeleteMeFromStore(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.ME.STORE", c, 1000, shared.JSONResponse)

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
