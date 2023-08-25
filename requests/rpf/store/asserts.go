// cSpell:ignore goginrpf, gonic, paulo ferreira
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
	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/object"
	"github.com/objectvault/api-services/requests/rpf/session"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// STORE //

func AssertStoreUnblocked(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Organization
	entry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)

	// Is the Store Blocked?
	if entry.IsBlocked() { // YES: Abort
		r.Abort(4203, nil) // TODO: Specific Error
		return
	}
}

func AssertStoreNotDeleted(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Organization
	entry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)

	// Is the Store Blocked?
	if entry.IsDeleted() { // YES: Abort
		r.Abort(4203, nil) // TODO: Specific Error
		return
	}
}

func AssertStoreOpen(r rpf.GINProcessor, c *gin.Context) {
	// Store ID
	id := r.MustGet("request-store").(uint64)

	// Get Session Store
	s := sessions.Default(c)

	// Get Store Key from Session
	skey := session.CreateStoreKey(id)
	key := s.Get(skey)

	// Does Store Key Exist in Session?
	if key == nil { // NO
		r.Abort(4202, nil)
		return
	}

	// TODO: Make the Store Open Configurable
	ss, e := common.ImportStoreSession(key.(string), 5)
	if e != nil {
		r.Abort(5010 /* TODO: Failed to Create Store Session */, nil)
		return
	}

	if ss.IsExpired() {
		s.Delete(skey)
		r.Abort(5010 /* TODO: Using Expired Store Session */, nil)
		return
	}

	r.SetLocal("store-key", ss.Key())
}

// REGISTRY STORE <--> USER  //

func AssertStoreUserUnblocked(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entry
	r.SetLocal("registry-object-user", r.MustGet("registry-store-user"))
	object.AssertObjectUserUnblocked(r, c)
}

func AssertUserHasAllRolesInStore(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entry
	r.SetLocal("registry-object-user", r.MustGet("registry-store-user"))
	object.AssertUserHasAllRolesInObject(r, c)
}

func AssertUserHasOneRoleInStore(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entry
	r.SetLocal("registry-object-user", r.MustGet("registry-store-user"))
	object.AssertUserHasOneRoleInObject(r, c)
}

// STORE <--> OBJECTS //

func AssertFolderObject(r rpf.GINProcessor, c *gin.Context) {
	obj := r.MustGet("store-object").(*orm.StoreObject)

	// Is Folder Object
	if obj.Type() != 0 { // NO: Abort
		r.Abort(4998 /* TODO: Error Code : Parent is Not Folder Object */, nil)
		return
	}
}
