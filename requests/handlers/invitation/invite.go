// cSpell:ignore goginrpf, gonic, orgs, paulo, ferreira
package invitation

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

// cSpell:ignore ASTAND, ASTEQ, OTYPE, vmap, xjson

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/invitation"
	"github.com/objectvault/api-services/requests/rpf/keys"
	"github.com/objectvault/api-services/requests/rpf/object"
	"github.com/objectvault/api-services/requests/rpf/org"
	"github.com/objectvault/api-services/requests/rpf/queue"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/store"
	"github.com/objectvault/api-services/requests/rpf/user"
	"github.com/objectvault/api-services/requests/rpf/utils"
	"github.com/objectvault/api-services/xjson"
	"github.com/objectvault/filter-parser/ast"
	"github.com/objectvault/filter-parser/builder"
	"github.com/objectvault/filter-parser/token"

	rpf "github.com/objectvault/goginrpf"
)

func groupAcceptOrganizationInvite(parent rpf.GINProcessor, u *orm.UserRegistry, i *orm.InvitationRegistry) *rpf.ProcessorGroup {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = parent

	group.Chain = rpf.ProcessChain{
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Session Store
			session := sessions.Default(c)

			// Remove Invitation ID from Session
			session.Delete("invitation-id")
		},
		invitation.DBGetInvitationFromRegistry,
		func(r rpf.GINProcessor, c *gin.Context) {
			inv := r.MustGet("invitation").(*orm.Invitation)
			r.SetLocal("object-id", inv.Object())
			r.SetLocal("org-id", inv.Object())
		},
		org.DBRegistryOrgFindByID,
		org.AssertOrgUnblocked,
		func(r rpf.GINProcessor, c *gin.Context) {
			inv := r.MustGet("invitation").(*orm.Invitation)

			// Do we have Invitation Roles Set?
			if !inv.IsRolesEmpty() { // YES
				r.SetLocal("register-roles", inv.Roles())
			}
		},
		object.DBObjectUserFindOrNil,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Does Registration Already Exist?
			if r.Has("registry-object-user") {
				u := r.Get("registry-object-user").(*orm.ObjectUserRegistry)
				if r.Has("register-roles") { // YES: Add New Roles (if any)
					u.AddRoles(r.MustGet("register-roles").([]uint32))
				}

				object.DBObjectUserFlush(r, c)
			} else { // NO: Register User with Org
				object.DBRegisterUserWithOrg(r, c)
			}
		},
		object.DBRegistryUserOrgFindOrNil,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Does Registration Already Exist?
			if !r.Has("registry-user-org") { // NO: Register Org with User
				object.DBRegisterOrgWithUser(r, c)
			}
		},
		// Update Invitation //
		invitation.DBInvitationAccepted,
		// Update Session
		session.SaveSession,
	}

	return group
}

func groupAcceptStoreInvite(parent rpf.GINProcessor, u *orm.UserRegistry, i *orm.InvitationRegistry) *rpf.ProcessorGroup {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = parent

	group.Chain = rpf.ProcessChain{
		// PROCESS JSON Body //
		session.AssertSessionRegistered, // Verify if Registered User Session
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Session Store
			session := sessions.Default(c)

			// Do we have a User Hash?
			r.SetLocal("hash", session.Get("user-hash"))
		},
		// Get Store
		func(r rpf.GINProcessor, c *gin.Context) {
			inv := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

			r.SetLocal("request-store", inv.Object())
		},
		store.DBStoreGetByID,
		// Verify Store Parent Organization Unlocked
		func(r rpf.GINProcessor, c *gin.Context) {
			store := r.MustGet("store").(*orm.Store)

			r.SetLocal("org-id", store.Organization())
		},
		org.DBRegistryOrgFindByID,
		org.AssertOrgUnblocked,
		// Verify Store Unlocked
		org.DBOrgStoreFind,
		store.AssertStoreUnblocked,
		// Get Store Key
		invitation.DBGetInvitationFromRegistry,
		func(r rpf.GINProcessor, c *gin.Context) {
			inv := r.MustGet("invitation").(*orm.Invitation)

			// Extract Key ID and Pick
			r.SetLocal("key-id", *inv.Key())
			r.SetLocal("key-pick", inv.KeyPick())
		},
		keys.DBGetKeyByID,
		keys.KeyExtractBytes,
		// Register User with Store
		func(r rpf.GINProcessor, c *gin.Context) {
			// Save Store Key
			r.SetLocal("store-key", r.MustGet("key-bytes"))

			// Do we have Invitation Roles Set?
			inv := r.MustGet("invitation").(*orm.Invitation)
			if !inv.IsRolesEmpty() { // YES
				r.SetLocal("register-roles", inv.Roles())
			}
		},
		store.DBRegisterUserWithExistingStore,
		store.DBRegisterStoreWithUser,
		// Update Invitation //
		invitation.DBInvitationAccepted,
		// Update Session
		session.SaveSession,
	}

	return group
}

func mixinCreateUserFromInvitation(c *rpf.ProcessChain) rpf.ProcessChain {
	// Create User from Invitation Request
	return append(*c,
		// GET JSON Body //
		shared.RequestExtractJSON,
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

			// Extract and Validate JSON Message
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

			// Did we have an Error Processing the Map?
			if vmap.Error != nil {
				fmt.Println(vmap.Error)
				fmt.Println(vmap.StringSrc())
				r.Abort(5998 /* TODO: ERROR [User Information Invalid] */, nil)
				return
			}
		},
		// Make Sure Invitation ORG Exists and is Active //
		func(r rpf.GINProcessor, c *gin.Context) {
			i := r.MustGet("invitation").(*orm.Invitation)
			r.SetLocal("org-id", i.Object())
		},
		org.DBRegistryOrgFindByID,
		org.AssertOrgUnblocked,
		// REGISTER User //
		// TODO: Check if User Alias Already Exists
		// TODO: Check if Email Already Exists
		user.DBInsertUser,
		user.DBRegisterUser,
		// REGISTER User With ORG //
		func(r rpf.GINProcessor, c *gin.Context) {
			i := r.MustGet("invitation").(*orm.Invitation)

			// Do we have Invitation Roles Set?
			if !i.IsRolesEmpty() { // YES
				r.SetLocal("register-roles", i.Roles())
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
	)
}

func doAcceptNotInUserSession(r rpf.GINProcessor, c *gin.Context) {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = r

	group.Chain = rpf.ProcessChain{
		// PROCESS INVITATION //
		invitation.DBGetInvitationFromRegistry,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Invitation
			inv := r.MustGet("invitation").(*orm.Invitation)

			// Find Invitee Email
			r.SetLocal("user-email", inv.InviteeEmail())
		},
		user.DBRegistryUserFindByEmailOrNil,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get User Registry Entry
			u := r.Get("registry-user")

			if u == nil { // User Does Not Exist
				// Create User from Invitation Request
				group.Chain = mixinCreateUserFromInvitation(&group.Chain)
			} else { // User Exists
				user := u.(*orm.UserRegistry)
				if user.IsActive() { // OK: User Exists and Is Active - Login to Accept
					group.Chain = append(group.Chain,
						// Update Session
						session.SaveSession,
						func(r rpf.GINProcessor, c *gin.Context) {
							r.Answer(1001 /* CODE: Login to Accept Code */)
						},
					)
				} else { // NOK: User Exist but is Blocked
					group.Chain = append(group.Chain,
						func(r rpf.GINProcessor, c *gin.Context) {
							// Get Session Store
							session := sessions.Default(c)

							// Remove Invitation ID from Session
							session.Delete("invitation-id")
						},
						session.SaveSession,
						func(r rpf.GINProcessor, c *gin.Context) {
							r.Abort(4998 /* CODE: Invalid Invitation Code */, nil)
						},
					)
				}
			}
		},
	}

	group.Run()
}

func doAcceptInUserSession(r rpf.GINProcessor, c *gin.Context) {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = r

	group.Chain = rpf.ProcessChain{
		// GET Session USER and Verify STATE //
		session.SessionUserToRegistry,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get User Registry Entry
			user := r.MustGet("registry-user").(*orm.UserRegistry)
			rif := session.IFCloseSessionOnBlockedUser(r, user)
			rif.Run()
		},
		// PROCESS INVITATION //
		func(r1 rpf.GINProcessor, c *gin.Context) {
			// Get Session User
			user := r1.MustGet("registry-user").(*orm.UserRegistry)

			// Get Invitation
			i := r1.MustGet("registry-invitation").(*orm.InvitationRegistry)

			// Is Session User Recipient?
			var g *rpf.ProcessorGroup
			if user.Email() == i.InviteeEmail() { // YES
				g = groupAcceptOrganizationInvite(r1, user, i)
			} else { // NO
				g = session.GroupCloseSessionWithError(r1, 4390)
			}

			g.Run()
		},
	}

	group.Run()
}

func doAcceptStoreInvite(r rpf.GINProcessor, c *gin.Context) {
	// Create Processing Group
	group := &rpf.ProcessorGroup{}
	group.Parent = r

	group.Chain = rpf.ProcessChain{
		// GET Session USER and Verify STATE //
		session.SessionUserToRegistry,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get User Registry Entry
			user := r.MustGet("registry-user").(*orm.UserRegistry)
			rif := session.IFCloseSessionOnBlockedUser(r, user)
			rif.Run()
		},
		// PROCESS INVITATION //
		func(r1 rpf.GINProcessor, c *gin.Context) {
			// Get Session User
			user := r1.MustGet("registry-user").(*orm.UserRegistry)

			// Get Invitation
			i := r1.MustGet("registry-invitation").(*orm.InvitationRegistry)

			// Is Session User Recipient?
			var g *rpf.ProcessorGroup
			if user.Email() == i.InviteeEmail() { // YES
				g = groupAcceptStoreInvite(r1, user, i)
			} else { // NO
				g = session.GroupCloseSessionWithError(r1, 4390)
			}

			g.Run()
		},
	}

	group.Run()
}

func InvitationAccept(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("POST.INVITATION", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		invitation.ExtractGINParameterUID,
		invitation.DBGetRegistryInvitationByUID,
		invitation.AssertInvitationActive, // Check if Invitation Still Active
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Invitation
			i := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

			// Get Session Store
			session := sessions.Default(c)

			// Do we have a User Session?
			id := session.Get("user-id")

			// Is Invitation to Store?
			oid := i.Object()
			if common.IsObjectOfType(oid, common.OTYPE_STORE) {

				if id == nil { // NO: Require Session by Invitee
					r.Abort(4301, nil)
					return
				}

				doAcceptStoreInvite(r, c)
			} else if common.IsObjectOfType(oid, common.OTYPE_ORG) {

				if id != nil { // YES
					doAcceptInUserSession(r, c)
				} else { // NO
					doAcceptNotInUserSession(r, c)
				}
			} else {
				r.Abort(4390, nil)
			}
		},
	}

	// Start Request Processing
	request.Run()
}

func InvitationDecline(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.INVITATION", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// GET Invitation //
		invitation.ExtractGINParameterUID,
		invitation.DBGetRegistryInvitationByUID,
		invitation.AssertInvitationActive, // Check if Invitation Still Active
		// UPDATE Invitation //
		invitation.DBInvitationDeclined,
	}

	// Start Request Processing
	request.Run()
}

func InvitationNoSessionInfo(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.INVITATION", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// GET Invitation //
		invitation.ExtractGINParameterUID,
		invitation.DBGetRegistryInvitationByUID,
		invitation.AssertInvitationActive, // Check if Invitation Still Active
		// Get Creator of Invite
		func(r rpf.GINProcessor, c *gin.Context) {
			i := r.Get("registry-invitation").(*orm.InvitationRegistry)
			r.SetLocal("user-id", i.Creator())
			r.SetLocal("user-email", i.InviteeEmail())
		},
		user.DBRegistryUserFindByID,
		user.AssertUserUnblocked,
		func(r rpf.GINProcessor, c *gin.Context) {
			r.SetLocal("invitation-creator", r.Get("registry-user"))
			r.Unset("registry-user")
		},
		user.DBRegistryUserFindByEmailOrNil,
		func(r rpf.GINProcessor, c *gin.Context) {
			if r.HasLocal("registry-user") {
				r.SetLocal("invitation-invitee", r.Get("registry-user"))
			}
		},
		// RESPONSE //
		invitation.ExportNoSessionRegistryInv,
	}

	// Start Request Processing
	request.Run()
}

func ListInvitesByObject(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.OBJECT.INVITES", c, 1000, shared.JSONResponse)

	// DYNAMIC Request
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameters
		object.ExtractGINParameterObjectID,
		// Dynamically Create Request
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Object ID
			oid := r.MustGet("request-object-id").(uint64)

			if common.IsObjectOfType(oid, common.OTYPE_ORG) {
				// Set Request Parameter as OID
				r.SetLocal("request-org", oid)

				// Treat as if Organization Request
				org.BaseValidateOrgRequest(request, func(o string) interface{} {
					// Required Roles : Organization Invite Access with List Function
					roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_INVITE, orm.FUNCTION_LIST)}

					if o == "roles" {
						return roles
					}

					return nil
				})
			} else if common.IsObjectOfType(oid, common.OTYPE_STORE) {
				// Set Request Parameter as OID
				r.SetLocal("request-store", oid)
				r.SetLocal("object-id", oid)

				// Treat as if Store Request
				store.BaseValidateStoreRequest(request, func(o string) interface{} {
					// Required Roles : Store Invite Access with List Function
					roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_INVITE, orm.FUNCTION_LIST)}

					if o == "roles" {
						return roles
					}

					return nil
				})
			} else {
				// TODO: errors.New("Invalid Object ID")
				r.Abort(5303, nil)
			}

			// Request Process //
			request.Append(
				// Extract Query Parameters //
				func(r rpf.GINProcessor, c *gin.Context) {
					g := shared.GroupExtractQueryConditions(r,
						func(n ast.Node) ast.Node {
							// Get Request Object ID
							oid := r.MustGet("request-object-id").(uint64)
							sid := strconv.FormatUint(oid, 10)

							// Basic Filter Limit Query to Invitations in State 0 for Object
							add := builder.ASTAND(
								builder.ASTEQ("object", builder.ASTValue(token.INT, sid)),
								builder.ASTEQ("state", builder.ASTValue(token.INT, "0")),
							)

							// Set Query Filter
							var f *ast.Filter
							if n == nil {
								// Set Filter to Object ID
								f = builder.ASTFilter(add)
							} else {
								// Add Object ID and(eq(object, request-object-id), ...existing...)
								f = n.(*ast.Filter)
								f.F = builder.ASTAND(add, f.F)
							}

							return f
						},
						func(e_to_o string) string {
							switch e_to_o {
							case "id":
								return "id_invite"
							case "uid":
								return "uid"
							case "object":
								return "id_object"
							case "creator":
								return "id_creator"
							case "invitee":
								return "invitee_email"
							case "state":
								return "state"
							}

							// Invalid Field
							return ""
						}, nil)

					g.Run()
					if !r.IsFinished() { // YES
						// Save Query Settings as Global
						g.LocalToGlobal("query-conditions")
					}
				},
				// Query System for List //
				invitation.DBRegistryInviteList,
				// Export Results //
				invitation.ExportRegistryInvList,
			)

			// Save Session
			session.AddinSaveSession(request, nil)
		},
	}

	// Start Request Processing
	request.Run()
}

func ListAllInvites(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.ALL.INVITES", c, 1000, shared.JSONResponse)

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

func GetObjectInvite(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.INVITE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameters
		object.ExtractGINParameterObjectID,
		// Basic Validation
		func(r rpf.GINProcessor, c *gin.Context) {
			// Required Roles : System Invite List Role
			roles := []uint32{orm.Role(orm.CATEGORY_SYSTEM|orm.SUBCATEGORY_INVITE, orm.FUNCTION_READ)}

			// Initialize Request
			org.GroupOrgRequestInitialize(r, common.SYSTEM_ORGANIZATION, roles, false).
				Run()
		},
		/* REQUEST VALIDATION */
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Abort(5999, nil)
		},
	}

	// Start Request Processing
	request.Run()
}

func ResendInvite(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.INVITE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{}

	// Session
	request.Append(
		// Make user we have an Active Session
		session.AssertUserSession,
	)

	// Get Invitation Information
	request.Append(
		// Extract Request Parameters
		invitation.ExtractGINParameterInvitationID,
		// Get Invitation by ID
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Object ID
			iid := r.MustGet("request-invite-id").(uint64)

			// Save as Request Object ID
			r.SetLocal("invitation-id", iid)
		},
		invitation.DBGetRegistryInvitationByID,
		invitation.AssertInvitationActive,
	)

	// Process Request
	request.Append(
		// Validate Session Information
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Invitation
			i := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

			// Get Request Object ID
			oid := i.Object()

			// Save as Request Object ID
			r.SetLocal("object-id", oid)

			if common.IsObjectOfType(oid, common.OTYPE_ORG) {
				// Set Request Parameter as OID
				r.SetLocal("request-org", oid)

				// Treat as if Organization Request
				org.BaseValidateOrgRequest(request, func(o string) interface{} {
					// Required Roles : Organization Invite Access with Delete Function
					roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_INVITE, orm.FUNCTION_CREATE)}

					if o == "roles" {
						return roles
					}

					return nil
				})
			} else if common.IsObjectOfType(oid, common.OTYPE_STORE) {
				// Set Request Parameter as OID
				r.SetLocal("request-store", oid)

				// Treat as if Store Request
				store.BaseValidateStoreRequest(request, func(o string) interface{} {
					// Required Roles : Store Invite Access with Delete Function
					roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_INVITE, orm.FUNCTION_CREATE)}

					if o == "roles" {
						return roles
					}

					return nil
				})
			} else {
				// TODO: errors.New("Invalid Object ID")
				r.Abort(5303, nil)
			}

			// Request Process //
			request.Append(
				func(r rpf.GINProcessor, c *gin.Context) {
					// Get Session Store
					session := sessions.Default(c)

					// Get Session User's Email and Name for Invitation
					r.SetLocal("from-user-email", session.Get("user-email"))
					r.SetLocal("from-user-name", session.Get("user-name"))

					// Message Queue
					r.SetLocal("queue", "action:start")
				},
				invitation.DBGetInvitationByID,
				queue.CreateInvitationMessage,
				queue.SendQueueMessage,
				// RESPONSE //
				invitation.ExportRegistryInv,
			)
		},
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}

func DeleteInvite(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.INVITE", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{}

	// Session
	request.Append(
		// Make user we have an Active Session
		session.AssertUserSession,
	)

	// Get Invitation Information
	request.Append(
		// Extract Request Parameters
		invitation.ExtractGINParameterInvitationID,
		// Get Invitation by ID
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Object ID
			iid := r.MustGet("request-invite-id").(uint64)

			// Save as Request Object ID
			r.SetLocal("invitation-id", iid)
		},
		invitation.DBGetRegistryInvitationByID,
	)

	// Get Invitation Information
	request.Append(
		// Dynamically Create Request
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Invitation
			i := r.MustGet("registry-invitation").(*orm.InvitationRegistry)

			// Save as Request Object ID
			oid := i.Object()
			r.SetLocal("object-id", oid)

			if common.IsObjectOfType(oid, common.OTYPE_ORG) {
				// Set Request Parameter as OID
				r.SetLocal("request-org", oid)

				// Treat as if Organization Request
				org.BaseValidateOrgRequest(request, func(o string) interface{} {
					// Required Roles : Organization Invite Access with Delete Function
					roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_INVITE, orm.FUNCTION_DELETE)}

					if o == "roles" {
						return roles
					}

					return nil
				})
			} else if common.IsObjectOfType(oid, common.OTYPE_STORE) {
				// Set Request Parameter as OID
				r.SetLocal("request-store", oid)

				// Treat as if Store Request
				store.BaseValidateStoreRequest(request, func(o string) interface{} {
					// Required Roles : Store Invite Access with Delete Function
					roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_INVITE, orm.FUNCTION_DELETE)}

					if o == "roles" {
						return roles
					}

					return nil
				})
			} else {
				r.Abort(4390, nil) // Invalid Invitation ID
			}

			// Revoke Invitation
			request.Append(
				invitation.DBInvitationRevoked,
			)

			// Save Session
			session.AddinSaveSession(request, nil)
		},
	)

	// Start Request Processing
	request.Run()
}
