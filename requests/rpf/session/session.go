// cSpell:ignore gonic, orgs, paulo, ferreira
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
	"log"
	"time"

	"github.com/objectvault/api-services/common"

	orm "github.com/objectvault/api-services/orm"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetUserFromSession(r rpf.GINProcessor, c *gin.Context) {
	// Get Session Store
	session := sessions.Default(c)

	// Extract User Information from Session
	id := session.Get("user-id").(uint64)
	username := session.Get("user-username").(string)
	email := session.Get("user-email").(string)
	name := session.Get("user-name").(string)

	// Create Basic User
	user := orm.User{}

	user.SetID(common.LocalIDFromID(id))
	user.SetUserName(username)
	user.SetEmail(email)
	user.SetName(name)

	// Set Context User
	r.SetLocal("user-id", id)
	r.SetLocal("user", &user)
}

func OpenUserSession(r rpf.GINProcessor, c *gin.Context) {
	// Get Request User
	user := r.MustGet("registry-user").(*orm.UserRegistry)

	// Get Session Store
	session := sessions.Default(c)

	// Save User Information to Session
	session.Set("user-id", user.ID())
	session.Set("user-username", user.UserName())
	session.Set("user-email", user.Email())
	session.Set("user-name", user.Name())

	// Have Conditions for Session Regitration?
	if r.Has("session-register") && r.Has("hash") { // YES

		// Session Register Requested?
		register := r.MustGet("session-register").(bool)
		if register { // YES: Register User Hash
			// TODO: Set Timeout for Registered Session
			session.Set("user-hash", r.MustGet("hash").(string))
		}
	}
}

func CloseUserSession(r rpf.GINProcessor, c *gin.Context) {
	// Get Session Store
	session := sessions.Default(c)

	// Clear Session Flag
	clear := false

	// Do we have a Session
	if session.Get("user-id") != nil { // YES
		// Default : Clear Session
		clear = true

		// Do we have a context user?
		if r.Has("registry-user") { // YES
			sid := session.Get("user-id")
			user := r.Get("registry-user").(*orm.UserRegistry)

			// Is Context User same as Session User?
			if sid == user.ID() && r.Has("session-reset") { // YES
				// IS Session Reset Requested?
				clear = r.MustGet("session-reset").(bool)
			}
		}
	}
	// ELSE: No User Logged IN

	// Clear Existing Session?
	if clear { // YES
		session.Clear()
	}
}

func SaveSession(r rpf.GINProcessor, c *gin.Context) {
	// Get Session Store
	session := sessions.Default(c)

	// Update Access Time (Also Forces Cookie Update)
	session.Set("timestamp", time.Now().Unix())

	// Did Initiate a Session?
	err := session.Save()
	if err != nil { // NO: Abort
		log.Printf("[SaveSession] ERROR! %s\n", err)
		r.Abort(5000, nil)
		return
	}
}
