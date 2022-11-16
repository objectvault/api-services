package queue

// cSpell:ignore amqp, otype

import (
	"github.com/gin-gonic/gin"

	rpf "github.com/objectvault/goginrpf"

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/queue-interface/messages"
)

func CreateInvitationMessage(r rpf.GINProcessor, c *gin.Context) {
	// Get the Required Invitation
	i := r.MustGet("invitation").(*orm.Invitation)
	o := r.MustGet("registry-org").(*orm.OrgRegistry)

	// DEFAULT: Organization Invitation
	ot := "organization"

	// Is invitation for a store?
	if common.IsObjectOfType(i.Object(), common.OTYPE_STORE) { // YES
		ot = "store"
	}

	// Create Email Message
	msg, err := messages.NewInviteMessage(ot, i.UID())
	if err != nil { // Failed: Abort
		r.Abort(5920, nil)
		return
	}

	// Set Message Parameters
	msg.SetTo(i.InviteeEmail())
	msg.SetByEmail(r.MustGet("from-user-email").(string))
	msg.SetByUser(r.MustGet("from-user-name").(string))
	msg.SetMessage(i.Message())
	msg.SetObjectName(o.Name())
	msg.SetExpiration(*i.Expiration())

	// Is invitation for a store?
	if common.IsObjectOfType(i.Object(), common.OTYPE_STORE) { // YES
		s := r.MustGet("registry-store").(*orm.OrgStoreRegistry)
		msg.SetStoreName(s.StoreAlias())
	}

	// Save Activation
	r.Set("queue-message", msg)
}
