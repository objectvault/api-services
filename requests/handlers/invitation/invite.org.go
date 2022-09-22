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

// cSpell:ignore vmap

import (
	"errors"
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	rpf "github.com/objectvault/goginrpf"

	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/invitation"
	"github.com/objectvault/api-services/requests/rpf/object"
	"github.com/objectvault/api-services/requests/rpf/org"
	"github.com/objectvault/api-services/requests/rpf/queue"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/utils"
	"github.com/objectvault/api-services/xjson"
)

// ORGANIZATION: Invitation Management //

func CreateOrgInvitation(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("POST.ORG.INVITATION", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{}

	// Required Roles : Organization Invite Access Role with Create Function
	roles := []uint32{orm.Role(orm.CATEGORY_ORG|orm.SUBCATEGORY_INVITE, orm.FUNCTION_CREATE)}

	// Basic Request Validate
	org.AddinGroupValidateOrgRequest(request, func(o string) interface{} {
		if o == "roles" {
			return roles
		}

		return nil
	})

	// Create Invitation
	request.Append(
		// EXTRACT : Invitation
		shared.RequestExtractJSON,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Extract and Validate JSON MEssage
			m := r.MustGet("request-json").(xjson.T_xMap)
			vmap := xjson.S_xJSONMap{Source: m}

			// Get User Identifier
			creator := r.MustGet("user-id").(uint64)

			// Get Organization
			org := r.MustGet("registry-org").(*orm.OrgRegistry)

			// Create Invitation
			i := &orm.Invitation{}
			i.SetCreator(creator)
			i.SetObject(org.ID())

			// Invitee EMAIL
			vmap.Required("invitee", nil, func(v interface{}) (interface{}, error) {
				v, e := xjson.F_xToTrimmedString(v)
				if e != nil {
					return nil, e
				}

				// Extract Invited Email
				s := v.(string)
				if !utils.IsValidEmail(s) {
					return nil, errors.New("Value is not a valid email address")
				}

				// TEST: Invitee != Invited //
				// Get Session Store
				session := sessions.Default(c)

				// Do we have a User Session Email?
				email := session.Get("user-email")
				if email == nil { // NO
					return nil, errors.New("System Error: Invalid Session")
				}

				if email.(string) == s {
					return nil, errors.New("Invalid Invite: Invitee is equal to Invited")
				}
				return s, nil
			}, func(v interface{}) error {
				i.SetInviteeEmail(v.(string))
				return nil
			})

			// OPTIONAL: Organizational Roles to Set if Invitation Accepted
			vmap.Optional("roles", nil, xjson.F_xToTrimmedString, nil, func(v interface{}) error {
				var s string

				// Invitation Roles?
				if v == nil { // NO: Use Minimum Roles for Organization Access
					s = "33882113"
				} else { // YES
					s = v.(string)
				}

				// Set Roles
				i.RolesFromCSV(s)
				return nil
			})

			// OPTIONAL: Message to Send as part of the email
			vmap.Optional("message", nil, xjson.F_xToString, nil, func(v interface{}) error {
				if v != nil {
					i.SetMessage(v.(string))
				}
				return nil
			})

			// OPTIONAL: Invitation Experiation Period
			vmap.Optional("expiry_in_days", nil, xjson.F_xToUint64, uint64(3), func(v interface{}) error {
				ui := v.(uint64)
				i.SetExpiresIn(uint16(ui))
				return nil
			})

			// Did we have an Error Processing the Map?
			if vmap.Error != nil {
				fmt.Println(vmap.Error)
				fmt.Println(vmap.StringSrc())
				r.Abort(5202, nil)
				return
			}

			// Did we have the requirements for an Invitation
			if !i.IsValid() { // NO
				r.Abort(5202, nil)
				return
			}

			// Save Invitation
			r.SetLocal("invitation", i)
		},
		// TODO: Make sure invitee is not already part of organization
		// Validations
		invitation.AssertNoPendingInvitation,
	)

	// Test if Object User Registration Already Exists
	object.AddinNoExistingUserRegistration(request, nil)

	// Register Invitation
	request.Append(
		// Register Invitation
		invitation.DBInsertInvitation,
		invitation.DBRegisterInvitation,
		// IMPORTANT: As long as the invitation is created (but not published to the queue) the handler passes
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Session Store
			session := sessions.Default(c)

			// Get Session User's Email and Name for Invitation
			r.SetLocal("from-user-email", session.Get("user-email"))
			r.SetLocal("from-user-name", session.Get("user-name"))

			// Message Queue
			r.SetLocal("queue", "action:start")
		},
		queue.CreateInvitationMessage,
		queue.SendQueueMessage,
		// RESPONSE //
		invitation.ExportRegistryInv,
	)

	// Save Session
	session.AddinSaveSession(request, nil)

	// Start Request Processing
	request.Run()
}
