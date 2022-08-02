package action

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
	"github.com/objectvault/api-services/orm/action"
	"github.com/objectvault/api-services/orm/request"
	rpf "github.com/objectvault/goginrpf"
)

func ActionCreatePasswordReset(r rpf.GINProcessor, c *gin.Context) {
	// Get the Required Invitation
	or := r.MustGet("request").(*request.Request)
	orr := r.MustGet("registry-request").(*request.RequestRegistry)
	our := r.MustGet("registry-user").(*orm.UserRegistry)

	oa := action.NewActionWithGUID(or.GUID(), "email:password:reset", or.Creator())

	// Set Action Parameters
	params := oa.Parameters()
	params.Import(or.Parameters().Export())

	// DEFAULT: Organization Invitation
	params.Set("template", "password-reset", true)

	// Set Action Properties
	props := oa.Properties()
	props.Import(or.Properties().Export())

	props.Set("to", our.Email(), true)
	props.Set("expiration", orr.ExpirationUTC(), true)

	// Save Activation
	r.Set("action", oa)
}
