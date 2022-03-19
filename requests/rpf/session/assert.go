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
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AssertUserSession(r rpf.GINProcessor, c *gin.Context) {
	// Get Session Store
	session := sessions.Default(c)

	// Do we have a User Session?
	id := session.Get("user-id")
	if id == nil { // NO: Exit
		r.Abort(3000, nil)
		return
	}
	// ELSE: User Logged In (Continue)
}

func AssertNoUserSession(r rpf.GINProcessor, c *gin.Context) {
	// Get Session Store
	session := sessions.Default(c)

	// Do we have a User Session?
	id := session.Get("user-id")
	if id != nil { // YES: Exit
		r.Abort(3003, nil)
		return
	}
	// ELSE: No User Logged In (Continue)
}

func AssertSessionRegistered(r rpf.GINProcessor, c *gin.Context) {
	// Get Session Store
	session := sessions.Default(c)

	// Do we have a User Hash?
	id := session.Get("user-hash")
	if id == nil { // NO: Exit
		r.Abort(3000, nil)
		return
	}
	// ELSE: User Logged In (Continue)
}

func AssertSystemAdmin(r rpf.GINProcessor, c *gin.Context) {
	// Get Session Store
	session := sessions.Default(c)

	// Do we have a Admin Session?
	id := session.Get("user-id")
	if id == nil || id != 0 { // NO: Not Admin Session - Abort
		r.Abort(3998 /* TODO: ERROR MESSAGE */, nil)
		return
	}
	// ELSE: Admnin User Logged In (Continue)
}

func AssertNotSystemAdmin(r rpf.GINProcessor, c *gin.Context) {
	// Get Session Store
	session := sessions.Default(c)

	// Do we have a Admin Session?
	id := session.Get("user-id")
	if id != nil && id == 0 { // NO: Admin Logged In - Abort
		r.Abort(3998 /* TODO: ERROR MESSAGE */, nil)
		return
	}
	// ELSE: Admnin User Logged In (Continue)
}

func AssertNotSelf(r rpf.GINProcessor, c *gin.Context) {
	// Get Session Store
	session := sessions.Default(c)

	// Do we have a User Session?
	sid := session.Get("user-id")
	if sid == nil { // NO: Exit
		r.Abort(3000, nil)
		return
	}

	// Is Request User === Session User?
	uid := r.MustGet("user-id").(uint64)
	if sid == uid { // YES: Action not permitted on self
		r.Abort(4004, nil)
		return
	}
	// ELSE: Not SELF
}
