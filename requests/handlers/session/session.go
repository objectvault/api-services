// cSpell:ignore ginrpf, gonic, paulo ferreira
package session

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
	"log"

	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/user"
	"github.com/objectvault/api-services/requests/rpf/utils"
	"github.com/objectvault/api-services/xjson"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func RPFAddCorsHeaders(r rpf.GINProcessor, c *gin.Context) {
	// Allow Local Host (DEBUG)
	c.Header("Access-Control-Allow-Origin", "http://localhost:5000")
	// Allow Session Cookies
	c.Header("Access-Control-Allow-Credentials", "true")
	// Allow all Methods
	c.Header("Access-Control-Allow-Methods", "*")
}

func Hello(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.SESSION", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Set CORS Headers
		//		RPFAddCorsHeaders,
		// IF Chain
		func(r rpf.GINProcessor, c *gin.Context) {
			rif := rpf.NestedIF(r,
				func(r rpf.ProcessorIF, c *gin.Context) {
					// Get Session Store
					session := sessions.Default(c)

					// Do we have User ID?
					id := session.Get("user-id")
					if id == nil { // NO:
						r.ContinueFalse()
					} else {
						r.ContinueTrue()
					}
				},
				func(r rpf.ProcessorIF, c *gin.Context) {
					// Create Processing Group
					group := &rpf.ProcessorGroup{}
					group.Parent = &r
					group.Chain = rpf.ProcessChain{
						// PREPARE RESPONSE //
						session.SessionUserToRegistry,
						user.DBGetUserByID, // Find User by Global ID
						user.ExportUserSession,
					}

					group.Run()
				},
				func(r rpf.ProcessorIF, c *gin.Context) {
					// Get Session Store
					session := sessions.Default(c)

					// Did we Create a Session?
					err := session.Save()
					if err == nil { // YES: Okay
						r.Answer(1001)
					} else { // NO: Return Error
						log.Printf("[Action GET.SESSION] ERROR! %s\n", err)
						r.Abort(5000, nil)
					}
				})

			rif.Run()
		},
	}

	// Start Request Processing
	request.Run()
}

func Login(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("POST.SESSION", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Set CORS Headers
		//		RPFAddCorsHeaders,
		// REQUEST Validation - GIN Parameters //
		utils.RPFReadyVFields,
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Fields Error Message Map
			fields := r.Get("v_fields").(map[string]string)

			// Initial Post Parameter Tests
			id, message := utils.ValidateGinParameter(c, "id", true, true, false)
			if message != "" {
				fields["id"] = message
				return
			}

			iid, message := utils.ValidateUserReference(id)
			if message != "" {
				fields["id"] = message
				return
			}

			r.Set("user", iid)
		},
		func(r rpf.GINProcessor, c *gin.Context) {
			utils.RPFTestVFields(3100, r, c)
		},
		// PROCESS JSON Body //
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
				r.Set("hash", v.(string))
				return nil
			})

			// OPTIONAL: Register Password Hash
			vmap.Optional("reset", nil, xjson.F_xToBoolean, false, func(v interface{}) error {
				r.SetLocal("session-reset", v.(bool))
				return nil
			})

			// OPTIONAL: Register Password Hash
			vmap.Optional("register", nil, xjson.F_xToBoolean, false, func(v interface{}) error {
				r.SetLocal("session-register", v.(bool))
				return nil
			})

			// Did we have an Error Processing the Map?
			if vmap.Error != nil {
				fmt.Println(vmap.Error)
				fmt.Println(vmap.StringSrc())
				r.Abort(3200, nil)
				return
			}
		},
		user.DBRegistryUserFind,  // Get User Registry
		session.CloseUserSession, // Reset Session if Required
		// Verify User State //
		user.AssertUserActive,  // See if the account active
		user.AssertUserBlocked, // See if Account Blocked by System Admin
		user.AssertCredentials, // See if User Password Correct
		// Verify User Password //
		/*
			func(r rpf.GINProcessor, c *gin.Context) {
				// Use Only User Global ID
				user := r.Get("registry-user").(*orm.UserRegistry)
				r.Set("user-id", user.ID())
			},
		*/
		// TODO Set a Time Limit on the Cookie (1 hour)
		session.OpenUserSession, // Open User Session
		user.ExportUserSession,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

func Logout(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.SESSION", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		session.CloseUserSession, // Clear Session Information
		session.SaveSession,      // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}
