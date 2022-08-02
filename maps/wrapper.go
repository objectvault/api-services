package maps

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
)

type MapWrapper struct {
	inner    map[string]interface{}
	modified bool
}

func NewMapWrapper(m map[string]interface{}) *MapWrapper {
	v := &MapWrapper{}

	// Is Incoming Map Set?
	if m != nil { // YES
		// TODO: Should we clone the incoming map (deep, shallow)
		v.inner = m
	}

	return v
}

func (o *MapWrapper) IsEmpty() bool {
	return len(o.inner) == 0
}

func (o *MapWrapper) IsModified() bool {
	return o.modified
}

func (o *MapWrapper) Map() *map[string]interface{} {
	// Is Map Empty?
	if len(o.inner) == 0 {
		return nil
	}

	return &o.inner
}

func (o *MapWrapper) Has(path string) bool {
	return Has(o.inner, path)
}

func (o *MapWrapper) Get(path interface{}) (interface{}, error) {
	return Get(o.inner, path)
}

func (o *MapWrapper) GetDefault(path string, d interface{}) (interface{}, error) {
	return GetDefault(o.inner, path, d)
}

func (o *MapWrapper) Set(path string, v interface{}, force bool) (interface{}, error) {
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
	if o.inner == nil { // NO: Create Container for Map
		o.inner = make(map[string]interface{})
	}

	// Create Parent?
	parent, e := createParent(o.inner, p, force)
	if e != nil { // FAILED
		return nil, e
	}

	// Child Name
	key := p[len(p)-1]

	// Get Current Value
	current, exists := parent[key]
	parent[key] = v
	o.modified = true

	// Do we have a current value?
	if exists { // YES: Return it
		return current, nil
	}
	// ELSE: No Current Value
	return nil, nil
}

func (o *MapWrapper) Clear(path interface{}) (interface{}, error) {
	// Map created?
	if o.inner == nil { // NO
		return nil, nil
	}

	// Convert Path to a string array
	p, e := pathToPathArray(path)
	if e != nil { // FAILED: Converting path to array
		return nil, e
	}

	// Empty Path?
	if len(p) == 0 { // YES: Clear Whole Map
		c := o.inner
		o.inner = nil
		o.modified = true
		return c, nil
	}

	// Parent Exists?
	parent, left := getParent(o.inner, p)
	if len(left) == 0 { // YES: Test if Child Exists
		// Child Name
		key := p[len(p)-1]
		value, exists := parent[key]
		if exists {
			delete(parent, key)
			o.modified = true
			return value, nil
		}
	}
	// ELSE: Parent or Key Does not exist
	return nil, nil
}

func (o *MapWrapper) ClearModified(path interface{}) bool {
	c := o.modified
	o.modified = false
	return c
}

func (o *MapWrapper) Import(s string) error {
	m, e := FromJSONString(s)
	if e == nil {
		o.inner = m
		o.modified = true
	}
	return e
}

func (o *MapWrapper) Export() string {
	json, _ := ToJSONString(o.inner)
	return json
}

func (o *MapWrapper) Reset() *MapWrapper {
	o.inner = nil
	o.modified = false
	return o
}
