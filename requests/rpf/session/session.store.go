// cSpell:ignore goginrpf, gonic, paulo ferreira
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
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	rpf "github.com/objectvault/goginrpf"

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
)

// Utilityu: Store ID to Store Session Key
func CreateStoreKey(id uint64) string {
	skey := fmt.Sprintf("_s:%d", id)
	return skey
}

/* TODO: PROBLEM: Currently if User Changes Password
 * All of the Stores Keys Have to be unsealed and re-sealed with
 * the new password, since it's used as the decryption key for store
 * passwords
 */
func SessionStoreOpen(r rpf.GINProcessor, c *gin.Context) {
	// Get Session Store
	session := sessions.Default(c)

	// Get Registry Object
	rus := r.MustGet("registry-store-user").(*orm.ObjectUserRegistry)

	// Create Session Key
	skey := CreateStoreKey(rus.Object())

	// Store Session Object
	var ss *common.StoreSession
	var e error

	// Have Existing Store Session?
	iss := session.Get(skey)
	if iss != nil { // YES: Import it and Validate

		// TODO: Make the Store Open Configurable
		ss, e = common.ImportStoreSession(iss.(string), 5)
		if e != nil || ss.IsExpired() { // Not Valid: Clear Session
			ss = nil
		} else { // Valid: Extend Session
			ss.Extend(0)
		}
	}

	// Have Existing Session?
	if ss == nil { // NO: Create Session
		// Validate User Credentials
		hash := r.MustGet("user-credentials").([]byte)
		key, e := rus.StoreKeyBytes(hash)
		if e != nil {
			r.Abort(3998 /* TODO: Error Code - Invalid Credentials */, nil)
			return
		}

		// TODO: Make the Store Open Configurable
		ss, e = common.NewStoreSession(rus.Object(), key, 5)
		if e != nil || !ss.IsValid() {
			r.Abort(5010 /* TODO: Failed to Create Store Session */, nil)
			return
		}
	}

	r.SetLocal("store-session", ss)
}

func ExtendStoreSession(r rpf.GINProcessor, c *gin.Context) {
	// Get Session Store
	session := sessions.Default(c)

	// Store ID
	sid := r.MustGet("request-store").(uint64)
	skey := CreateStoreKey(sid)

	// Have Existing Store Session?
	iss := session.Get(skey)
	if iss == nil {
		r.Abort(5010 /* TODO: No Existing Store Session */, nil)
		return
	}

	// TODO: Make the Store Open Configurable
	ss, e := common.ImportStoreSession(iss.(string), 5)
	if e != nil || ss.IsExpired() { // Not Valid: Clear Session
		r.Abort(5010 /* TODO: Failed to Create Store Session */, nil)
		return
	}

	// Valid: Extend Session
	ss.Extend(0)

	// Save Session and Key
	r.SetLocal("store-session", ss)
	r.SetLocal("store-key", ss.Key())
}

func SessionStoreSave(r rpf.GINProcessor, c *gin.Context) {
	// Get Session Store
	session := sessions.Default(c)

	// Get Store Session
	ss := r.MustGet("store-session").(*common.StoreSession)

	// Create Link between Store ID and Store Key
	svalue, e := ss.Export()
	if e != nil {
		r.Abort(5010 /* TODO: Failed to Create Store Session */, nil)
		return
	}

	// Update Session
	skey := CreateStoreKey(ss.Store())
	session.Set(skey, svalue)
}

func SessionStoreClose(r rpf.GINProcessor, c *gin.Context) {
	// Get Session Store
	session := sessions.Default(c)

	// Store ID
	sid := r.MustGet("request-store").(uint64)
	skey := CreateStoreKey(sid)

	// Have Existing Store Session?
	iss := session.Get(skey)
	if iss != nil { // YES: Import it and Validate
		session.Delete(skey)
	}
}
