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
	"encoding/json"
	"errors"
	"strings"
)

// Templated Object for Storage
type StoreTemplateObject struct {
	template string                 // TEMPLATE: Name
	version  uint16                 // TEMPLATE: Version
	title    string                 // OBJECT Title
	values   map[string]interface{} // OBJECT TEMPLATED Values
}

func (o *StoreTemplateObject) isValid() bool {
	return o.template != "" && o.version > 0 && o.title != "" && o.values != nil
}

func (o *StoreTemplateObject) Template() string {
	return o.template
}

func (o *StoreTemplateObject) Version() uint16 {
	return o.version
}

func (o *StoreTemplateObject) Title() string {
	return o.title
}

func (o *StoreTemplateObject) Values() interface{} {
	return o.values
}

func (o *StoreTemplateObject) SetTemplate(v string) (string, error) {
	v = strings.TrimSpace(v)
	if v != "" {
		// Current State
		current := o.template

		// New State
		o.template = strings.ToLower(v)
		return current, nil
	}

	return "", errors.New("Template Name is Invalid")
}

func (o *StoreTemplateObject) SetVersion(v uint16) (uint16, error) {
	if v > 0 {
		// Current State
		current := o.version

		// New State
		o.version = v
		return current, nil
	}

	return 0, errors.New("Template Version is Invalid")
}

func (o *StoreTemplateObject) SetTitle(v string) (string, error) {
	v = strings.TrimSpace(v)
	if v != "" {
		// Current State
		current := o.title

		// New State
		o.title = v
		return current, nil
	}

	return "", errors.New("Object Title is Invalid")
}

func (o *StoreTemplateObject) SetValues(v map[string]interface{}) (interface{}, error) {
	if v != nil {
		// Current State
		current := o.values

		// New State
		o.values = v
		return current, nil
	}

	return nil, errors.New("Object Value Missing")
}

func (o *StoreTemplateObject) ToString() (string, error) {
	bs, e := o.MarshalJSON()
	if e != nil {
		return "", e
	}
	return string(bs), e
}

func (o *StoreTemplateObject) FromString(s string) error {
	// Extract JSON from String
	if s != "" {
		return o.UnmarshalJSON([]byte(s))
	}

	return nil
}

func (o *StoreTemplateObject) EncryptObject(key []byte) ([]byte, error) {
	// Convert Object to JSON String
	json, e := o.ToString()
	if e != nil {
		return nil, e
	}

	if len(json) > 65535 {
		return nil, errors.New("Object too big")
	}

	// Encrypt the JSON String
	encrypted, e := toCypherBytes(key, []byte(json))
	if e != nil {
		return nil, e
	}

	if len(encrypted) > 65535 {
		return nil, errors.New("Encrypted Object too big")
	}

	return encrypted, nil
}

func (o *StoreTemplateObject) DecryptObject(key []byte, cbs []byte) error {
	// Validate Incoming Parameters
	if len(key) == 0 {
		return errors.New("Missing Decryption Key")
	}

	if len(cbs) == 0 {
		return errors.New("Missing Encrypted Bytes")
	}

	if len(cbs) > 65535 {
		return errors.New("Encrypted Object too big")
	}

	// Decrypted Bytes is JSON String
	decrypted, e := toPlainBytes(key, cbs)
	if e != nil {
		return e
	}

	return o.FromString(string(decrypted))
}

func (o *StoreTemplateObject) MarshalJSON() ([]byte, error) {
	// Template Information
	t := &struct {
		Name    string `json:"name"`
		Version uint16 `json:"version"`
	}{
		Name:    o.template,
		Version: o.version,
	}

	return json.Marshal(&struct {
		Template interface{} `json:"template"`
		Values   interface{} `json:"values"`
	}{
		Template: t,
		Values:   o.values,
	})
}

func (o *StoreTemplateObject) UnmarshalJSON(b []byte) error {
	// Reset Object
	o.reset()

	// Do we have a byte array (string) to convert to map?
	if len(b) == 0 { // NO
		return errors.New("Nothing to Unmarshal")
	}

	// Convert Byte Array to JSON
	var i interface{}
	e := json.Unmarshal(b, &i)
	if e != nil {
		return e
	}

	// Typed Alias
	m := i.(map[string]interface{})

	// Extract Template Information
	t := m["template"]
	if t == nil {
		return errors.New("JSON Missing 'template' structure")
	}

	e = o.extractTemplate(t.(map[string]interface{}))
	if e != nil {
		return e
	}

	// Extract Values
	v := m["values"]
	if v == nil {
		return errors.New("JSON Missing 'values' structure")
	}
	o.values = v.(map[string]interface{})

	// Extract Header Properties
	e = o.extractHeaderProps()
	if e != nil {
		return e
	}

	return nil
}

func (o *StoreTemplateObject) extractTemplate(t map[string]interface{}) error {
	// Extract Template Name
	n := t["name"]
	if n == nil {
		return errors.New("JSON Missing 'template.name'")
	}

	if name, ok := n.(string); ok {
		_, e := o.SetTemplate(name)
		if e != nil {
			return e
		}
	} else {
		return errors.New("JSON 'template.name' is not an non-empty string")
	}

	// Extract Template Version
	v := t["version"]
	if v == nil {
		return errors.New("JSON Missing 'template.version'")
	}

	if version, ok := v.(float64); ok {
		_, e := o.SetVersion(uint16(version))
		if e != nil {
			return e
		}
	} else {
		return errors.New("JSON 'template.version' is not valid")
	}

	return nil
}

func (o *StoreTemplateObject) extractHeaderProps() error {
	// Extract Template Name
	t := o.values["__title"]
	if t == nil {
		return errors.New("JSON Missing '__title'")
	}

	if title, ok := t.(string); ok {
		_, e := o.SetTitle(title)
		if e != nil {
			return e
		}
	} else {
		return errors.New("JSON '__title' is not an non-empty string")
	}

	return nil
}

func (o *StoreTemplateObject) reset() {
	// Clean Entry
	o.template = ""
	o.version = 0
	o.title = ""
	o.values = nil
}
