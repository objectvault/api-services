// cSpell:ignore gonic, paulo, ferreira
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
	"encoding/json"
	"fmt"
	"os"

	"github.com/objectvault/api-services/common"
)

// CONTAINER for SERVER CONFIGURATION (GENERIC)
var Config map[string]interface{}

// START: DEFINITION OF SERVER CONFIGURATION FILE //

// Server Session Store Configuration
type CookieStore struct {
	ID            string `json:"id"`
	KeyEncryption string `json:"encryption,omitempty"`
	KeyHash       string `json:"hash"`
}

type RedisStore struct {
	Protocol string `json:"protocol,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Database string `json:"database,omitempty"`
	Password string `json:"password,omitempty"`
}

type SessionStore struct {
	Storetype string      `json:"type"`
	Cookie    CookieStore `json:"cookie"`
	Redis     RedisStore  `json:"redis,omitempty"`
}

type Session struct {
	Store SessionStore `json:"store"`
}

type ServerConfig struct {
	BindAddress *common.Server          `json:"bind,omitempty"`
	Session     *Session                `json:"session,omitempty"`
	Database    *common.ShardedDatabase `json:"database,omitempty"`
}

// END: DEFINITION OF SERVER CONFIGURATION FILE //

// Load Configuration File
func loadConfiguration(path string) {
	// Open Configuration File
	file, errFile := os.Open(path)
	if errFile != nil {
		fmt.Printf("Error [%s]\n", errFile)
		fmt.Println("ERROR: Configuration File Required")
		os.Exit(1)
	}
	defer file.Close()

	// Decode JSON
	decoder := json.NewDecoder(file)
	errDecoder := decoder.Decode(&Config)
	if errDecoder != nil {
		fmt.Printf("JSON Parse Error [%s]\n", errDecoder)
		fmt.Println("ERROR: Invalid Configuration File")
		os.Exit(2)
	}
}
