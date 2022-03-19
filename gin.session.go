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

import (
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
	keyHASH := common.ConfigProperty(cookieSettings, "hash", nil)
	if keyHASH == nil {
		keyHASH = "**HASH-KEY-REQUIRED**"
		log.Println("[InitializeSessionStore] No Key Set for Cookie Hashing")
	}
	hash := []byte(keyHASH.(string))

	// Secure Cookie Encryption String
	keyEncryption := common.ConfigProperty(cookieSettings, "encryption", nil)
	var secure []byte
	if keyEncryption != nil {
		secure = []byte(keyEncryption.(string))
	} else {
		log.Println("[InitializeSessionStore] No Key Set for Cookie Encryption")
	}

	// Create Store based on Type
	switch storeType {
	case "cookie":
		store = cookie.NewStore(hash, secure)
		options := common.ConfigProperty(cookieSettings, "options", nil).(map[string]interface{})

		// Set Defaults Cookie Options
		cookieOptions := sessions.Options{
			Path: "/",
		}

		// Do we have Cookie Options
		if options != nil { // YES

			// Convert Cookie Options to Structure
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
