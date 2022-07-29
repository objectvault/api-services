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

// cSpell:ignore ccors, gindump

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/objectvault/api-services/common"
)

func ginEngine() *gin.Engine {
	// Equivalent to gin.Default()
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	// TODO: Fix CORS - For Now Use Default Allow All
	ccors := cors.DefaultConfig()
	ccors.AllowOriginFunc = func(origin string) bool {
		return true
	}
	ccors.AllowCredentials = true
	engine.Use(cors.New(ccors))

	// Initialize Session Store
	if !InitializeSessionStore(engine) {
		log.Println("[ginEngine] Failed to Initialize Session Store")
		return nil
	}

	// START:DEBUG
	// r.Use(gindump.Dump())
	// END:DEBUG

	return engine
}

// MAIN //
func main() {
	// COMMAND LINE PARSER //
	flag.Usage = func() {
		usage := `
		Object Vault API Service

		Usage:
		  server -c /path/to/conf
		  server -v | --version
		  server -h | --help

		  Options:
		    -h --help     Show this screen.
		    -v            Show version.
		    -c            Path to configuration file [default: ./server.json].
		`

		fmt.Println(usage)
	}
	sConfPath := flag.String("c", "./server.json", "Path to configuration file")
	bVersion := flag.Bool("v", false, "Path to configuration file")
	flag.Parse()

	// Version Flag Set?
	if *bVersion { // YES: Display Version and Exit
		fmt.Print("Object Vault API Service [0.0.1]\n")
		os.Exit(0)
	}

	// Load Configuration File
	loadConfiguration(*sConfPath)

	// After everything is Done Make Sure to Close Everything
	defer func() {
		fmt.Println("EXIT: Close All Connections")
	}()

	// Create and Configure Gin Engine
	r := ginEngine() // *gin.Engine
	if r == nil {
		panic("Failed to Initialize GIN Engine")
	}

	// Establish Routes
	ginRouter(r)

	// Run Web Server //
	// BUILD Listen Address from Server Configuration //
	address := common.ConfigProperty(Config, "bind.host", "")
	/* DEFAULT = 3000.0 not 3000 because json decoder converts
	 * "port": 3000 to "port": (float64)(3000) and not int
	 */
	port := common.ConfigProperty(Config, "bind.port", 3000.0)
	var listen strings.Builder
	fmt.Fprintf(&listen, "%s:%d", address.(string), int64(port.(float64)))

	// Run Server
	r.Run(listen.String())
}
