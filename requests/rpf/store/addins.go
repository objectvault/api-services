// cSpell:ignore ginrpf, gonic, paulo ferreira
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
	"github.com/gin-gonic/gin"
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/object"
	"github.com/objectvault/api-services/requests/rpf/org"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/user"
	rpf "github.com/objectvault/goginrpf"
)

// COMMON //
func BaseValidateStoreRequest(g rpf.GINGroupProcessor, opts shared.TAddinCallbackOptions) rpf.GINGroupProcessor {
	session.AddinActiveUserSession(g, opts)
	g.Append(
		DBStoreUserGet, // FIND User by Searching Store User Registry
	)

	/* NOTES:
	 * Store are Children of Organizations.
	 * So given only the Store ID how do we get the Organization?
	 * Answer: From the Store Object (which has the Parent Organization ID)
	 */

	// OPTION: Check if organization is unblocked? (DEFAULT: Check)
	if shared.HelperAddinOptionsCallback(opts, "check-org-unlocked", true).(bool) {
		g.Append(
			DBStoreGetByID, // Get Store from ID
			func(r rpf.GINProcessor, c *gin.Context) { // Get Store Organization
				s := r.MustGet("store").(*orm.Store)
				r.SetLocal("org-id", s.Organization())
			},
			org.DBRegistryOrgFindByID, // Get Organization Registry
			org.AssertOrgUnblocked,    // Make Sure Org Unblocked
		)

		// OPTION: Check if store is unblocked? (DEFAULT: Check)
		if shared.HelperAddinOptionsCallback(opts, "check-store-unlocked", true).(bool) {
			g.Append(
				org.DBOrgStoreFind,   // Find Store's Organization  Registry
				AssertStoreUnblocked, // Make Store Unblocked
			)
		}
	} else // OPTION: Check if store is unblocked (but skip org check)? (DEFAULT: Check)
	if shared.HelperAddinOptionsCallback(opts, "check-store-unlocked", true).(bool) {
		g.Append(
			DBStoreGetByID, // Get Store from ID
			func(r rpf.GINProcessor, c *gin.Context) { // Get Store Organization
				s := r.MustGet("store").(*orm.Store)
				r.SetLocal("org-id", s.Organization())
			},
			org.DBRegistryOrgFindByID, // Get Organization Registry
			org.DBOrgStoreFind,        // Find Store's Organization  Registry
			AssertStoreUnblocked,      // Make Store Unblocked
		)
	}

	// OPTION: Check user's store roles? (DEFAULT: Check)
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

// REQUEST TYPE :org/:store //

// Extract GIN Request Parameters for a organization/store request
func AddinRequestParamsOrgStore(g rpf.GINGroupProcessor) rpf.GINGroupProcessor {
	g.Append(
		org.ExtractGINParameterOrg,
		ExtractGINParameterStore,
	)
	return g
}

// Global initial organization/user request validation
func AddinGroupValidateOrgStoreRequest(g rpf.GINGroupProcessor, opts shared.TAddinCallbackOptions) rpf.GINGroupProcessor {
	// Extract Request Parameter
	AddinRequestParamsOrgStore(g)

	// Validate Basic User Request Requirements
	return org.BaseValidateOrgRequest(g,
		func(o string) interface{} {
			if o == "assert-not-system" { // Org Store Requests not allowed on system organization
				return true
			}

			return opts(o)
		})
}

// REQUEST TYPE :store //

// Extract GIN Request Parameters for a store request
func AddinRequestParamsStore(g rpf.GINGroupProcessor) rpf.GINGroupProcessor {
	g.Append(
		ExtractGINParameterStore,
	)
	return g
}

// Global initial store request validation
func AddinGroupValidateStoreRequest(g rpf.GINGroupProcessor, opts shared.TAddinCallbackOptions) rpf.GINGroupProcessor {
	// Extract Request Parameter
	AddinRequestParamsStore(g)

	// Validate Basic User Request Requirements
	return BaseValidateStoreRequest(g, opts)
}

// REQUEST TYPE :store/:user //

// Extract GIN Request Parameters for a store/user request
func AddinRequestParamsStoreUser(g rpf.GINGroupProcessor) rpf.GINGroupProcessor {
	g.Append(
		ExtractGINParameterStore,
		user.ExtractGINParameterUser,
	)
	return g
}

// Global initial store/user request validation
func AddinGroupValidateStoreUserRequest(g rpf.GINGroupProcessor, opts shared.TAddinCallbackOptions) rpf.GINGroupProcessor {
	// Extract Request Parameter
	AddinRequestParamsStoreUser(g)

	// OPTION: Check if session user is requested user ? (DEFAULT: NO Check)
	if shared.HelperAddinOptionsCallback(opts, "assert-if-self", false).(bool) {
		g.Append(
			session.AssertIfSelf, // Make Sure user is Not Applying Action to himself
		)
	}

	// Validate Basic User Request Requirements
	return BaseValidateStoreRequest(g, opts)
}

func AddinRequestStoreUserRegistry(g rpf.GINGroupProcessor) rpf.GINGroupProcessor {
	g.Append(
		func(r rpf.GINProcessor, c *gin.Context) {
			r.SetLocal("object-id", r.MustGet("request-store"))
			r.SetLocal("user-id", r.MustGet("request-user"))
		},
		object.DBObjectUserFind,
	)
	return g
}
