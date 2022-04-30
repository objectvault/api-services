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
	"github.com/gin-gonic/gin"

	rpf "github.com/objectvault/goginrpf"

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/object"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/user"
)

// COMMON //
func BaseValidateOrgRequest(g rpf.GINGroupProcessor, opts shared.TAddinCallbackOptions) rpf.GINGroupProcessor {
	session.AddinActiveUserSession(g, opts)

	// OPTION: Check if organization is unblocked? (DEFAULT: Check)
	if shared.HelperAddinOptionsCallback(opts, "check-org-unlocked", true).(bool) {
		g.Append(
			func(r rpf.GINProcessor, c *gin.Context) {
				r.SetLocal("org", r.MustGet("request-org"))
			},
			DBRegistryOrgFind,
			AssertOrgUnblocked,
			func(r rpf.GINProcessor, c *gin.Context) {
				o := r.MustGet("registry-org").(*orm.OrgRegistry)
				r.SetLocal("request-org", o.ID())
				r.SetLocal("org-id", o.ID())
			},
		)
	}

	// OPTION: Check if user is unblocked? (DEFAULT: Check)
	if shared.HelperAddinOptionsCallback(opts, "check-user-unlocked", true).(bool) {
		g.Append(
			object.DBRegistryOrgUserFind,
			object.AssertObjectUserUnblocked,
		) // Make sure user is Unblocked
	}

	// OPTION: Check if session user is requested user ? (DEFAULT: NO Check)
	if shared.HelperAddinOptionsCallback(opts, "check-not-self", false).(bool) {
		g.Append(
			session.AssertNotSelf, // Make Sure user is Not Applying Action to himself
		)
	}

	// OPTION: Check user's org roles? (DEFAULT: Check)
	if shared.HelperAddinOptionsCallback(opts, "check-user-roles", true).(bool) {
		g.Append(
			func(r rpf.GINProcessor, c *gin.Context) {
				skipSelf := shared.HelperAddinOptionsCallback(opts, "skip-roles-if-self", false).(bool)
				skipRoles := skipSelf || !session.IsSelf(c, r.MustGet("user-id").(uint64))

				// Skip Roles Check
				if !skipRoles { // NO
					roles := shared.HelperAddinOptionsCallback(opts, "roles", nil)
					if roles == nil {
						// TODO: errors.New("System Error: Missing Option Value")
						r.Abort(5303, nil)
						return
					}

					r.SetLocal("roles-required", roles)
					object.AssertUserHasAllRolesInObject(r, c) // ASSERT User has required Access Roles
				}
			},
		)
	}

	return g
}

// REQUEST TYPE :org //

// Extract GIN Request Parameters for a org request
func AddinRequestParamsOrg(g rpf.GINGroupProcessor) rpf.GINGroupProcessor {
	g.Append(
		ExtractGINParameterOrg,
	)
	return g
}

// Global initial organization request validation
func AddinGroupValidateOrgRequest(g rpf.GINGroupProcessor, opts shared.TAddinCallbackOptions) rpf.GINGroupProcessor {
	// OPTION: Request ORganization System Organization? (DEFAULT: NO)
	if shared.HelperAddinOptionsCallback(opts, "system-organization", false).(bool) {
		// ORGANIZATION for Request is System Organization //
		g.Append(
			func(r rpf.GINProcessor, c *gin.Context) {
				// This is so the API does not have to know the System Organization ID
				r.SetLocal("request-org", common.SYSTEM_ORGANIZATION)
			})
	} else {
		// Extract Request Parameter
		AddinRequestParamsOrg(g)
	}

	// Validate Basic User Request Requirements
	return BaseValidateOrgRequest(g, opts)
}

// REQUEST TYPE :store/:user //

// Extract GIN Request Parameters for a organization/user request
func AddinRequestParamsOrgUser(g rpf.GINGroupProcessor) rpf.GINGroupProcessor {
	g.Append(
		ExtractGINParameterOrg,
		user.ExtractGINParameterUser,
	)
	return g
}

// Global initial organization/user request validation
func AddinGroupValidateOrgUserRequest(g rpf.GINGroupProcessor, opts shared.TAddinCallbackOptions) rpf.GINGroupProcessor {
	// Extract Request Parameter
	AddinRequestParamsOrgUser(g)

	// Validate Basic User Request Requirements
	return BaseValidateOrgRequest(g, opts)
}
