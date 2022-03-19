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
	"fmt"

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

func DBGetOrgByID(r rpf.GINProcessor, c *gin.Context) {
	// Get User Identifier
	id := r.MustGet("org-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Organization's Shard
	db, err := dbm.Connect(id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Get Org based on ID Type
	entry := &orm.Organization{}
	err = entry.ByID(db, common.LocalIDFromID(id))

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
	r.SetLocal("org", entry)
}

func DBInsertOrg(r rpf.GINProcessor, c *gin.Context) {
	// Get Organization ID
	org := r.MustGet("org").(*orm.Organization)

	// User ID of Creator
	user_id := r.MustGet("user-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Random Shard
	shard := common.RandomShardID()

	// Get Connection to Random Shard
	db, err := dbm.ConnectTo(1, shard)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Set Organization Creator
	org.SetCreator(user_id)

	// Save Organization
	err = org.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Calculate Global Organization ID
	org_id := common.ShardGlobalID(1, common.OTYPE_ORG, shard, org.ID())
	h := fmt.Sprintf("%X", org_id)
	fmt.Printf("ORG ID [%s]\n", h)

	// Save Shard Information and Organization Entry
	r.SetLocal("org-id", org_id)
}

func DBUpdateOrg(r rpf.GINProcessor, c *gin.Context) {
	// Organization ID
	org_id := r.MustGet("org-id").(uint64)

	// Get Organization
	org := r.MustGet("org").(*orm.Organization)

	// User ID of Modifier
	user_id := r.MustGet("user-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Organization Shard
	db, err := dbm.Connect(org_id)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Set Organization Creator
	org.SetModifier(user_id)

	// Save Organization
	err = org.Flush(db, false)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}
