package maps

import (
	"encoding/json"
	"errors"
	"strings"
)

/*
 * This file is part of the ObjectVault Project.
 * Copyright (C) 2020-2022 Paulo Ferreira <vault at sourcenotes.org>
 *
 * This work is published under the GNU AGPLv3.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

func Has(m map[string]interface{}, path string) bool {
	// Map created?
	if len(m) == 0 { // NO
		return false
	}

	// Convert Path to a string array
	p, e := pathToPathArray(path)
	if e != nil { // FAILED: Converting path to array
		return false
	}

	// Empty Path?
	if len(p) == 0 { // YES: Return True if Map Created
		return m != nil
	}

	return has(m, p)
}

func Get(m map[string]interface{}, path interface{}) (interface{}, error) {
	// Map created?
	if len(m) == 0 { // NO
		return nil, nil
	}

	// Convert Path to a string array
	p, e := pathToPathArray(path)
	if e != nil { // FAILED: Converting path to array
		return nil, e
	}

	// Empty Path?
	if len(p) == 0 { // YES: Return Whole Map
		return m, nil
	}

	// Parent Exists?
	parent, left := getParent(m, p)
	if len(left) == 0 { // YES: Test if Child Exists
		// Child Name
		key := p[len(p)-1]
		value, exists := parent[key]
		if exists {
			return value, nil
		}
		return nil, nil
	}
	// ELSE: Parent Does not exist, so neither does the child
	return nil, nil
}

func GetDefault(m map[string]interface{}, path string, d interface{}) (interface{}, error) {
	// Map created?
	if len(m) == 0 { // NO
		return d, nil
	}

	// Convert Path to a string array
	p, e := pathToPathArray(path)
	if e != nil { // FAILED: Converting path to array
		return nil, e
	}

	// Empty Path?
	if len(p) == 0 { // YES: Return Whole Map
		return m, nil
	}

	// Parent Exists?
	parent, left := getParent(m, p)
	if len(left) == 0 { // YES: Test if Child Exists
		// Child Name
		key := p[len(p)-1]
		value, exists := parent[key]
		if exists {
			return value, nil
		}
		return d, nil
	}
	// ELSE: Parent Does not exist, so neither does the child
	return d, nil
}

func Set(m map[string]interface{}, path string, v interface{}, force bool) (interface{}, error) {
	// Convert Path to a string array
	p, e := pathToPathArray(path)
	if e != nil { // FAILED: Converting path to array
		return nil, e
	}

	// Empty Path?
	if len(p) == 0 { // YES: Abort
		return nil, errors.New("Missing path")
	}

	// Map created?
	if m == nil { // NO: Create Container for Map
		m = make(map[string]interface{})
	}

	// Create Parent?
	parent, e := createParent(m, p, force)
	if e != nil { // FAILED
		return nil, e
	}

	// Child Name
	key := p[len(p)-1]

	// Get Current Value
	parent[key] = v
	return m, nil
}

func Clear(m map[string]interface{}, path interface{}) (map[string]interface{}, error) {
	// Map created?
	if m == nil { // NO
		return nil, nil
	}

	// Convert Path to a string array
	p, e := pathToPathArray(path)
	if e != nil { // FAILED: Converting path to array
		return nil, e
	}

	// Empty Path?
	if len(p) == 0 { // YES: Clear Whole Map
		return nil, nil
	}

	// Parent Exists?
	parent, left := getParent(m, p)
	if len(left) == 0 { // YES: Test if Child Exists
		// Child Name
		key := p[len(p)-1]
		_, exists := parent[key]
		if exists {
			delete(parent, key)
		}
	}
	// ELSE: Parent or Key Does not exist
	return m, nil
}

func ToJSONString(m map[string]interface{}) (string, error) {
	if len(m) == 0 {
		return "", nil
	}

	// Convert to JSON
	b, e := json.MarshalIndent(m, "", "  ")
	if e != nil {
		return "", e
	}

	return string(b), nil
}

func FromJSONString(s string) (map[string]interface{}, error) {
	// Is String Empty?
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return nil, nil
	}

	// Convert JSON String to map
	m := make(map[string]interface{})
	e := json.Unmarshal([]byte(s), &m)
	if e != nil {
		return nil, e
	}

	// Is map Empty?
	if len(m) == 0 {
		return nil, nil
	}

	return m, nil
}

func Clone(m map[string]interface{}) (map[string]interface{}, error) {
	// Shallow Clone
	return nil, errors.New("TODO Implement")
}

func CloneDeep(m map[string]interface{}) (map[string]interface{}, error) {
	// Deep Clone
	return nil, errors.New("TODO Implement")
}

// INTERNAL //

func getParent(m map[string]interface{}, path []string) (map[string]interface{}, []string) {
	// REQUIRE o.inner != nil
	// REQUIRE len(path) >= 1

	// Parent Path Length (parent:child)
	pp_len := len(path) - 1

	// Is Parent Path empty?
	if pp_len == 0 { // YES: Return ROOT
		return m, nil
	}

	p := m
	for i := 0; i < pp_len; i++ {
		n, exists := p[path[i]]
		if !exists {
			return p, path[i:pp_len]
		}

		m, ok := n.(map[string]interface{})
		if !ok {
			return p, path[i:pp_len]
		}

		p = m
	}

	return p, nil
}

func createParent(m map[string]interface{}, path []string, force bool) (map[string]interface{}, error) {
	parent, left := getParent(m, path)

	if len(left) == 0 {
		return parent, nil
	}

	parent, e := makeParent(parent, left, force)
	if e != nil {
		return nil, e
	}

	return parent, nil
}

func makeParent(m map[string]interface{}, left []string, force bool) (map[string]interface{}, error) {
	var parent map[string]interface{}
	var name string
	var node interface{}
	var exists, ok bool

	// Descend while creating
	parent = m
	for i := 0; i < len(left); i++ {
		// Next Node Name
		name = strings.TrimSpace(left[i])
		if len(name) == 0 {
			return nil, errors.New("Invalid Node Name")
		}

		// Does Node Exist?
		node, exists = parent[name]
		if exists { // YES: Verify if Valid Node
			parent, ok = node.(map[string]interface{})
			if !ok && !force {
				return nil, errors.New("Node is not a map")
			}
			continue
		}

		// Create New Node
		n := make(map[string]interface{})
		parent[name] = n
		parent = n
	}

	return parent, nil
}

func has(m map[string]interface{}, path []string) bool {
	// Parent Exists?
	parent, left := getParent(m, path)
	if len(left) == 0 { // YES: Test if Child Exists
		// Child Name
		key := path[len(path)-1]
		_, exists := parent[key]
		return exists
	}
	// ELSE: Not Found
	return false
}

func pathToPathArray(p interface{}) ([]string, error) {
	// Path Provided?
	if p == nil { // NO: Abort
		return nil, nil
	}

	// NOTE: Empty Path Elements Should be Handled in Map Functions
	var path []string
	switch v := p.(type) {
	case string:
		sp := strings.TrimSpace(v)
		if len(sp) == 0 { // YES: Abort
			return nil, nil
		}
		// ELSE: Split String into Path Components
		path = strings.Split(sp, ".")
	case []string:
		path = v
		// Is Empty Array?
		if len(path) == 0 { // YES: Abort
			return nil, nil
		}
	default:
		return nil, errors.New("Invalid Value for Path")
	}

	return path, nil
}
