// cSpell:ignore gonic, orgs, paulo, ferreira
package common

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
	"errors"
	"strings"
)

// COMMON CONFIGURATION STRUCTURES and HELPERS //

// A Server TCP/IP Connection Addres and Port
type Server struct {
	Host string `json:"host,omitempty"`
	Port uint16 `json:"port,omitempty"`
}

func (s *Server) FromConfig(base map[string]interface{}) error {
	var e error
	s.Host, e = ConfigPropertyString(base, "host", "", nil)
	ui, e := ConfigPropertyUINT(base, "port", 0, e)
	if e != nil {
		return e
	}
	s.Port = uint16(ui)
	return nil
}

// DEFINITION: Database Connection
type DBConnection struct {
	Database string                 `json:"database"`
	User     string                 `json:"user,omitempty"`
	Password string                 `json:"password,omitempty"`
	Server   *Server                `json:"server"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

func (s *DBConnection) FromConfig(base map[string]interface{}) error {
	var e error
	s.Database, e = ConfigPropertyString(base, "database", "", nil)
	s.User, e = ConfigPropertyString(base, "user", "", e)
	s.Password, e = ConfigPropertyString(base, "password", "", e)
	s.Options, e = ConfigPropertyObject(base, "options", nil, e)

	// Was there an error extracting parameters
	if e != nil { // YES
		return e
	}

	// Do we Have Minimum Set of Parameters
	if (s.Database == "") || (s.User == "") { // NO
		return errors.New("Shard DB Connection Missing Required Parameters")
	}

	// Do we have a Server Object?
	o, e := ConfigPropertyObject(base, "server", nil, nil)
	if e != nil { // YES
		return e
	}

	// Is Server Object Valid?
	s.Server = &Server{}
	err := s.Server.FromConfig(o)
	if err != nil { // NO
		return err
	}
	return nil
}

// DEFINITION: Database Shard
type DBShard struct {
	Range      []uint32     `json:"range"`
	Connection DBConnection `json:"connection"`
}

func (s *DBShard) FromConfig(base map[string]interface{}) error {
	var err error

	// EXTRACT Range for Shard
	v_ai := ConfigProperty(base, "range", nil).([]interface{})
	if v_ai == nil {
		return errors.New("Missing Range for Shard")
	}

	var start, end uint32
	var f float64
	var ok bool
	switch len(v_ai) {
	case 0:
		return errors.New("Missing Range for Shard")
	case 1:
		f, ok = v_ai[0].(float64)
		if !ok {
			return errors.New("Invalid Range for Shard")
		}
		end = uint32(f)
	default:
		f, ok = v_ai[0].(float64)
		if !ok {
			return errors.New("Invalid Start of Range for Shard")
		}
		start = uint32(f)

		f, ok = v_ai[0].(float64)
		if !ok {
			return errors.New("Invalid End of Range for ShardRange")
		}
		end = uint32(f)
	}
	s.Range = append(s.Range, start)
	s.Range = append(s.Range, end)

	// Do we have a Valid Connection Object?
	o, e := ConfigPropertyObject(base, "connection", nil, nil)
	if e != nil { // NO
		return e
	}

	// Is Connection for Shard Valid?
	s.Connection = DBConnection{}
	err = s.Connection.FromConfig(o)
	if err != nil {
		return err
	}
	return nil
}

// DEFINITION: Database Shard Group
type DBShardGroup struct {
	Shards [](*DBShard) `json:"shards"`
}

func (s *DBShardGroup) FromConfig(base map[string]interface{}) error {
	var err error

	// EXTRACT Shards for Range
	v_as := ConfigProperty(base, "shards", nil)
	if v_as == nil {
		return errors.New("Missing Shards List")
	}

	var n *DBShard
	for _, rv := range v_as.([]interface{}) {
		n = &DBShard{}
		err = n.FromConfig(rv.(map[string]interface{}))
		if err != nil {
			return err
		}
		s.Shards = append(s.Shards, n)
	}

	if len(s.Shards) == 0 {
		return errors.New("NO ShardRanges Configured")
	}

	return nil
}

// DEFINITION: Sharded Database Definition
type ShardedDatabase struct {
	Groups [](*DBShardGroup) `json:"shard-groups"`
}

func (s *ShardedDatabase) FromConfig(base map[string]interface{}) error {
	var err error
	v_asr, ok := ConfigProperty(base, "shard-groups", nil).([]interface{})
	if v_asr == nil || !ok {
		return errors.New("Missing Shard Ranges")
	}

	var sr *DBShardGroup
	for _, r := range v_asr {
		sr = &DBShardGroup{}
		err = sr.FromConfig(r.(map[string]interface{}))
		if err != nil {
			return err
		}
		s.Groups = append(s.Groups, sr)
	}

	if len(s.Groups) == 0 {
		return errors.New("NO ShardRanges Configured")
	}

	return nil
}

// COMMON CONFIGURATION HELPERS //
func getChildProperty(source map[string]interface{}, elements []string, i int, dvalue interface{}) interface{} {
	if i >= len(elements) {
		return source
	}

	element := strings.TrimSpace(elements[i])
	if len(element) > 0 {
		value, exists := source[elements[i]]
		if exists {
			switch v := value.(type) {
			case map[string]interface{}:
				return getChildProperty(v, elements, i+1, dvalue)
			default:
				return v
			}
		}
	}

	return dvalue
}

func ConfigProperty(source map[string]interface{}, path string, dvalue interface{}) interface{} {
	elements := strings.Split(path, ".")
	if source == nil {
		return dvalue
	}

	element := strings.TrimSpace(elements[0])
	if len(element) > 0 {
		value, exists := source[elements[0]]
		if exists {
			switch v := value.(type) {
			case map[string]interface{}:
				return getChildProperty(v, elements, 1, dvalue)
			default:
				return v
			}
		}
	}

	return dvalue
}

func ConfigPropertyObject(source map[string]interface{}, path string, dvalue map[string]interface{}, nested error) (map[string]interface{}, error) {
	if nested != nil {
		return nil, nested
	}

	if source != nil {
		v := ConfigProperty(source, path, dvalue)
		if v != nil {
			v_o, ok := v.(map[string]interface{})
			if ok {
				return v_o, nil
			} else {
				return nil, errors.New("[" + path + "] is not an object")
			}
		}
	}
	return dvalue, nil
}

func ConfigPropertyString(source map[string]interface{}, path string, dvalue string, nested error) (string, error) {
	if nested != nil {
		return "", nested
	}

	if source != nil {
		v := ConfigProperty(source, path, dvalue)
		if v != nil {
			v_s, ok := v.(string)
			if ok {
				return v_s, nil
			} else {
				return "", errors.New("[" + path + "] is not a string")
			}
		}
	}

	return dvalue, nil
}

func ConfigPropertyINT(source map[string]interface{}, path string, dvalue int64, nested error) (int64, error) {
	f, e := ConfigPropertyFloat64(source, path, 0, nested)
	if e != nil {
		return 0, nested
	}

	if f == 0 {
		return dvalue, nil
	}

	return int64(f), nil
}

func ConfigPropertyUINT(source map[string]interface{}, path string, dvalue uint64, nested error) (uint64, error) {
	f, e := ConfigPropertyFloat64(source, path, 0, nested)
	if e != nil {
		return 0, nested
	}

	if f == 0 {
		return dvalue, nil
	}

	return uint64(f), nil
}

func ConfigPropertyFloat64(source map[string]interface{}, path string, dvalue float64, nested error) (float64, error) {
	if nested != nil {
		return 0, nested
	}

	if source != nil {
		v := ConfigProperty(source, path, dvalue)
		if v != nil {
			v_i, ok := v.(float64)
			if ok {
				return v_i, nil
			} else {
				return 0, errors.New("[" + path + "] is not an integer")
			}
		}
	}

	return dvalue, nil
}
