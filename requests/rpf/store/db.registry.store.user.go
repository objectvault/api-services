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

import (
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/object"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func DBRegistryStoreUserList(r rpf.GINProcessor, c *gin.Context) {
	r.SetLocal("object-id", r.MustGet("store-id").(uint64))

	// Redirect to Object Registry
	object.DBRegistryObjectUsersList(r, c)
}

func DBGetRegistryStoreUser(r rpf.GINProcessor, c *gin.Context) {
	// Store User Registry Entry Already Exists?
	if r.Has("registry-store-user") { // YES: Do Nothing
		return
	}

	// Get Object Registry from Store ID
	r.SetLocal("object-id", r.MustGet("store-id").(uint64))
	object.DBRegistryObjectUserFind(r, c)
	if !r.Aborted() {
		// Save Entry
		r.SetLocal("registry-store-user", r.MustGet("registry-object-user"))
	}
}

func DBGetRegistryUserStore(r rpf.GINProcessor, c *gin.Context) {
	// Store User Registry Entry Already Exists?
	if r.Has("registry-user-store") { // YES: Do Nothing
		return
	}

	// Get Object Registry from Store ID
	r.SetLocal("object-id", r.MustGet("store-id").(uint64))
	object.DBRegistryUserObjFind(r, c)
	if !r.Aborted() {
		// Save Entry
		r.SetLocal("registry-user-store", r.MustGet("registry-user-object"))
	}
}

func DBRegisterStoreWithUser(r rpf.GINProcessor, c *gin.Context) {
	// Get Store Global ID
	storeID := r.MustGet("store-id").(uint64)

	// Get User Global ID
	userID := r.MustGet("user-id").(uint64)

	// Get Store Entry
	store := r.MustGet("store").(*orm.Store)

	// Create Registry Entry
	o := &orm.UserObjectRegistry{}
	o.SetKey(userID, storeID)
	o.SetAlias(store.Alias())

	// Flush Registry Entry
	r.SetLocal("registry-user-object", o)
	object.DBRegistryUserObjFlush(r, c)
	if !r.Aborted() {
		// Save Entry
		r.SetLocal("registry-user-store", o)
	}
}

func DBRegisterUserWithNewStore(r rpf.GINProcessor, c *gin.Context) {
	// Get Store Global ID
	storeID := r.MustGet("store-id").(uint64)

	// Get User Informtaion
	user := r.MustGet("registry-user").(*orm.UserRegistry)
	userHash := r.MustGet("hash").(string)

	// Create Registry Entry
	o := &orm.ObjectUserRegistry{}
	o.SetKey(storeID, user.ID())
	o.SetUserName(user.UserName())

	// Are specific roles to be set?
	if r.Has("register-roles") { // YES
		roles := r.Get("register-roles").([]uint32)
		o.AddRoles(roles)
	}

	// Is User Store Admin?
	if r.Has("register-as-admin") { // YES
		o.SetState(orm.STATE_SYSTEM)
	}

	// Create Store Key
	err := o.CreateStoreKey(userHash)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Flush Changes
	r.SetLocal("registry-object-user", o)
	object.DBRegistryObjectUserFlush(r, c)
	if !r.Aborted() {
		// Save Entry
		r.SetLocal("registry-store-user", o)
	}
}

func DBRegisterUserWithExistingStore(r rpf.GINProcessor, c *gin.Context) {
	// Get Store Global ID
	storeID := r.MustGet("store-id").(uint64)
	storeKey := r.MustGet("store-key").([]byte)

	// Get User Information
	user := r.MustGet("registry-user").(*orm.UserRegistry)
	userHash := r.MustGet("hash").(string)

	// Create Registry Entry
	o := &orm.ObjectUserRegistry{}
	o.SetKey(storeID, user.ID())
	o.SetUserName(user.UserName())

	// Are specific roles to be set?
	if r.Has("register-roles") { // YES
		roles := r.MustGet("register-roles").([]uint32)
		o.AddRoles(roles)
	}

	// Is User Store Admin?
	if r.Has("register-as-admin") { // YES
		o.SetState(orm.STATE_SYSTEM)
	}

	// Set Store Key
	err := o.SetStoreKey(userHash, storeKey)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Flush Changes
	r.SetLocal("registry-object-user", o)
	object.DBRegistryObjectUserFlush(r, c)
	if !r.Aborted() {
		// Save Entry
		r.SetLocal("registry-store-user", o)
	}
}

func DBRegistryStoreUserUpdate(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entry
	r.SetLocal("registry-object-user", r.MustGet("registry-store-user"))
	object.DBRegistryObjectUserFlush(r, c)
}

func DBRegistryUpdateFromUser(r rpf.GINProcessor, c *gin.Context) {
	// Get User
	u := r.MustGet("user").(*orm.User)

	// Do we need to Update the Registry?
	if u.UpdateRegistry() { // YES
		// Get Registry Entry
		o := r.MustGet("registry-object-user").(*orm.ObjectUserRegistry)

		// Update Registry Fields
		o.SetUserName(u.UserName())
		object.DBRegistryObjectUserFlush(r, c)
	}
}
