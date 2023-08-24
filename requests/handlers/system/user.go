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
	"errors"
	"fmt"

	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/invitation"
	"github.com/objectvault/api-services/requests/rpf/object"
	"github.com/objectvault/api-services/requests/rpf/org"
	"github.com/objectvault/api-services/requests/rpf/queue"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/user"
	"github.com/objectvault/api-services/requests/rpf/utils"
	"github.com/objectvault/api-services/xjson"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// TODO: How to Implement (Use Invitation?)
func PostCreateUser(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("POST.SYSTEM.USER", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Validate Request Basics //
		session.AssertNoUserSession, // Can't be in user session
		func(r rpf.GINProcessor, c *gin.Context) { // Has to Have Invitation ID
			// Get Session Store
			session := sessions.Default(c)

			// Do we have a User Invitation ID Set?
			id := session.Get("invitation-id")
			if id == nil {
				r.Abort(5998 /* TODO: ERROR [Invitation Required]*/, nil)
			}

			// Save User
			r.SetLocal("invitation-id", id)
		},
		shared.RequestExtractJSON, // Has to have a JSON Body
		// Do we have a Valid invitation? //
		invitation.DBGetRegistryInvitationByID,
		invitation.AssertInvitationActive,
		// EXTRACT : Required Information //
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Invitation Object
			i := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

			// Create User Object
			u := &orm.User{}

			// User's Email comes From Invitation
			u.SetEmail(i.InviteeEmail())
			u.SetCreator(i.Creator())

			// Save User
			r.SetLocal("user", u)
		},
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get User Object
			u := r.MustGet("user").(*orm.User)

			// Extract and Validate JSON MEssage
			m := r.MustGet("request-json").(xjson.T_xMap)
			vmap := xjson.S_xJSONMap{Source: m}

			// User Name
			vmap.Required("alias", nil, func(v interface{}) (interface{}, error) {
				v, e := xjson.F_xToTrimmedString(v)
				if e != nil {
					return nil, e
				}

				s := v.(string)
				if !utils.IsValidUserName(s) {
					return nil, errors.New("Value is not a valid user name")
				}
				return s, nil
			}, func(v interface{}) error {
				u.SetUserName(v.(string))
				return nil
			})

			// Password Hash
			vmap.Required("hash", nil, func(v interface{}) (interface{}, error) {
				v, e := xjson.F_xToTrimmedString(v)
				if e != nil {
					return nil, e
				}

				s := v.(string)
				if !utils.IsValidPasswordHash(s) {
					return nil, errors.New("Value is not a valid password hash")
				}
				return s, nil
			}, func(v interface{}) error {
				e := u.SetHash(v.(string))
				return e
			})

			// Name
			vmap.Required("name", nil, func(v interface{}) (interface{}, error) {
				v, e := xjson.F_xToTrimmedString(v)
				if e != nil {
					return nil, e
				}

				s := v.(string)
				return s, nil
			}, func(v interface{}) error {
				u.SetName(v.(string))
				return nil
			})

			/*
				// OPTIONAL: User Country
				vmap.Optional("country", nil, xjson.F_xToString, nil, func(v interface{}) error {
					if v != nil {
						u.SetCountry(v.(string))
					}
					return nil
				})
			*/

			// Did we have an Error Processing the Map?
			if vmap.Error != nil {
				fmt.Println(vmap.Error)
				fmt.Println(vmap.StringSrc())
				r.Abort(5998 /* TODO: ERROR [User Information Invalid] */, nil)
				return
			}
		},
		// Make Sure Invitation ORG Exists and is Active //
		invitation.DBGetInvitationFromRegistry,
		func(r rpf.GINProcessor, c *gin.Context) {
			inv := r.MustGet("invitation").(*orm.Invitation)
			r.SetLocal("org-id", inv.Object())
		},
		org.DBRegistryOrgFindByID,
		org.AssertOrgUnblocked,
		// REGISTER User //
		// TODO: Check if User Name Already Exists
		// TODO: Check if Email Already Exists
		user.DBInsertUser,
		user.DBRegisterUser,
		// REGISTER User With ORG //
		func(r rpf.GINProcessor, c *gin.Context) {
			inv := r.MustGet("invitation").(*orm.Invitation)

			// Do we have Invitation Roles Set?
			if !inv.IsRolesEmpty() { // YES
				r.SetLocal("register-roles", inv.Roles())
			}
		},
		object.DBRegisterUserWithOrg,
		object.DBRegisterOrgWithUser,
		// Update Invitation //
		invitation.DBInvitationAccepted,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Session Store
			session := sessions.Default(c)

			// Clear Invitation ID from Session
			session.Delete("invitation-id")
		},
		// RESPONSE //
		user.ExportUserMe,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func GetUserProfile(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.SYSTEM.USER", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract : GIN Parameter 'user' //
		user.ExtractGINParameterUser,
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			id := r.MustGet("request-user").(string)
			r.Set("user", id)
		},
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()
			if !r.IsFinished() { // YES
				// Get Session User
				user_id := gSessionUser.MustGet("user-id").(uint64)

				// Required Roles : System User Access with Read Function
				roles := []uint32{orm.Role(orm.CATEGORY_SYSTEM|orm.SUBCATEGORY_USER, orm.FUNCTION_READ)}

				// Check User has Permissions in System Organization
				org.GroupAssertUserOrganizationPermissions(r, user_id, uint64(0), roles, true, true, false).
					Run()

				// Session Requirements Passed?
				if !r.IsFinished() { // YES: Save User Information
					r.SetLocal("user-id", user_id)
				}
			}
		},
		// REQUEST: Get User (by Way of Registry) //
		user.DBRegistryUserFind,
		func(r rpf.GINProcessor, c *gin.Context) {
			user := r.MustGet("registry-user").(*orm.UserRegistry)
			r.SetLocal("user-id", user.ID())
		},
		user.DBGetUserByID,
		// Export Results //
		user.ExportUserSystem,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: DELETE User from the System
func DeleteUser(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.SYSTEM.USER", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{}

	// Required Roles : System User Role with Delete Function
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

	// Validate User
	request.Append(
		// Extract : GIN Parameter 'user' //
		user.ExtractGINParameterUserID,
		// Can't Delete Self
		session.AssertIfSelf,
		// REQUEST: Get User (by Way of Registry) //
		func(r rpf.GINProcessor, c *gin.Context) {
			r.SetLocal("user-id", r.MustGet("request-user"))
		},
		user.DBRegistryUserFindByID,
	)

	// Block User and Mark as Being Deleted
	request.Append(
		user.AssertUserNotDeleted,
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-user").(*orm.UserRegistry)
			registry.SetStates(orm.STATE_BLOCKED)
			registry.SetStates(orm.STATE_DELETE)
		},
		user.DBRegistryUserUpdate,
	)

	// Queue Action
	request.Append(
		// IMPORTANT: As long as the invitation is created (but not published to the queue) the handler passes
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Session Store
			session := sessions.Default(c)

			// Get Session User's Information for User
			r.SetLocal("action-user", session.Get("user-id"))
			r.SetLocal("action-user-name", session.Get("user-name"))
			r.SetLocal("action-user-email", session.Get("user-email"))

			// Message Queue
			r.SetLocal("queue", "q.actions.inbox")
		},
		queue.CreateMessageDeleteUserFromSystem,
		queue.SendQueueMessage,
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: MODIFY User (HOW TO - Should another user be allowed to modify user profile?)
func PutUserProfile(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.SYSTEM.USER", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{}

	// Implementation Incomplete
	shared.AddinToDo(request, nil)

	// Start Request Processing
	request.Run()
}

func GetUserLockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.SYSTEM.USER.LOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract : GIN Parameter 'user' //
		user.ExtractGINParameterUser,
		func(r rpf.GINProcessor, c *gin.Context) {
			id := r.MustGet("request-user").(string)
			r.Set("user", id)
		},
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()
			if !r.IsFinished() { // YES
				// Get Session User
				user_id := gSessionUser.MustGet("user-id").(uint64)

				// Required Roles : System User Access with Read Function
				roles := []uint32{orm.Role(orm.CATEGORY_SYSTEM|orm.SUBCATEGORY_USER, orm.FUNCTION_READ)}

				// Check User has Permissions in System Organization
				org.GroupAssertUserOrganizationPermissions(r, user_id, uint64(0), roles, true, true, false).
					Run()

				// Session Requirements Passed?
				if !r.IsFinished() { // YES: Save User Information
					r.SetLocal("user-id", user_id)
				}
			}
		},
		// SEARCH Registry for Entry
		user.DBRegistryUserFind,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-user").(*orm.UserRegistry)
			r.SetResponseDataValue("locked", registry.HasAnyStates(orm.STATE_READONLY))
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func PutUserLockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.SYSTEM.USER.LOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract : GIN Parameter 'user' //
		user.ExtractGINParameterUser,
		func(r rpf.GINProcessor, c *gin.Context) {
			id := r.MustGet("request-user").(string)
			r.Set("user", id)
		},
		// Extract : GIN Parameter 'bool' //
		shared.ExtractGINParameterBooleanValue,
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()
			if !r.IsFinished() { // YES
				// Get Session User
				user_id := gSessionUser.MustGet("user-id").(uint64)

				// Required Roles : System User Access with Modify Function
				roles := []uint32{orm.Role(orm.CATEGORY_SYSTEM|orm.SUBCATEGORY_USER, orm.FUNCTION_UPDATE)}

				// Check User has Permissions in System Organization
				org.GroupAssertUserOrganizationPermissions(r, user_id, uint64(0), roles, true, true, false).
					Run()

				// Session Requirements Passed?
				if !r.IsFinished() { // YES: Save User Information
					r.SetLocal("user-id", user_id)
				}
			}
		},
		// SEARCH Regisrty for Entry
		user.DBRegistryUserFind,
		// UPDATE Registry Entry
		user.AssertNotSystemUserRegistry,
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-user").(*orm.UserRegistry)
			lock := r.MustGet("request-value").(bool)

			if lock {
				registry.SetStates(orm.STATE_READONLY)
			} else {
				registry.ClearStates(orm.STATE_READONLY)
			}
		},
		user.DBRegistryUserUpdate,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-user").(*orm.UserRegistry)
			r.SetResponseDataValue("locked", registry.HasAnyStates(orm.STATE_READONLY))
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func GetUserBlockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.SYSTEM.USER.BLOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract : GIN Parameter 'user' //
		user.ExtractGINParameterUser,
		func(r rpf.GINProcessor, c *gin.Context) {
			id := r.MustGet("request-user").(string)
			r.Set("user", id)
		},
		// Validate Session Users Permission
		func(r rpf.GINProcessor, c *gin.Context) {
			// Is User Session?
			gSessionUser := session.GroupGetSessionUser(r, true, false)
			gSessionUser.Run()
			if !r.IsFinished() { // YES
				// Get Session User
				user_id := gSessionUser.MustGet("user-id").(uint64)

				// Required Roles : System User Access with Read Function
				roles := []uint32{orm.Role(orm.CATEGORY_SYSTEM|orm.SUBCATEGORY_USER, orm.FUNCTION_READ)}

				// Check User has Permissions in System Organization
				org.GroupAssertUserOrganizationPermissions(r, user_id, uint64(0), roles, true, true, false).
					Run()

				// Session Requirements Passed?
				if !r.IsFinished() { // YES: Save User Information
					r.SetLocal("user-id", user_id)
				}
			}
		},
		// SEARCH Registry for Entry
		user.DBRegistryUserFind,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-user").(*orm.UserRegistry)
			r.SetResponseDataValue("blocked", registry.HasAnyStates(orm.STATE_BLOCKED))
		},
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func PutUserBlockState(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.SYSTEM.USER.BLOCK", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{}

	// Required Roles : System User Role with Delete Function
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

	// Validate User
	request.Append(
		// Extract : GIN Parameter 'user' //
		user.ExtractGINParameterUserID,
		// Can't Delete Self
		session.AssertIfSelf,
		// Extract : GIN Parameter 'bool' //
		shared.ExtractGINParameterBooleanValue,
		// REQUEST: Get User (by Way of Registry) //
		func(r rpf.GINProcessor, c *gin.Context) {
			r.SetLocal("user-id", r.MustGet("request-user"))
		},
		user.DBRegistryUserFindByID,
		// UPDATE Registry Entry
		user.AssertNotSystemUserRegistry,
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-user").(*orm.UserRegistry)
			lock := r.MustGet("request-value").(bool)

			if lock {
				registry.SetStates(orm.STATE_BLOCKED)
			} else {
				registry.ClearStates(orm.STATE_BLOCKED)
			}
		},
		user.DBRegistryUserUpdate,
		// CALCULATE RESPONSE //
		func(r rpf.GINProcessor, c *gin.Context) {
			registry := r.MustGet("registry-user").(*orm.UserRegistry)
			r.SetResponseDataValue("blocked", registry.HasAnyStates(orm.STATE_BLOCKED))
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}
