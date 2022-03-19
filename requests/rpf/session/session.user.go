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
	"github.com/objectvault/api-services/requests/rpf/user"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SessionExtractUser(r rpf.GINProcessor, c *gin.Context) {
	// Get Session Store
	session := sessions.Default(c)

	// Extract User Information from Session
	id := session.Get("user-id").(uint64)
	r.SetLocal("user-id", id)
}

func SessionUserToRegistry(r rpf.GINProcessor, c *gin.Context) {
	// Extract User Information from Session
	SessionExtractUser(r, c)

	// Get Entry for User
	user.DBRegistryUserFindByID(r, c)
}
