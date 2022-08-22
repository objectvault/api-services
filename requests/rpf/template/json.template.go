// cSpell:ignore goginrpf, gonic, paulo ferreira
package template

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

	"github.com/objectvault/api-services/maps"
	"github.com/objectvault/api-services/orm"
)

type FullTemplateToJSON struct {
	Template *orm.Template
}

func (o *FullTemplateToJSON) MarshalJSON() ([]byte, error) {
	if !o.Template.IsValid() {
		return nil, errors.New("Missing Required Structure Value [Template]")
	}

	// Extract Template Model in Object Format
	model, e := o.Template.ModelJSON()
	if e != nil {
		return nil, e
	}

	wrappedModel := maps.NewMapWrapper(model)

	// Add Standard Definition for Title Field
	wrappedModel.Set("fields.__title.label", "Title", false)
	wrappedModel.Set("fields.__title.type", "string", false)
	wrappedModel.Set("fields.__title.settings.max-length", 40, false)
	wrappedModel.Set("fields.__title.settings.required", true, false)
	wrappedModel.Set("fields.__title.checks.allow-empty", false, false)
	wrappedModel.Set("fields.__title.transforms.trim", true, false)
	wrappedModel.Set("fields.__title.transforms.single-space-between", true, false)
	wrappedModel.Set("fields.__title.transforms.null-on-empty", true, false)

	return json.Marshal(&struct {
		Name        string `json:"name"`
		Version     uint16 `json:"version"`
		Description string `json:"description"`
		//		Model       string `json:"model"`
		Model map[string]interface{} `json:"model"`
	}{
		Name:        o.Template.Template(),
		Version:     o.Template.Version(),
		Description: o.Template.Description(),
		//		Model:       o.Template.Model(),
		Model: model,
	})
}

type MinimalTemplateRegistryToJSON struct {
	Registry *orm.ObjectTemplateRegistry
}

func (o *MinimalTemplateRegistryToJSON) MarshalJSON() ([]byte, error) {
	if !o.Registry.IsValid() {
		return nil, errors.New("Missing Required Structore Value [Template]")
	}

	return json.Marshal(&struct {
		Name  string `json:"name"`
		Title string `json:"title"`
	}{
		Name:  o.Registry.Template(),
		Title: o.Registry.Title(),
	})
}

type FullTemplateRegistryToJSON struct {
	Registry *orm.ObjectTemplateRegistry
}

func (o *FullTemplateRegistryToJSON) MarshalJSON() ([]byte, error) {
	if !o.Registry.IsValid() {
		return nil, errors.New("Missing Required Structore Value [Template]")
	}

	return json.Marshal(&struct {
		Object string `json:"object"`
		Name   string `json:"name"`
		Title  string `json:"title"`
	}{
		Object: fmt.Sprintf(":%x", o.Registry.Object()),
		Name:   o.Registry.Template(),
		Title:  o.Registry.Title(),
	})
}
