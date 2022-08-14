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

// cSpell:ignore objs, vmap, xjson
import (
	"errors"
	"fmt"
	"strings"

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/api-services/requests/rpf/entry"
	"github.com/objectvault/api-services/requests/rpf/session"
	"github.com/objectvault/api-services/requests/rpf/shared"
	"github.com/objectvault/api-services/requests/rpf/store"
	"github.com/objectvault/api-services/xjson"

	rpf "github.com/objectvault/goginrpf"

	"github.com/gin-gonic/gin"
)

// TODO IMPLEMENT: LIST Store Objects
func GetStoreObjects(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.STORE.OBJS", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameter 'store'
		store.ExtractGINParameterStore,
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Store
			storeID := r.MustGet("request-store").(uint64)

			// Required Roles : Store Access with Read Function
			roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_OBJECT, orm.FUNCTION_LIST)}

			// Initialize Request
			store.GroupStoreRequestInitialize(r, storeID, roles).
				Run()
		},
		entry.ExtractGINParameterParentID,
		// Extract Query Parameters //
		func(r rpf.GINProcessor, c *gin.Context) {
			gQuery := shared.GroupExtractQueryConditions(r, nil, func(f string) string {
				switch f {
				case "id":
					return "id"
				case "title":
					return "title"
				case "type": // Cannot Sort but can Filter
					return "type"
				default: // Invalid Field
					return ""
				}
			}, nil)

			gQuery.Run()
			if !r.IsFinished() { // YES
				// Save Query Settings as Global
				gQuery.LocalToGlobal("query-conditions")
			}
		},
		// Query System for List //
		func(r rpf.GINProcessor, c *gin.Context) {
			r.SetLocal("store-parent-id", r.MustGet("request-parent-id"))
		},
		entry.DBStoreObjectList,
		// Export Results //
		entry.ExportStoreObjectList,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: MASS Delete Store Objects
func DeleteStoreObjects(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.STORE.OBJS", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		/* REQUEST VALIDATION */
		func(r rpf.GINProcessor, c *gin.Context) {
			r.Abort(5999, nil)
		},
	}

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: READ Store Object
func GetStoreObject(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("GET.STORE.OBJ", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameter 'store'
		store.ExtractGINParameterStore,
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Store
			storeID := r.MustGet("request-store").(uint64)

			// Required Roles : Store Access with Create Function
			roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_OBJECT, orm.FUNCTION_READ_LIST)}

			// Initialize Request
			// TODO: Assert Store Unlocked
			store.GroupStoreRequestInitialize(r, storeID, roles).
				Run()
		},
		// Assert Store is Open
		store.AssertStoreOpen,
		// Extract Required Parameters
		entry.ExtractGINParameterEntryID,
		entry.AssertNotRootFolder,
		entry.DBGetStoreObjectByID,
		entry.DecryptStoreObject,
		// Export Results //
		func(r rpf.GINProcessor, c *gin.Context) {
			obj := r.MustGet("store-object").(*orm.StoreObject)

			switch obj.Type() {
			case orm.OBJECT_TYPE_FOLDER:
				entry.ExportStoreObjectFolder(r, c)
			case orm.OBJECT_TYPE_JSON:
				entry.ExportStoreObjectJSON(r, c)
			default:
				r.Abort(4998 /* TODO: Error [Unknown Object Type] */, nil)
			}
		},
		// Extend and Save Store Session //
		session.ExtendStoreSession,
		session.SessionStoreSave,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: Create Store Object
func PostStoreObjectJSON(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("POST.STORE.OBJ.JSON", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameter 'store'
		store.ExtractGINParameterStore,
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Store
			storeID := r.MustGet("request-store").(uint64)

			// Required Roles : Store Access with Create Function
			roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_OBJECT, orm.FUNCTION_CREATE)}

			// Initialize Request
			// TODO: Assert Store Unlocked
			store.GroupStoreRequestInitialize(r, storeID, roles).
				Run()
		},
		// Assert Store is Open
		store.AssertStoreOpen,
		// Make sure parent Exists
		entry.ExtractGINParameterParentID, // Extract Required Parameters
		func(r rpf.GINProcessor, c *gin.Context) {
			pid := r.MustGet("request-parent-id").(uint32)

			if pid != 0 {
				// Create Processing Group
				group := &rpf.ProcessorGroup{}
				group.Parent = r

				// Set Object ID to Match
				group.SetLocal("store-object-id", pid)

				// See if Object ID Exists and is Folder Object
				group.Chain = rpf.ProcessChain{
					entry.DBGetStoreObjectByID,
					entry.AssertFolderObject,
					func(r rpf.GINProcessor, c *gin.Context) {
						r.SetLocal("store-parent-object", r.MustGet("store-object"))
					},
				}

				group.Run()
			}
		},
		// PROCESS JSON Object //
		shared.RequestExtractJSON, // Has to have a JSON Body
		func(r rpf.GINProcessor, c *gin.Context) {
			// Create Object
			o := &orm.StoreObject{}
			t := &orm.StoreTemplateObject{}

			// Extract and Validate JSON MEssage
			m := r.MustGet("request-json").(xjson.T_xMap)
			vmap := xjson.S_xJSONMap{Source: m}

			// Template Name
			vmap.Required("template.name", nil, func(v interface{}) (interface{}, error) {
				v, e := xjson.F_xToTrimmedString(v)
				if e != nil {
					return nil, e
				}

				s := v.(string)
				if s == "" {
					return nil, errors.New("Object Missing Template")
				}
				return strings.ToLower(s), nil
			}, func(v interface{}) error {
				switch v.(string) {
				case "folder":
					o.SetType(orm.OBJECT_TYPE_FOLDER)
				default:
					o.SetType(orm.OBJECT_TYPE_JSON)
				}
				t.SetTemplate(v.(string))
				return nil
			})

			// Template Name
			vmap.Required("template.version", nil, func(v interface{}) (interface{}, error) {
				v, e := xjson.F_xToUint64(v)
				if e != nil {
					return nil, e
				}

				version := v.(uint64)
				if version == 0 {
					return nil, errors.New("Object Missing Template")
				}
				return uint16(version), nil
			}, func(v interface{}) error {
				t.SetVersion(v.(uint16))
				return nil
			})

			// Object Template Values
			vmap.Required("values", nil, func(v interface{}) (interface{}, error) {
				if v == nil {
					return nil, errors.New("Object Missing Values")
				}

				if _, ok := v.(map[string]interface{}); ok {
					return v, nil
				}
				return nil, errors.New("Object has Invalid Values")
			}, func(v interface{}) error {
				t.SetValues(v.(map[string]interface{}))
				return nil
			})

			// Object Title
			vmap.Required("values.__title", nil, func(v interface{}) (interface{}, error) {
				v, e := xjson.F_xToTrimmedString(v)
				if e != nil {
					return nil, e
				}

				s := v.(string)
				if s == "" {
					return nil, errors.New("Object Missing Title")
				}
				return s, nil
			}, func(v interface{}) error {
				o.SetTitle(v.(string))
				t.SetTitle(v.(string))
				return nil
			})

			// Did we have an Error Processing the Map?
			if vmap.Error != nil {
				fmt.Println(vmap.Error)
				fmt.Println(vmap.StringSrc())
				r.Abort(4998 /* TODO: ERROR [Object Missing Title] */, nil)
				return
			}

			// Save Object
			r.SetLocal("store-object", o)
			r.SetLocal("store-template-object", t)
		},
		entry.EncryptStoreObject,
		// Set Object
		func(r rpf.GINProcessor, c *gin.Context) {
			o := r.MustGet("store-object").(*orm.StoreObject)
			sid := r.MustGet("store-id").(uint64)
			pid := r.MustGet("request-parent-id").(uint32)

			// Initialize Store Object
			o.SetStore(common.LocalIDFromID(sid))
			o.SetParent(pid)

			r.SetLocal("store-parent-id", pid)
		},
		entry.DBInsertStoreObject,
		// Export Results //
		func(r rpf.GINProcessor, c *gin.Context) {
			obj := r.MustGet("store-object").(*orm.StoreObject)

			switch obj.Type() {
			case orm.OBJECT_TYPE_FOLDER:
				entry.ExportStoreObjectFolder(r, c)
			case orm.OBJECT_TYPE_JSON:
				entry.ExportStoreObjectJSON(r, c)
			default:
				r.Abort(4998 /* TODO: Error [Unknown Object Type] */, nil)
			}
		},
		// Extend and Save Store Session //
		session.ExtendStoreSession,
		session.SessionStoreSave,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: UPDATE Store Object
func PutStoreObjectJSON(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("PUT.STORE.OBJ.JSON", c, 1000, shared.JSONResponse)

	/* NOTE: Update Requires that all of the object information be resent
	 * EVEN THE INFORMATION THAT HAS NOT CHANGED
	 */
	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameter 'store'
		store.ExtractGINParameterStore,
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Store
			storeID := r.MustGet("request-store").(uint64)

			// Required Roles : Store Access with Create Function
			roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_OBJECT, orm.FUNCTION_UPDATE)}

			// Initialize Request
			// TODO: Assert Store Unlocked
			store.GroupStoreRequestInitialize(r, storeID, roles).
				Run()
		},
		// Assert Store is Open
		store.AssertStoreOpen,
		// Extract Parent ID
		entry.ExtractGINParameterParentID,
		func(r rpf.GINProcessor, c *gin.Context) {
			pid := r.MustGet("request-parent-id").(uint32)

			if pid != 0 {
				// Create Processing Group
				group := &rpf.ProcessorGroup{}
				group.Parent = r

				// Set Object ID to Match
				group.SetLocal("store-object-id", pid)

				// See if Object ID Exists and is Folder Object
				group.Chain = rpf.ProcessChain{
					entry.DBGetStoreObjectByID,
					entry.AssertFolderObject,
					func(r rpf.GINProcessor, c *gin.Context) {
						r.SetLocal("store-parent-object", r.MustGet("store-object"))
					},
				}

				group.Run()
			}
		},
		// Extract Updated Object ID
		entry.ExtractGINParameterEntryID,
		entry.AssertNotRootFolder,
		entry.DBGetStoreObjectByID,
		entry.DecryptStoreObject,
		// PROCESS JSON Object //
		shared.RequestExtractJSON, // Has to have a JSON Body
		func(r rpf.GINProcessor, c *gin.Context) {
			o := r.MustGet("store-object").(*orm.StoreObject)
			t := r.MustGet("store-template-object").(*orm.StoreTemplateObject)

			// Extract and Validate JSON MEssage
			m := r.MustGet("request-json").(xjson.T_xMap)
			vmap := xjson.S_xJSONMap{Source: m}

			// TODO: Can't update Template Name (But Can update Template Version)
			// Template Name
			vmap.Required("template.name", nil, func(v interface{}) (interface{}, error) {
				v, e := xjson.F_xToTrimmedString(v)
				if e != nil {
					return nil, e
				}

				s := v.(string)
				if s == "" {
					return nil, errors.New("Object Missing Template")
				}
				return strings.ToLower(s), nil
			}, func(v interface{}) error {
				var ot uint8
				template := v.(string)
				switch template {
				case "folder":
					ot = orm.OBJECT_TYPE_FOLDER
				default:
					ot = orm.OBJECT_TYPE_JSON
				}

				if ot != o.Type() {
					return errors.New("Can't Modify Object Type After Creation")
				}

				if template != t.Template() {
					return errors.New("Can't Modify Object Template After Creation")
				}
				return nil
			})

			// Object Template Values
			vmap.Required("values", nil, func(v interface{}) (interface{}, error) {
				if v == nil {
					return nil, errors.New("Object Missing Values")
				}

				if _, ok := v.(map[string]interface{}); ok {
					return v, nil
				}
				return nil, errors.New("Object has Invalid Values")
			}, func(v interface{}) error {
				t.SetValues(v.(map[string]interface{}))
				return nil
			})

			// Object Title
			vmap.Required("values.__title", nil, func(v interface{}) (interface{}, error) {
				v, e := xjson.F_xToTrimmedString(v)
				if e != nil {
					return nil, e
				}

				s := v.(string)
				if s == "" {
					return nil, errors.New("Object Missing Title")
				}
				return s, nil
			}, func(v interface{}) error {
				o.SetTitle(v.(string))
				return nil
			})

			// Did we have an Error Processing the Map?
			if vmap.Error != nil {
				fmt.Println(vmap.Error)
				fmt.Println(vmap.StringSrc())
				r.Abort(4998 /* TODO: ERROR [Object Missing Title] */, nil)
				return
			}
		},
		// Set Object
		entry.EncryptStoreObject,
		func(r rpf.GINProcessor, c *gin.Context) {
			o := r.MustGet("store-object").(*orm.StoreObject)
			pid := r.MustGet("request-parent-id").(uint32)

			o.SetParent(pid)

			r.SetLocal("store-parent-id", pid)
		},
		entry.DBUpdateStoreObject,
		// Export Results //
		func(r rpf.GINProcessor, c *gin.Context) {
			obj := r.MustGet("store-object").(*orm.StoreObject)

			switch obj.Type() {
			case orm.OBJECT_TYPE_FOLDER:
				entry.ExportStoreObjectFolder(r, c)
			case orm.OBJECT_TYPE_JSON:
				entry.ExportStoreObjectJSON(r, c)
			default:
				r.Abort(4998 /* TODO: Error [Unknown Object Type] */, nil)
			}
		},
		// Extend and Save Store Session //
		session.ExtendStoreSession,
		session.SessionStoreSave,
		session.SaveSession, // Update Session Cookie
	}

	// Start Request Processing
	request.Run()
}

// TODO IMPLEMENT: DELETE Store Object
func DeleteStoreObject(c *gin.Context) {
	// Create Request
	request := rpf.RootProcessor("DELETE.STORE.OBJ", c, 1000, shared.JSONResponse)

	// Request Processing Chain
	request.Chain = rpf.ProcessChain{
		// Extract Route Parameter 'store'
		store.ExtractGINParameterStore,
		// Validate Basic Request Settings
		func(r rpf.GINProcessor, c *gin.Context) {
			// Get Request Store
			storeID := r.MustGet("request-store").(uint64)

			// Required Roles : Store Access with Create Function
			roles := []uint32{orm.Role(orm.CATEGORY_STORE|orm.SUBCATEGORY_OBJECT, orm.FUNCTION_UPDATE)}

			// Initialize Request
			// TODO: Assert Store Unlocked
			store.GroupStoreRequestInitialize(r, storeID, roles).
				Run()
		},
		// Assert Store is Open
		store.AssertStoreOpen,
		// Extract Required Parameters
		entry.ExtractGINParameterEntryID,
		entry.AssertNotRootFolder,
		entry.DBGetStoreObjectByID,
		entry.DBDeleteStoreObject,
		entry.ExportStoreObjectRegistry,
		// Extend and Save Store Session //
		session.ExtendStoreSession,
		session.SessionStoreSave,
		session.SaveSession,
	}

	// Start Request Processing
	request.Run()
}
