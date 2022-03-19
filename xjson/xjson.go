package xjson

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
	"errors"
	"fmt"
	"strings"
)

type T_xMap = map[string]interface{}

// Standard Process Handler Function
type T_xValueValidator func(interface{}) (interface{}, error)

// Standard On Value Handler Function
type T_xOnValueHandler func(interface{}) error

// STANDARD Value Mappers
func F_xToBoolean(v interface{}) (interface{}, error) {
	var b bool

	b, ok := v.(bool)
	if !ok {
		i, e := F_xToInt64(v)
		if e != nil {
			return nil, errors.New("Value is not a boolean")
		}

		b = i != 0
	}

	return b, nil
}

func F_xToString(v interface{}) (interface{}, error) {
	s, ok := v.(string)
	if !ok {
		return nil, errors.New("Value is invalid")
	}

	return s, nil
}

func F_xToTrimmedString(v interface{}) (interface{}, error) {
	s, ok := v.(string)
	if !ok {
		return nil, errors.New("Value is invalid")
	}

	s = strings.TrimSpace(s)
	return s, nil
}

func F_xToInt64(v interface{}) (interface{}, error) {
	f, ok := v.(float64)
	if !ok {
		return nil, errors.New("Value is invalid")
	}

	i := int64(f)
	return i, nil
}

func F_xToUint64(v interface{}) (interface{}, error) {
	vi, e := F_xToInt64(v)
	if e != nil {
		return nil, e
	}

	i := vi.(int64)
	if i < 0 {
		return nil, errors.New("Value is not a positive integer")
	}

	u := uint64(i)
	return u, nil
}

type S_xJSONMap struct {
	Source      T_xMap // Original Map
	Destination T_xMap // New Map
	Error       error  // Last Error
}

func (j *S_xJSONMap) ClearError() *S_xJSONMap {
	j.Error = nil
	return j
}

func (j *S_xJSONMap) StringSrc() string {
	return jsonString(j.Source)
}

func (j *S_xJSONMap) StringDst() string {
	return jsonString(j.Destination)
}

func (j *S_xJSONMap) GetSourceValue(p interface{}) (interface{}, error) {
	// Any Processing Errors?
	if j.Error != nil { // YES: Abort
		return nil, j.Error
	}

	// Do we have a Source Map?
	if j.Source == nil { // NO: Abort
		return nil, errors.New("Missing Source Map")
	}

	// Do we have a Path Value?
	path, e := pathToPathArray(p)
	if e != nil { // NO: Abort
		return nil, e
	}

	return getValue(j.Source, path, false)
}

func (j *S_xJSONMap) HasValue(p interface{}) bool {
	// Do we have a Source Map?
	if j.Destination == nil { // NO: Abort
		return false
	}

	// Do we have a Path Value?
	path, e := pathToPathArray(p)
	if e != nil { // NO: Abort
		return false
	}

	return hasValue(j.Destination, path)
}

func (j *S_xJSONMap) OnValue(p interface{}, handler T_xOnValueHandler) error {
	// Do we have a Source Map?
	if j.Destination == nil { // NO: Abort
		return errors.New("Missing Destination Map")
	}

	// Do we have a Path Value?
	path, e := pathToPathArray(p)
	if e != nil { // NO: Abort
		return e
	}

	// Has Value?
	value, e := getValue(j.Destination, path, true)
	if e != nil { // NO: Abort
		return e
	}
	// ELSE: Call Handler with Value
	return handler(value)
}

func (j *S_xJSONMap) GetValue(p interface{}) (interface{}, error) {
	// Do we have a Source Map?
	if j.Destination == nil { // NO: Abort
		return nil, errors.New("Missing Destination Map")
	}

	// Do we have a Path Value?
	path, e := pathToPathArray(p)
	if e != nil { // NO: Abort
		return nil, e
	}

	return getValue(j.Destination, path, false)
}

func (j *S_xJSONMap) SetValue(p interface{}, v interface{}) *S_xJSONMap {
	// Any Processing Errors?
	if j.Error != nil { // YES: Abort
		return j
	}

	// Do we have a Path Value?
	path, e := pathToPathArray(p)
	if e != nil { // NO: Abort
		j.Error = e
		return j
	}

	// Is Destination Empty?
	if j.Destination == nil { // YES: Make sure it has a space
		j.Destination = make(map[string]interface{})
	}

	j.Error = setValue(j.Destination, path, v)
	return j
}

func (j *S_xJSONMap) Required(src interface{}, dst interface{}, val T_xValueValidator, on T_xOnValueHandler) *S_xJSONMap {
	// Any Processing Errors?
	if j.Error != nil { // YES: Abort
		return j
	}

	// Do we have a Source Map?
	if j.Source == nil { // NO: Abort
		j.Error = errors.New("Missing Source Map")
		return j
	}

	// Do we have a Path Value?
	path, e := pathToPathArray(src)
	if e != nil { // NO: Abort
		j.Error = e
		return j
	}

	// Do we have a Source Value?
	v, e := getValue(j.Source, path, true)
	if e != nil { // NO: Abort
		j.Error = e
		return j
	}

	// Do we have a Validator?
	if val != nil { // YES: Use IT
		// Is Value Valid?
		v, e = val(v)
		if e != nil { // NO: Abort
			j.Error = e
			return j
		}
	}

	// Do we have a Destination Path?
	if dst != nil { // YES: Use IT
		// Do we have a Path Value?
		path, e = pathToPathArray(dst)
		if e != nil { // NO: Abort
			j.Error = e
			return j
		}
	}
	// ELSE: No Destination Path == Source Path

	// Is Destination Empty?
	if j.Destination == nil { // YES: Make sure it has a space
		j.Destination = make(map[string]interface{})
	}

	j.Error = setValue(j.Destination, path, v)

	// Do we have On Apply Handler?
	if on != nil { // YES: Call it
		j.Error = on(v)
	}

	return j
}

func (j *S_xJSONMap) Optional(src interface{}, dst interface{}, val T_xValueValidator, d interface{}, on T_xOnValueHandler) *S_xJSONMap {
	// Any Processing Errors?
	if j.Error != nil { // YES: Abort
		return j
	}

	// Do we have a Source Map?
	if j.Source == nil { // NO: Abort
		j.Error = errors.New("Missing Source Map")
		return j
	}

	// Do we have a Path Value?
	path, e := pathToPathArray(src)
	if e != nil { // NO: Abort
		j.Error = e
		return j
	}

	// Do we have a Source Value?
	v, e := getValue(j.Source, path, false)
	if e != nil { // NO: Abort
		j.Error = e
		return j
	}

	// Do we have a Value and Validator?
	if v != nil && val != nil { // YES: Use IT
		// Is Value Valid?
		v, e = val(v)
		if e != nil { // NO: Abort
			j.Error = e
			return j
		}
	}

	// Do we need to use Default Value?
	if v == nil && d != nil { // YES: Use IT
		v = d
	}

	// Do we have a value to set?
	if v != nil { // YES
		// Do we have a Destination Path?
		if dst != nil { // YES: Use IT
			// Do we have a Path Value?
			path, e = pathToPathArray(dst)
			if e != nil { // NO: Abort
				j.Error = e
				return j
			}
		}
		// ELSE: No Destination Path == Source Path

		// Is Destination Empty?
		if j.Destination == nil { // YES: Make sure it has a space
			j.Destination = make(map[string]interface{})
		}

		j.Error = setValue(j.Destination, path, v)
	}

	// Do we have On Apply Handler?
	if on != nil { // YES: Call it
		j.Error = on(v)
	}

	return j
}

func pathToPathArray(p interface{}) ([]string, error) {
	// Path Provided?
	if p == nil { // NO: Abort
		return nil, errors.New("Missing Path Value")
	}

	var path []string

	switch v := p.(type) {
	case string:
		sp := strings.TrimSpace(v)
		if len(sp) == 0 { // YES: Abort
			return nil, errors.New("Invalid Value for Path")
		}
		// ELSE: Split String into Path Components
		path = strings.Split(sp, ".")
	case []string:
		path = v
		// Is Empty Array?
		if len(path) == 0 { // YES: Abort
			return nil, errors.New("Missing Path Value")
		}
	default:
		return nil, errors.New("Invalid Value for Path")
	}

	return path, nil
}

func jsonString(m T_xMap) string {
	if m == nil {
		return "NIL"
	} else {
		b, e := json.MarshalIndent(m, "", "  ")
		if e != nil {
			msg := fmt.Sprintf("ERROR: %s", e)
			return msg
		}
		return string(b)
	}
}

func hasValue(d T_xMap, path []string) bool {
	// Key
	key := path[len(path)-1]

	// Can we get Parent?
	parent, left := getDeepestMap(d, path[:len(path)-1])
	if len(left) == 0 { // YES: Abort
		return hasChildValue(parent, key)
	}
	// ELSE: Not Found
	return false
}

func getValue(d T_xMap, path []string, errorOnNotFound bool) (interface{}, error) {
	// Key
	key := path[len(path)-1]

	// Can we get Parent?
	parent, left := getDeepestMap(d, path[:len(path)-1])
	if len(left) == 0 { // YES: Abort
		return getChildValue(parent, key, errorOnNotFound)
	}

	if errorOnNotFound {
		msg := fmt.Sprintf("[%s] does not exist in map", key)
		return nil, errors.New(msg)
	}
	// ELSE: Not Found
	return nil, nil
}

func setValue(d T_xMap, path []string, v interface{}) error {
	// Can we Create the Parent Map?
	parent, e := createParentMap(d, path[:len(path)-1])
	if e != nil { // NO: Abort
		return e
	}

	// Set Value in Parent Map
	key := path[len(path)-1]
	return setParentValue(parent, key, v)
}

func hasChildValue(parent T_xMap, key string) bool {
	_, exists := parent[key]
	return exists
}

func getChildValue(parent T_xMap, key string, errorOnNotFound bool) (interface{}, error) {
	// Does Parent have Value?
	value, exists := parent[key]
	if exists { // YES: Done
		return value, nil
	}

	if errorOnNotFound {
		msg := fmt.Sprintf("[%s] does not exist in map", key)
		return nil, errors.New(msg)
	}
	// ELSE: Not Found
	return nil, nil
}

func getDeepestMap(s T_xMap, path []string) (T_xMap, []string) {
	// Path to Parent?
	if len(path) == 0 { // YES
		return s, path
	}

	p := s
	for i := 0; i < len(path); i++ {
		n, exists := p[path[i]]
		if !exists {
			return p, path[i:]
		}

		m, ok := n.(T_xMap)
		if !ok {
			return p, path[i:]
		}

		p = m
	}

	return p, nil
}

func getParentMap(s T_xMap, path []string) (T_xMap, error) {
	var e error
	p := s
	for i := 0; i < len(path); i++ {
		p, e = getChildMap(p, path[i])
		if e != nil {
			return nil, e
		}
	}

	return p, nil
}

func getChildMap(s T_xMap, name string) (T_xMap, error) {
	value, exists := s[name]
	if !exists {
		return nil, nil
	}

	m, ok := value.(T_xMap)
	if !ok {
		msg := fmt.Sprintf("[%s] is not a map entry", name)
		return nil, errors.New(msg)
	}

	return m, nil
}

func createMap(p T_xMap, name string) (T_xMap, error) {
	value, exists := p[name]
	if !exists {
		m := make(T_xMap)
		p[name] = m
		return m, nil
	}

	m, ok := value.(T_xMap)
	if ok {
		return m, nil
	}

	msg := fmt.Sprintf("[%s] is not a map entry", name)
	return nil, errors.New(msg)
}

func createParentMap(d T_xMap, path []string) (T_xMap, error) {
	// Is Root Path?
	if len(path) == 0 { // YES
		return d, nil
	}

	// Get Depest Map
	dp, r := getDeepestMap(d, path)

	// Try to Create Child Map
	var e error
	p := dp
	for i := 0; i < len(r); i++ {
		p, e = createMap(p, r[i])
		if e != nil {
			return nil, e
		}
	}

	return p, e
}

func setParentValue(d T_xMap, name string, v interface{}) error {
	d[name] = v
	return nil
}
