// cSpell:ignore gonic, orgs, paulo, ferreira
package shared

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
)

// TYPE of Addin Optional Parameters Callback
type TAddinCallbackOptions func(string) interface{}

func HelperAddinOptionsCallback(c TAddinCallbackOptions, opt string, d interface{}) interface{} {
	// Do we have a callback?
	if c != nil { // YES: Request Option Value
		v := c(opt)
		// Is Option Value NIL
		if v == nil { // YES: Return Default
			return d
		}
		// ELSE: Return Value
		return v
	}
	// ELSE: Return Default
	return d
}

// Common Addin : Maps Multiple Parameters to new names
func AddinRemapParameters(g rpf.GINGroupProcessor, p_map map[string]string) {
	g.Append(func(r rpf.GINProcessor, c *gin.Context) {
		for key, to_key := range p_map {
			r.SetLocal(to_key, r.MustGet(key))
		}
	})
}

// Common Addin : Handler has Incomplete Implementation
func AddinIncomplete(g rpf.GINGroupProcessor, p_map map[string]string) {
	// Handler Incomplete
	g.Append(func(r rpf.GINProcessor, c *gin.Context) {
		r.Abort(5999, nil)
	})
}

// Common Addin : Handler has not been Implemented
func AddinToDo(g rpf.GINGroupProcessor, p_map map[string]string) {
	// Handler Not Implemented
	g.Append(func(r rpf.GINProcessor, c *gin.Context) {
		r.Abort(5999, nil)
	})
}
