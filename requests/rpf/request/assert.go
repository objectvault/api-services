package request

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

// cSpell:ignore ormrequest, reqtype
import (
	"github.com/gin-gonic/gin"

	ormrequest "github.com/objectvault/api-services/orm/request"

	rpf "github.com/objectvault/goginrpf"
)

func AssertRequestRegOfType(r rpf.GINProcessor, c *gin.Context) {
	// Get Invitation
	reqtype := r.MustGet("registry-type").(string)
	request := r.MustGet("registry-request").(*ormrequest.RequestRegistry)

	// Is Request of Correct Type?
	if request.RequestType() != reqtype { // NO
		r.Abort(4390, nil)
		return
	}
}

func AssertRequestRegActive(r rpf.GINProcessor, c *gin.Context) {
	// Get Request
	request := r.MustGet("registry-request").(*ormrequest.RequestRegistry)

	// Is Request Still Active?
	if request.State() != ormrequest.STATE_ACTIVE { // YES
		r.Abort(4391, nil)
		return
	}

	// Is Request Expired?
	if request.IsExpired() { // YES
		r.Abort(4491, nil)
		return
	}
}

func AssertInvitationNotExpired(r rpf.GINProcessor, c *gin.Context) {
	// Get Request
	request := r.MustGet("registry-request").(*ormrequest.RequestRegistry)

	// Is Request Expired?
	if request.IsExpired() { // YES
		r.Abort(4491, nil)
		return
	}
}
