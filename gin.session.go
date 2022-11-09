// cSpell:ignore ginrpf, gonic, paulo ferreira
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

// cSpell:ignore keypairs, staticcheck

import (
	"crypto/sha256"
	"log"
	"strings"

	"github.com/objectvault/api-services/common"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// Initialize Session Store
func InitializeSessionStore(r *gin.Engine) bool {
	var store sessions.Store

	// Get the Configuration Options for the Store
	sessionSettings := common.ConfigProperty(Config, "session.store", nil).(map[string]interface{})
	cookieSettings := common.ConfigProperty(sessionSettings, "cookie", nil).(map[string]interface{})

	// Get Store Type
	storeType := common.ConfigProperty(sessionSettings, "type", "cookie")

	// Get Store ClientCookie ID
	cookieID := common.ConfigProperty(cookieSettings, "id", "__sid")

	// Secure Cookie HASH Key (SALT for Authentication)
	secret := common.ConfigProperty(cookieSettings, "secret", nil)
	if secret == nil {
		secret = "**HASH-KEY-REQUIRED**"
		log.Println("[InitializeSessionStore] No Key Set for Cookie Security")
	}
	// Create Keypairs for Gorilla Cookies secure cookie
	// SEE: Links to See how keypairs us used (Basically half the array is used for authentication)
	// half is used for encryption)
	// session store: https://github.com/gorilla/sessions/blob/master/store.go
	// secure cookie: https://github.com/gorilla/securecookie/blob/master/securecookie.go

	// Convert Secret to a Hash for More Security
	keypairs := sha256.Sum256([]byte(secret.(string)))

	// Create Store based on Type
	switch storeType {
	case "cookie":
		store = cookie.NewStore(keypairs[:])
		options := common.ConfigProperty(cookieSettings, "options", nil).(map[string]interface{})

		// Set Defaults Cookie Options
		cookieOptions := sessions.Options{
			Path: "/",
		}

		// Convert Cookie Options to Structure
		// NOTE: staticcheck S1031 - Loop will not execute if options == nil
		for k, v := range options {
			// Use Lowe Case for Switch
			k = strings.ToLower(k)

			switch k {
			case "path":
				path := strings.TrimSpace(v.(string))
				if path != "" {
					cookieOptions.Path = path
				}
			case "domain":
				domain := strings.TrimSpace(v.(string))
				if domain != "" {
					cookieOptions.Domain = domain
				}
			case "maxage":
				cookieOptions.MaxAge = int(v.(float64))
			case "secure":
				cookieOptions.Secure = v.(bool)
			case "httponly":
				cookieOptions.HttpOnly = v.(bool)
			}
		}

		// Set Cookie Options
		store.Options(cookieOptions)
	default:
		// TODO Unknown Session Store Type
	}

	if store == nil {
		// TODO Log Error
		return false
	}

	/* TODO Session Cookie Options
	   * Path:     options.Path,
		 * Domain:   options.Domain,
		 * MaxAge:   options.MaxAge,
		 * Secure:   options.Secure,
	   * HttpOnly: options.HttpOnly,
	*/
	r.Use(sessions.Sessions(cookieID.(string), store))
	return true
}
