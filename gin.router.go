// cSpell:ignore gonic, orgs, paulo, ferreira
package main

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

// cSpell:ignore amqp, objs, pkginvites, pkgme, pkgorg, pkpwd, pkgsession, pkgstore, pkgsystem, sharded

import (
	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/queue-interface/queue"
	"github.com/objectvault/queue-interface/shared"

	pkginvites "github.com/objectvault/api-services/requests/handlers/invitation"
	pkgme "github.com/objectvault/api-services/requests/handlers/me"
	pkgorg "github.com/objectvault/api-services/requests/handlers/org"
	pkgpwd "github.com/objectvault/api-services/requests/handlers/password"
	pkgsession "github.com/objectvault/api-services/requests/handlers/session"
	pkgstore "github.com/objectvault/api-services/requests/handlers/store"
	pkgsystem "github.com/objectvault/api-services/requests/handlers/system"

	"github.com/gin-gonic/gin"
)

var gDBManager *orm.DBSessionManager
var gQConnection *queue.AMQPServerConnection

func databaseManager() (*orm.DBSessionManager, error) {
	if gDBManager == nil {
		dbm := common.ShardedDatabase{}

		// Do we have a Database Configuration Object?
		o, e := common.ConfigPropertyObject(Config, "database", nil, nil)
		if e != nil { // NO
			return nil, e
		}
		e = dbm.FromConfig(o)
		if e != nil { // NO
			return nil, e
		}

		// Create Global Database Manager for Sessions
		gDBManager = orm.NewDBManager(&dbm)
	}

	return gDBManager, nil
}

func queueConnection() (*queue.AMQPServerConnection, error) {
	if gQConnection == nil {
		// Set Message Activation Queue Connection Settings
		q, err := shared.ToQueue(
			common.ConfigProperty(Config, "queues.default", nil),
		)

		if err != nil {
			return nil, err
		}

		// Create Connection Configuration
		gQConnection = &queue.AMQPServerConnection{}
		gQConnection.SetConnection(q.Servers)
		gQConnection.SetPrefix(q.QueuePrefix)
	}

	return gQConnection, nil
}

func initializeGinSession(c *gin.Context) {
	dbm, err := databaseManager()
	if err != nil {
		panic(err)
	}

	mq, err := queueConnection()
	if err != nil {
		panic(err)
	}

	c.Set("dbm", dbm)
	c.Set("mq-connection", mq)
}

// GIN Router
func ginRouter(r *gin.Engine) *gin.Engine {
	// SESSION
	r.GET("/session", initializeGinSession, pkgsession.Hello) // IMPLEMENTED

	// API Version 1 Interface //
	v1 := r.Group("/1", initializeGinSession) // *gin.RouterGroup
	{
		// SESSION MANAGEMENT //
		session := v1.Group("/session")
		{
			session.POST("/:id", pkgsession.Login) // IMPLEMENTED
			session.DELETE("", pkgsession.Logout)  // IMPLEMENTED
		}

		// PASSWORD MANAGEMENT //
		password := v1.Group("/password")
		{
			password.DELETE("/:email", pkgpwd.Recover)
			password.POST("/:guid", pkgpwd.Reset)
		}

		// INVITATION : NO SESSION REQUIRED //
		invitation := v1.Group("/invitation")
		{
			invitation.GET("/:uid", pkginvites.InvitationNoSessionInfo)
			invitation.POST("/:uid", pkginvites.InvitationAccept)
			invitation.DELETE("/:uid", pkginvites.InvitationDecline)
		}

		// INVITATION MANAGEMENT : SESSION REQUIRED //

		// LIST
		invites := v1.Group("/invites")
		{
			// LIST Invites for Container
			invites.GET("/:object", pkginvites.ListInvitesByObject)
			// LIST All Invites :
			// Probably Shouldn't Exist because:
			// 1. Would Requires System Level Permission to List All Invites
			// 2. We Might Associate the Invite Registry to Object Allow us to spread the invite registry around the shards
			invites.GET("", pkginvites.ListAllInvites)
		}

		// SINGLE
		invite := v1.Group("/invite")
		{
			// CRUD INVITE
			invite.GET("/:id", pkginvites.GetObjectInvite)
			invite.PUT("/:id", pkginvites.ResendInvite)    // IMPLEMENTED: Needs Testing
			invite.DELETE("/:id", pkginvites.DeleteInvite) // IMPLEMENTED: Needs Testing
		}

		// GLOBAL SYSTEM MANAGEMENT //
		system := v1.Group("/system")
		{
			// GLOBAL SYSTEM USERS MANAGEMENT //
			// LIST BASED
			system.GET("/users", pkgsystem.GetUsers) // IMPLEMENTED
			system.DELETE("/users", pkgsystem.DeleteUsers)
			system.PUT("/users/lock/:bool", pkgsystem.PutUsersLock)
			system.PUT("/users/block/:bool", pkgsystem.PutUsersBlock)

			// SINGLE USER
			system.POST("/user", pkgsystem.PostCreateUser)                     // HOW TO? SHOULD? Use Invite?
			system.GET("/user/:user", pkgsystem.GetUserProfile)                // IMPLEMENTED: Needs Testing
			system.DELETE("/user/:user", pkgsystem.DeleteUser)                 // IMPLEMENTED: Tested 20230824
			system.PUT("/user/:user", pkgsystem.PutUserProfile)                // WHAT Options Can the System User Change?
			system.GET("/user/:user/lock", pkgsystem.GetUserLockState)         // IMPLEMENTED: Tested 20230825
			system.GET("/user/:user/block", pkgsystem.GetUserBlockState)       // IMPLEMENTED: Tested 20230825
			system.PUT("/user/:user/lock/:bool", pkgsystem.PutUserLockState)   // IMPLEMENTED: Tested 20230825
			system.PUT("/user/:user/block/:bool", pkgsystem.PutUserBlockState) // IMPLEMENTED: Tested 20230824

			// GLOBAL SYSTEM ORGS MANAGEMENT //
			// LIST BASED
			system.GET("/orgs", pkgsystem.GetOrgs) // IMPLEMENTED
			system.PUT("/orgs/lock/:bool", pkgsystem.PutOrgsLock)
			system.PUT("/orgs/block/:bool", pkgsystem.PutOrgsBlock)

			// SINGLE ORGANIZATION
			system.POST("/org", pkgsystem.PostCreateOrg)                    // IMPLEMENTED
			system.GET("/org/:org", pkgsystem.GetOrgProfile)                // IMPLEMENTED
			system.PUT("/org/:org", pkgsystem.PutOrgProfile)                // IMPLEMENTED
			system.DELETE("/org/:org", pkgsystem.DeleteOrg)                 // IMPLEMENTED: Tested 20230825
			system.GET("/org/:org/lock", pkgsystem.GetOrgLockState)         // IMPLEMENTED: Tested 20230825
			system.GET("/org/:org/block", pkgsystem.GetOrgBlockState)       // IMPLEMENTED: Tested 20230825
			system.PUT("/org/:org/lock/:bool", pkgsystem.PutOrgLockState)   // IMPLEMENTED: Tested 20230825
			system.PUT("/org/:org/block/:bool", pkgsystem.PutOrgBlockState) // IMPLEMENTED: Tested 20230824

			// TEMPLATE ACCESS (LIST / GET)
			system.GET("/templates", pkgsystem.ListTemplates)
			system.GET("/template/:template", pkgsystem.GetTemplate)

			// TODO: Anti Lockout Rule - User Can't modify is own Roles
			// TODO: Anti Lockout Rule - Can't Modify Admin User Roles in System Org??
		}

		// SINGLE ORGANIZATION MANAGEMENT //
		organization := v1.Group("/org/:org")
		{
			// ORGANIZATION
			organization.GET("", pkgorg.Get) // IMPLEMENTED: Needs Testing
			organization.PUT("", pkgorg.Put) // IMPLEMENTED: Needs Testing

			// ORGANIZATION INVITATION
			// LIST: Use GET /invites

			// SINGLE INVITE
			organization.POST("/invite", pkginvites.CreateOrgInvitation) // IMPLEMENTED
			// RESEND: Use PUT /invite/:id
			// DELETE: Use DELETE /invite/:id

			// ORGANIZATION STORE MANAGEMENT //
			// LIST BASED
			organization.GET("/stores", pkgorg.GetOrgStores) // IMPLEMENTED: Needs Testing

			// SINGLE STORE
			organization.POST("/store", pkgorg.PostCreateStore)       // IMPLEMENTED
			organization.GET("/store/:store", pkgorg.GetStoreProfile) // IMPLEMENTED
			organization.DELETE("/store/:store", pkgorg.DeleteStore)
			organization.PUT("/store/:store", pkgorg.PutStoreProfile) // IMPLEMENTED: Needs Testing
			organization.POST("/store/:store/open", pkgstore.OpenStore)
			organization.GET("/store/:store/lock", pkgorg.GetOrgStoreLockState)          // IMPLEMENTED
			organization.PUT("/store/:store/lock/:bool", pkgorg.PutOrgStoreLockState)    // IMPLEMENTED
			organization.GET("/store/:store/block", pkgorg.GetOrgStoreBlockState)        // IMPLEMENTED
			organization.PUT("/store/:store/block/:bool", pkgorg.PutOrgStoreBlockState)  // IMPLEMENTED
			organization.GET("/store/:store/state", pkgorg.GetOrgStoreState)             // IMPLEMENTED
			organization.PUT("/store/:store/state/:uint", pkgorg.PutOrgStoreState)       // IMPLEMENTED
			organization.DELETE("/store/:store/state/:uint", pkgorg.DeleteOrgStoreState) // IMPLEMENTED

			// MASS USER MANAGEMENT
			organization.GET("/users", pkgorg.GetOrgUsers) // IMPLEMENTED
			organization.DELETE("/users", pkgorg.DeleteOrgUsers)
			organization.PUT("/users/lock/:bool", pkgorg.PutOrgUsersLock)
			organization.PUT("/users/block/:bool", pkgorg.PutOrgUsersBlock)
			organization.PUT("/users/roles", pkgorg.PutOrgUsersRoles)

			// SINGLE USER MANAGEMENT
			organization.GET("/user/:user", pkgorg.GetOrgUser)                        // IMPLEMENTED
			organization.DELETE("/user/:user", pkgorg.DeleteOrgUser)                  // IMPLEMENTED: Tested 20230831
			organization.GET("/user/:user/lock", pkgorg.GetOrgUserLock)               // IMPLEMENTED: Needs Testing
			organization.PUT("/user/:user/lock/:bool", pkgorg.PutOrgUserLock)         // IMPLEMENTED: Tested 20230831
			organization.GET("/user/:user/block", pkgorg.GetOrgUserBlock)             // IMPLEMENTED: Needs Testing
			organization.PUT("/user/:user/block/:bool", pkgorg.PutOrgUserBlock)       // IMPLEMENTED: Tested 20230831
			organization.GET("/user/:user/state", pkgorg.GetOrgUserState)             // IMPLEMENTED: Needs Testing
			organization.PUT("/user/:user/state/:uint", pkgorg.PutOrgUserState)       // IMPLEMENTED: Needs Testing
			organization.PUT("/user/:user/admin", pkgorg.ToggleOrgUserAdmin)          // IMPLEMENTED: Needs Testing
			organization.DELETE("/user/:user/state/:uint", pkgorg.DeleteOrgUserState) // IMPLEMENTED: Needs Testing
			organization.PUT("/user/:user/roles", pkgorg.PutOrgUserRoles)

			// ORGANIZATION TEMPLATE ACCESS (LIST / GET)
			organization.GET("/templates/system", pkgorg.ListSystemTemplates)
			organization.GET("/templates", pkgorg.ListTemplates)
			organization.GET("/template/:template", pkgorg.GetTemplate)
			organization.POST("/template/:template", pkgorg.AddTemplateToOrg)
			organization.DELETE("/template/:template", pkgorg.DeleteTemplateFromOrg)
		}

		// STORE MANAGEMENT //
		store := v1.Group("/store/:store")
		{
			// STORE
			store.GET("", pkgstore.GetStore) // IMPLEMENTED: Needs Testing
			store.PUT("", pkgstore.PutStorePUT)

			// STORE OPEN / CLOSE
			store.GET("/open", pkgstore.IsStoreOpen)          // IMPLEMENTED
			store.DELETE("/close", pkgstore.DeleteCloseStore) // IMPLEMENTED

			// STORE INVITATION
			// LIST: Use GET /invites

			// SINGLE INVITE
			store.POST("/invite", pkginvites.CreateStoreInvitation) // IMPLEMENTED
			// DELETE: Use DELETE /invite/:id

			// MASS USER MANAGEMENT
			store.GET("/users", pkgstore.GetStoreUsers) // IMPLEMENTED: Needs Testing
			store.DELETE("/users", pkgstore.DeleteStoreUsers)
			store.PUT("/users/lock", pkgstore.PutStoreUsersLock)
			store.PUT("/users/block", pkgstore.PutStoreUsersBlock)
			store.PUT("/users/roles", pkgstore.PutStoreUsersRoles)

			// SINGLE USER MANAGEMENT
			store.GET("/user/:user", pkgstore.GetStoreUser)                        // IMPLEMENTED
			store.DELETE("/user/:user", pkgstore.DeleteStoreUser)                  // IMPLEMENTED: Tested 20230901
			store.GET("/user/:user/lock", pkgstore.GetStoreUserLock)               // IMPLEMENTED: Needs Testing
			store.PUT("/user/:user/lock/:bool", pkgstore.PutStoreUserLock)         // IMPLEMENTED: Tested 20230901
			store.GET("/user/:user/block", pkgstore.GetStoreUserBlock)             // IMPLEMENTED: Needs Testing
			store.PUT("/user/:user/block/:bool", pkgstore.PutStoreUserBlock)       // IMPLEMENTED: Tested 20230901
			store.GET("/user/:user/state", pkgstore.GetStoreUserState)             // IMPLEMENTED: Needs Testing
			store.PUT("/user/:user/state/:uint", pkgstore.PutStoreUserState)       // IMPLEMENTED: Needs Testing
			store.PUT("/user/:user/admin", pkgstore.ToggleStoreUserAdmin)          // IMPLEMENTED: Needs Testing
			store.DELETE("/user/:user/state/:uint", pkgstore.DeleteStoreUserState) // IMPLEMENTED: Needs Testing
			store.PUT("/user/:user/roles", pkgstore.PutStoreUserRoles)             // IMPLEMENTED

			// STORE OBJECT MANAGEMENT //
			// MASS OBJECT
			store.GET("/objs/:parent", pkgstore.GetStoreObjects) // IMPLEMENTED
			store.DELETE("/objs", pkgstore.DeleteStoreObjects)

			// OBJECT
			store.GET("/obj/:object", pkgstore.GetStoreObject)             // IMPLEMENTED
			store.POST("/obj/:parent", pkgstore.PostStoreObjectJSON)       // IMPLEMENTED
			store.PUT("/obj/:parent/:object", pkgstore.PutStoreObjectJSON) // IMPLEMENTED
			store.DELETE("/obj/:object", pkgstore.DeleteStoreObject)

			// STORE TEMPLATE MANAGEMENT //
			store.GET("/templates", pkgstore.ListStoreTemplates)   // IMPLEMENTED - REQUIRED: List Permission to Store
			store.GET("/template/:template", pkgstore.GetTemplate) // IMPLEMENTED - REQUIRED: Read Permission to Store
			store.POST("/template/:template", pkgstore.AddTemplateToStore)
			store.DELETE("/template/:template", pkgstore.DeleteTemplateFromStore)
		}

		// SELF MANAGEMENT //
		self := v1.Group("/me")
		{
			// USER
			self.GET("", pkgme.GetMe) // IMPLEMENTED: Needs Testing
			self.PUT("", pkgme.PutMe)
			self.DELETE("", pkgme.DeleteMe)

			// PASSWORD Change
			self.POST("/password", pkgme.ChangePassword)

			// LINKS
			self.GET("/objects", pkgme.GetMyObjects)                       // IMPLEMENTED
			self.GET("/objects/:object", pkgme.GetMyObject)                // IMPLEMENTED
			self.GET("/favorites", pkgme.GetMyFavoriteObjects)             // IMPLEMENTED
			self.PUT("/favorite/toggle/:object", pkgme.ToggleLinkFavorite) // IMPLEMENTED

			// ORGANIZATION
			self.GET("/orgs", pkgme.GetMyOrgs) // IMPLEMENTED
			self.DELETE("/org/:org", pkgme.DeleteMeFromOrg)
			self.GET("/org/:org/stores", pkgme.GetMyOrgStores)

			// STORE
			self.GET("/stores", pkgme.GetMyStores) // IMPLEMENTED
			self.DELETE("/store/:store", pkgme.DeleteMeFromStore)
		}
	}

	return r
}
