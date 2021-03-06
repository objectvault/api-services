// cSpell:ignore gonic, orgs, paulo, ferreira
package orm

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
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/objectvault/api-services/common"

	_ "github.com/go-sql-driver/mysql"
)

type DBSessionManager struct {
	config *common.ShardedDatabase
}

// Local Random Number Generator
var randgen *rand.Rand

// Constructor Create an RPF Instance
func NewDBManager(c *common.ShardedDatabase) *DBSessionManager {
	manager := &DBSessionManager{
		config: c,
	}

	return manager
}

func (m *DBSessionManager) Connect(id uint64) (*sql.DB, error) {
	g := common.ShardGroupFromID(id)
	sID := common.ShardFromID(id)
	return m.ConnectTo(g, sID)
}

func (m *DBSessionManager) ConnectTo(g uint16, sID uint32) (*sql.DB, error) {
	shards, err := m.getShardGroup(&m.config.Groups, g)
	if err != nil {
		return nil, err
	}

	var shard *common.DBShard
	shard, err = m.getShard(shards, sID)
	if err != nil {
		return nil, err
	}

	return m.connection(shard.Connection)
}

func (m *DBSessionManager) getShardGroup(rs *([](*common.DBShardGroup)), g uint16) (*common.DBShardGroup, error) {
	if (rs == nil) || (len(*rs) == 0) {
		return nil, errors.New("Missing Shard Groups")
	}

	if int(g) <= len(*rs) {
		return (*rs)[g], nil
	}

	return nil, fmt.Errorf("Shard Group [%d] Does not Exist", g)
}

func (m *DBSessionManager) getShard(sr *common.DBShardGroup, sID uint32) (*common.DBShard, error) {
	if (sr == nil) || (sr.Shards == nil) || (len((*sr).Shards) == 0) {
		return nil, errors.New("No Shards")
	}

	s := (*sr).Shards
	if len(s) == 1 {
		return s[0], nil
	}

	// TODO Look for Shard in Range
	return nil, errors.New("TODO")
}

func (m *DBSessionManager) connection(c common.DBConnection) (*sql.DB, error) {
	// Get Connection Parameters
	database := common.StringNilOnEmpty(c.Database)
	user := common.StringNilOnEmpty(c.User)
	password := common.StringNilOnEmpty(c.Password)
	host := common.StringNilOnEmpty(c.Server.Host)
	port := c.Server.Port

	if (database == nil) || (user == nil) || (host == nil) || (port == 0) {
		return nil, errors.New("Missing Required Database Connection Parameters")
	}

	// Build Connection String
	var conn strings.Builder
	if password == nil {
		fmt.Fprintf(&conn, "%s@tcp(%s:%d)/%s", *user, *host, port, *database)
	} else {
		fmt.Fprintf(&conn, "%s:%s@tcp(%s:%d)/%s", *user, *password, *host, port, *database)
	}

	// Open up our database connection.
	db, err := sql.Open("mysql", conn.String())

	// if there is an error opening the connection, handle it
	if err != nil {
		return nil, err
	}

	// TODO Use Config File  Connection Settings
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)

	return db, nil
}
