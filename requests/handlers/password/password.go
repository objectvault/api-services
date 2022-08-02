/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package password

// cSpell:ignore pkgaction, pkgrequest, vmap, ormaction, ormrequest, xjson
import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	ormrequest "github.com/objectvault/api-services/orm/request"
	pkgaction "github.com/objectvault/api-services/requests/rpf/action"
	"github.com/objectvault/api-services/requests/rpf/queue"
	pkgrequest "github.com/objectvault/api-services/requests/rpf/request"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/user"
	"github.com/objectvault/api-services/requests/rpf/utils"
	"github.com/objectvault/api-services/xjson"

	rpf "github.com/objectvault/goginrpf"
)

func Reset(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("POST.PASSWORD", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Make sure we don't have an active session
		session.AssertNoUserSession,
		// Extract Route Parameter 'guid'
		shared.ExtractGINParameterGUID,
		// REQUEST Validation - JSON Body //
		shared.RequestExtractJSON,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Extract and Validate JSON Message
			m := r.MustGet("request-json").(xjson.T_xMap)
			vmap := xjson.S_xJSONMap{Source: m}

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
				r.Set("request-hash", v.(string))
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
		/* PROCESS:
		 * - Get Password Change Request by GUID
		 * - Assert Not Expired
		 * - Get User using email from Password Request
		 * - Assert User not Block or Deleted
		 * - Register Action Message with Queue Manager to Remove User from All Stores
		 * - Modify User Password
		 */
		/// RETRIEVE REQUEST //
		pkgrequest.DBGetRegistryRequestByGUID,
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Set("request-type", "password:reset")
		},
		pkgrequest.AssertRequestRegOfType,
		pkgrequest.AssertRequestRegActive,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request - Reference Object (i.e. the requesting user)
			rr := r.MustGet("registry-request").(*ormrequest.RequestRegistry)
			r.Set("user-id", rr.Object())
		},
		user.DBRegistryUserFindByID,
		user.AssertUserActive,
		user.DBGetUserByID,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get User Object
			user := r.MustGet("user").(*orm.User)
			userReg := r.MustGet("registry-user").(*orm.UserRegistry)

			// New Password Hash
			hash := r.MustGet("request-hash").(string)

			// Update Hash
			user.ResetHash(hash)
			userReg.UpdateRegistry(user)
		},
		user.DBUserUpdate,
		user.DBRegistryUserUpdate,
		func(r rpf.GINProcessor, c *gin.Context) {
			rr := r.MustGet("registry-request").(*ormrequest.RequestRegistry)
			rr.SetState(ormrequest.STATE_CLOSED)
		},
		pkgrequest.DBRegistryRequestUpdate,
	}

	// Start Request Processing
	request.Run()
}

func Recover(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.PASSWORD", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Make sure we don't have an active session
		session.AssertNoUserSession,
		// Extract Route Parameter 'email'
		shared.ExtractGINParameterEmail,
		// Does User Exist?
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Set("user-email", r.MustGet("request-email"))
		},
		user.DBRegistryUserFindByEmailOrNil,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request User
			u := r.Get("registry-user")

			// Does the user exist?
			if u == nil { // NO: Abort Processing
				r.Answer(1000)
			} else { // YES: Test if Marked for Deletion or Blocked
				user := u.(*orm.UserRegistry)
				if user.States.HasAnyStates(orm.STATE_BLOCKED | orm.STATE_DELETE) { // YES: Can't Change Password
					r.Answer(1099)
				}
			}
		},
		func(r rpf.GINProcessor, c *gin.Context) {
			rif := rpf.NestedIF(r,
				func(r rpf.ProcessorIF, c *gin.Context) {
					// Get Request User
					u := r.MustGet("registry-user").(*orm.UserRegistry)
					id := u.ID()

					// Get Database Connection Manager
					dbm := c.MustGet("dbm").(*orm.DBSessionManager)

					// Get Connection to User Registry
					db, err := dbm.ConnectTo(0, 0)
					if err != nil { // YES: Database Error
						r.Abort(5100, nil)
						return
					}

					// Expire Request before Counting
					_, err = ormrequest.ExpiryRequestsByType(db, "password:reset", &id)
					if err != nil { // YES
						r.Abort(5100, nil)
						return
					}

					// Count number of (NOT EXPIRED) Requests
					count, err := ormrequest.CountRequestsByType(db, "password:reset", true, &id)
					if err != nil { // YES
						r.Abort(5100, nil)
						return
					}

					// Do we already have a request?
					if count > 0 { // YES: Resend Email (Re-Create new Action Based on Request)
						r.ContinueTrue()
					} else { // NO: Create Reset Request
						r.ContinueFalse()
					}
				},
				func(r rpf.ProcessorIF, c *gin.Context) {
					// Create Processing Group
					group := &rpf.ProcessorGroup{}
					group.Parent = &r
					group.Chain = rpf.ProcessChain{
						user.DBGetUserFromRegistry,
						func(r rpf.GINProcessor, c *gin.Context) {
							// Get Request User
							u := r.MustGet("registry-user").(*orm.UserRegistry)
							id := u.ID()

							// Get Database Connection Manager
							dbm := c.MustGet("dbm").(*orm.DBSessionManager)

							// Get Connection to User Registry
							db, err := dbm.ConnectTo(0, 0)
							if err != nil { // YES: Database Error
								r.Abort(5100, nil)
								return
							}

							// Get Last Password Request for User
							reg, err := ormrequest.GetLastActiveRegistryByType(db, "password:reset", &id)
							if err != nil { // YES
								r.Abort(5100, nil)
								return
							}

							r.Set("registry-request", reg)
						},
						pkgrequest.DBGetRequestFromRegistry,
					}

					group.Run()
				}, // Use Request to Send Email
				func(r rpf.ProcessorIF, c *gin.Context) {
					// Create Processing Group
					group := &rpf.ProcessorGroup{}
					group.Parent = &r
					group.Chain = rpf.ProcessChain{
						user.DBGetUserFromRegistry,
						func(r rpf.GINProcessor, c *gin.Context) {
							// Get Request User
							our := r.MustGet("registry-user").(*orm.UserRegistry)

							or := ormrequest.NewRequest("password:reset", common.SYSTEM_ADMINISTRATOR)
							or.SetObject(our.ID())
							or.SetExpiresIn(1)

							// Save Request
							r.Set("request", or)
						},
						pkgrequest.DBInsertRequest,
						pkgrequest.DBRegisterRequest,
					}

					group.Run()
				},
			)

			rif.Run()
		},
		pkgaction.ActionCreatePasswordReset,
		pkgaction.DBRegisterAction,
		queue.CreateActionMessage,
		queue.QueueActionMessage,
		pkgaction.DBMarkActionQueued,
		queue.SendQueueMessage,
	}

	// Start Request Processing
	request.Run()
}
