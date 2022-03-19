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
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func IsSelf(c *gin.Context, uid uint64) bool {
	// Get Session Store
	session := sessions.Default(c)

	// Does Session USer match the ID?
	sid := session.Get("user-id")
	if sid != nil && sid == uid { // YES
		return true
	}

	// ELSE: No Session or User ID Doesn't match
	return false
}
