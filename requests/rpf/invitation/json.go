// cSpell:ignore goginrpf, gonic, paulo ferreira
package invitation

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

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
)

// Invitation Registry Export
type RegistryInviteToJSON struct {
	Entry *orm.InvitationRegistry
}

func (ore *RegistryInviteToJSON) MarshalJSON() ([]byte, error) {
	if !ore.Entry.IsValid() {
		return nil, errors.New("Missing Required Structore Value [Entry]")
	}

	return json.Marshal(&struct {
		ID         string `json:"id"`
		UID        string `json:"uid"`
		Creator    uint64 `json:"creator"`
		Invitee    string `json:"invitee"`
		Expiration string `json:"expiration"`
		State      uint16 `json:"state"`
	}{
		ID:         fmt.Sprintf(":%x", ore.Entry.ID()),
		UID:        ore.Entry.UID(),
		Creator:    ore.Entry.Creator(),
		Invitee:    ore.Entry.InviteeEmail(),
		Expiration: ore.Entry.ExpirationUTC(),
		State:      ore.Entry.State(),
	})
}

type NoSessionInviteToJSON struct {
	Invite  *orm.InvitationRegistry
	Creator *orm.UserRegistry
	Invitee *orm.UserRegistry
}

func (ore *NoSessionInviteToJSON) MarshalJSON() ([]byte, error) {
	if !ore.Invite.IsValid() {
		return nil, errors.New("Missing Required Structore Value [Invite]")
	}

	if !ore.Creator.IsValid() {
		return nil, errors.New("Missing Required Structore Value [Creator]")
	}

	response := &struct {
		UID               string `json:"id"`
		Creator           string `json:"creator"`
		Invitee           string `json:"invitee"`
		RegisteredInvitee bool   `json:"registered_invitee"`
		RegisterSession   bool   `json:"register_session"`
	}{
		UID:               ore.Invite.UID(),
		Creator:           ore.Creator.Email(),
		Invitee:           ore.Invite.InviteeEmail(),
		RegisteredInvitee: false,
		RegisterSession:   common.IsObjectOfType(ore.Invite.Object(), common.OTYPE_STORE),
	}

	// Does Invitee Exist?
	if ore.Invitee != nil { // YES: Flag user as registered
		response.RegisteredInvitee = true
	}

	return json.Marshal(response)
}
