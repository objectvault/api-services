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

func DBRegistryOrgList(r rpf.GINProcessor, c *gin.Context) {
	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to User Registry (Always in Group 0: Shard 0)
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// TODO Add Query Options to Requrest

	// List Registered Orgs
	q := r.MustGet("query-conditions").(*query.QueryConditions)
	orgs, err := orm.RegisteredOrgsQuery(db, q, true)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save List
	r.Set("orgs", orgs)
}

func DBRegistryOrgFind(r rpf.GINProcessor, c *gin.Context) {
	// Get User Identifier
	id := r.MustGet("org")

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Global Organization Registry
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Org based on ID Type
	entry := &orm.OrgRegistry{}
	err = entry.Find(db, id)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Did we find the User?
	if !entry.IsValid() { // NO: User does not exist
		r.Abort(4000, nil)
		return
	}

	// Save Org Registry
	r.SetLocal("registry-org", entry)
}

func DBRegistryOrgFindByID(r rpf.GINProcessor, c *gin.Context) {
	// Do we already have the registry loaded?
	registry := r.Get("registry-org")
	if registry == nil { // NO: Load
		// Get User Identifier
		id := r.MustGet("org-id").(uint64)

		// Get Database Connection Manager
		dbm := c.MustGet("dbm").(*orm.DBSessionManager)

		// Get Connection to Org Registry
		db, err := dbm.ConnectTo(0, 0)
		if err != nil { // YES: Database Error
			r.Abort(5100, nil)
			return
		}

		// Get Org by ID
		entry := &orm.OrgRegistry{}
		err = entry.ByID(db, id)

		// Failed Retrieving Org?
		if err != nil { // YES: Database Error
			r.Abort(5100, nil)
			return
		}

		// Did we find the Org?
		if !entry.IsValid() { // NO: Org does not exist
			r.Abort(4000, nil)
			return
		}

		// Save ORG
		r.Set("registry-org", entry)
	}
}

func DBRegisterOrg(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization Information
	org := r.MustGet("org").(*orm.Organization)
	org_id := r.MustGet("org-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Create ORG Entry
	e := &orm.OrgRegistry{}
	e.SetID(org_id)
	e.SetAlias(org.Alias())
	e.SetName(org.Name())

	// Connect to Global Registry Shard
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Organization
	err = e.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Entry
	r.SetLocal("registry-org", e)
}

func DBRegistryOrgUpdate(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization Registry Entry
	e := r.MustGet("registry-org").(*orm.OrgRegistry)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Connect to Global Registry Shard
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Modifications
	err = e.Flush(db, false)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}

func DBRegistryOrgUpdateFromOrg(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization Information
	org := r.MustGet("org").(*orm.Organization)

	// Get Organization Registry Entry
	e := r.MustGet("org").(*orm.OrgRegistry)

	// Do we need to Update the Organization Registry?
	if org.UpdateRegistry() { // YES
		e.SetAlias(org.Alias())
		e.SetName(org.Name())
	}

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Connect to Global Registry Shard
	db, err := dbm.ConnectTo(0, 0)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save Organization
	err = e.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}
