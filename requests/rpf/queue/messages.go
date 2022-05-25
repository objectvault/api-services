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

	// Create Email Message
	msg := &messages.InvitationEmailMessage{}

	// DEFAULT: Organization Invitation
	template := "org-invitation"

	// Is invitation for a store?
	if common.IsObjectOfType(i.Object(), common.OTYPE_STORE) { // YES
		template = "store-invitation"
	}

	// Set Message Parameters
	msg.SetTemplate(template)
	msg.SetTo(i.InviteeEmail())
	msg.SetAtUser(r.MustGet("from-user-email").(string))
	msg.SetByUser(r.MustGet("from-user-name").(string))
	msg.SetCode(i.UID())
	msg.SetMessage(i.Message())
	msg.SetObjectName(o.Name())

	// Save Activation
	r.Set("queue-message", msg)
}
