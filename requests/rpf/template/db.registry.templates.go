// cSpell:ignore goginrpf, gonic, paulo ferreira
package template

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

	rpf "github.com/objectvault/goginrpf"
)

func DBGetObjectTemplates(r rpf.GINProcessor, c *gin.Context) {
	// Get Object Identifier
	obj := r.MustGet("object-id").(uint64)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Object Shard
	db, err := dbm.Connect(obj)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// List Registered Org Users
	q := r.MustGet("query-conditions").(*orm.QueryConditions)
	templates, err := orm.QueryRegisteredObjectTemplates(db, obj, q, true)

	// Failed Retrieving User?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Save List
	r.Set("registry-object-templates", templates)
}

func DBGetObjectTemplateRegistry(r rpf.GINProcessor, c *gin.Context) {
	// Get Object Identifier
	obj := r.MustGet("object-id").(uint64)

	// Get Template
	template := r.MustGet("request-template").(string)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Object Shard
	db, err := dbm.Connect(obj)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Create Registry Entry
	o := &orm.ObjectTemplateRegistry{}
	err = o.ByTemplate(db, obj, template)

	// Failed To Retrive Entry?
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Did we find the Template?
	if !o.IsValid() { // NO: Template Not Registered
		r.Abort(4998 /* TODO: Error [Template not registered with Object] */, nil)
		return
	}

	// Save List
	r.Set("registry-object-template", o)
}

func DBRegisterTemplateWithObject(r rpf.GINProcessor, c *gin.Context) {
	// Get Object Identifier
	obj := r.MustGet("object-id").(uint64)

	// Get Template Name
	template := r.MustGet("request-template").(string)
	title := r.MustGet("template-title").(string)

	// Create Registry Entry
	o := &orm.ObjectTemplateRegistry{}
	o.SetKey(obj, template)
	o.SetTitle(title)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Object Shard
	db, err := dbm.Connect(obj)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Write Entry
	err = o.Flush(db, true)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	r.SetLocal("registry-object-template", o)
}

func DBDeleteTemplateFromObject(r rpf.GINProcessor, c *gin.Context) {
	// Get Object Identifier
	obj := r.MustGet("object-id").(uint64)

	// Get Template Name
	template := r.MustGet("template-name").(string)

	// Get Database Connection Manager
	dbm := c.MustGet("dbm").(*orm.DBSessionManager)

	// Get Connection to Object Shard
	db, err := dbm.Connect(obj)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}

	// Delete Template from Object
	err = orm.DeleteRegisteredObjectTemplate(db, obj, template)
	if err != nil { // YES: Database Error
		r.Abort(5100, nil)
		return
	}
}
