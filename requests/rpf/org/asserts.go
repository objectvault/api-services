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
	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

// ORGANIZATION //

func AssertNotSystemOrgRequest(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Organization ID
	oid := r.MustGet("request-org").(uint64)

	// Is System Organization?
	if oid == common.SYSTEM_ORGANIZATION { // YES: Abort
		r.Abort(4101, nil)
		return
	}
}

func AssertNotSystemOrgRegistry(r rpf.GINProcessor, c *gin.Context) {
	registry := r.MustGet("registry-org").(*orm.OrgRegistry)

	// Is System Organization (LOCAL ID = 0)?
	if registry.IsSystem() { // YES: Abort
		r.Abort(4101, nil)
		return
	}
}

func AssertOrgNotDeleted(r rpf.GINProcessor, c *gin.Context) {
	// Get Request User
	org := r.MustGet("registry-org").(*orm.OrgRegistry)

	// Is the User Account in Deleted Mode?
	if org.IsDeleted() { // YES
		r.Abort(4002, nil)
		return
	}
}

func AssertOrgUnblocked(r rpf.GINProcessor, c *gin.Context) {
	// Get Request Organization
	org := r.MustGet("registry-org").(*orm.OrgRegistry)

	// Is the Organization Blocked? (HARD CODE: Can't Block System Organization)
	if !org.IsSystem() && org.IsBlocked() { // YES: Abort
		r.Abort(4199, nil) // TODO: Specific Error
		return
	}
}
