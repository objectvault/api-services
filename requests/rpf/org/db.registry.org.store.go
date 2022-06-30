// cSpell:ignore goginrpf, gonic, paulo ferreira
package org

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
	"github.com/objectvault/api-services/orm/query"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func DBRegistryOrgStoreList(r rpf.GINProcessor, c *gin.Context) {
	// Get User Identifier
	org := r.MustGet("org-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Org Shard
	db, err := dbm.Connect(org)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// List Registered Org Stores
	q := r.MustGet("query-conditions").(*query.QueryConditions)
	stores, err := orm.QueryRegisteredStores(db, org, q, true)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save List
	r.Set("registry-stores", stores)
}

func DBRegistryOrgStoreFind(r rpf.GINProcessor, c *gin.Context) {
	// GetSearch Parameters
	org := r.MustGet("org-id").(uint64)
	store := r.MustGet("request-store")

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Org Store Registry
	db, err := dbm.Connect(org)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Org based on ID Type
	entry := &orm.OrgStoreRegistry{}
	err = entry.Find(db, org, store)

	// Failed Retrieve?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Did we find the Store?
	if !entry.IsValid() { // NO: Store does not exist
		r.Abort(4998 /* TODO: Error [Store does not exist] */, nil)
		return
	}

	// Save Org Registry
	r.SetLocal("registry-store", entry)
}

func DBRegisterStoreWithOrg(r rpf.GINProcessor, c *gin.Context) {
	// Get Store
	store := r.MustGet("store").(*orm.Store)

	// Get Store Global ID
	store_id := r.MustGet("store-id").(uint64)

	// Get Parent Organization ID
	org_id := r.MustGet("org-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Registry ORG <--> STORE is On Organization Shard
	db, err := dbm.Connect(org_id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Create ORG Entry
	e := &orm.OrgStoreRegistry{}
	e.SetKey(org_id, store_id)
	e.SetStoreAlias(store.Alias())

	// Save
	err = e.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Entry
	r.SetLocal("registry-store", e)
}

func DBRegistryDeleteStore(r rpf.GINProcessor, c *gin.Context) {
	// Get Store
	// registry := r.MustGet("registry-store").(*orm.OrgStoreRegistry)

	// TODO : Implement

	/* IDEA:
	 * - Mark Store Registration as Deleted
	 * - Queue Action to Delete Store from System (With All Associated Objects)
	 */
	r.Abort(5999, nil)
}

func DBRegistryOrgStoreUpdate(r rpf.GINProcessor, c *gin.Context) {
	// Get Store Registry Entry
	e := r.MustGet("registry-store").(*orm.OrgStoreRegistry)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Connect to Organization Registry Shard
	db, err := dbm.Connect(e.Organization())
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save
	err = e.Flush(db, false)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}

func DBRegistryUpdateFromStore(r rpf.GINProcessor, c *gin.Context) {
	// Get Registry Entry
	e := r.MustGet("registry-store").(*orm.OrgStoreRegistry)

	// Get Store
	store := r.MustGet("store").(*orm.Store)

	// Do we need to Update the Store Registry?
	if store.UpdateRegistry() { // YES
		// Get Database Connection Manager
		dbm := c.MustGet("dbm").(*orm.DBSessionManager)

		// Connect to Registry Shard
		db, err := dbm.Connect(e.Organization())
		if err != nil { // YES: Database Error
			r.Abort(5100, nil)
			return
		}

		// Update Registry Fields
		e.SetStoreAlias(store.Alias())

		// Flush Registry
		err = e.Flush(db, false)
		if err != nil { // YES: Database Error
			r.Abort(5100, nil)
			return
		}
	}
}
