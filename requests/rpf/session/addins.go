// cSpell:ignore addin, goginrpf, gonic, paulo ferreira
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

	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/user"
)

func AddinActiveUserSession(g rpf.GINGroupProcessor, opts shared.TAddinCallbackOptions) rpf.GINGroupProcessor {
	g.Append(AssertUserSession)

	// OPTION: Check if user is admin? (DEFAULT: No Check)
	if shared.HelperAddinOptionsCallback(opts, "check-not-admin", false).(bool) {
		g.Append(AssertNotSystemAdmin)
	}

	// OPTION: Check Session User Status
	sub := shared.HelperAddinOptionsCallback(opts, "assert-session-user-blocked", true).(bool)
	suro := shared.HelperAddinOptionsCallback(opts, "assert-session-user-readonly", true).(bool)

	if sub || suro {
		g.Append(
			SessionUserToRegistry,
		)

		if sub {
			g.Append(
				user.AssertUserBlocked,
			)
		}

		if suro {
			g.Append(
				user.AssertUserReadOnly,
			)
		}
	} else { // NO
		g.Append(SessionExtractUser)
	}

	return g
}

func AddinSaveSession(g rpf.GINGroupProcessor, opts shared.TAddinCallbackOptions) rpf.GINGroupProcessor {
	// Update Session Cookie
	g.Append(SaveSession)

	return g
}
