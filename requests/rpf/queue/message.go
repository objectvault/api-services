package queue

// cSpell:ignore amqp, otype

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"

	rpf "github.com/objectvault/goginrpf"

	"github.com/objectvault/api-services/common"
	"github.com/objectvault/api-services/orm"
	"github.com/objectvault/queue-interface/messages"
)

func CreateMessageDeleteUserFromSystem(r rpf.GINProcessor, c *gin.Context) {
	// Get the Required Invitation
	user := r.MustGet("request-user").(uint64)

	// Create Action Message
	msg := &messages.ActionMessage{}

	// Create GUID (V4 see https://www.sohamkamani.com/uuid-versions-explained/)
	uid, err := uuid.NewV4()
	if err != nil {
		r.Abort(5920, nil)
		return
	}

	// Initialize Action Message
	err = messages.InitQueueAction(msg, uid.String(), "system:user:delete")
	if err != nil { // Failed: Abort
		r.Abort(5920, nil)
		return
	}

	// Set User Coordinates
	msg.SetParameter("user", fmt.Sprintf(":%x", user))

	//Set Action Creator's Information
	actionUser := r.MustGet("action-user").(uint64)
	msg.SetParameter("action-user", fmt.Sprintf(":%x", actionUser))
	msg.SetParameter("action-user-name", r.MustGet("action-user-name"))
	msg.SetParameter("action-user-email", r.MustGet("action-user-email"))

	// Save Activation
	r.Set("queue-message", msg)
}

func CreateMessageDeleteOrgFromSystem(r rpf.GINProcessor, c *gin.Context) {
	// Get the Required Invitation
	org := r.MustGet("request-org").(uint64)

	// Create Action Message
	msg := &messages.ActionMessage{}

	// Create GUID (V4 see https://www.sohamkamani.com/uuid-versions-explained/)
	uid, err := uuid.NewV4()
	if err != nil {
		r.Abort(5920, nil)
		return
	}

	// Initialize Action Message
	err = messages.InitQueueAction(msg, uid.String(), "system:org:delete")
	if err != nil { // Failed: Abort
		r.Abort(5920, nil)
		return
	}

	// Set Organization Coordinates
	msg.SetParameter("organization", fmt.Sprintf(":%x", org))

	//Set Action Creator's Information
	actionUser := r.MustGet("action-user").(uint64)
	msg.SetParameter("action-user", fmt.Sprintf(":%x", actionUser))
	msg.SetParameter("action-user-name", r.MustGet("action-user-name"))
	msg.SetParameter("action-user-email", r.MustGet("action-user-email"))

	// Save Activation
	r.Set("queue-message", msg)
}

func CreateMessageDeleteUserFromOrg(r rpf.GINProcessor, c *gin.Context) {
	// Get the Required Invitation
	org := r.MustGet("request-org").(uint64)
	user := r.MustGet("request-user").(uint64)

	// Create Action Message
	msg := &messages.ActionMessage{}

	// Create GUID (V4 see https://www.sohamkamani.com/uuid-versions-explained/)
	uid, err := uuid.NewV4()
	if err != nil {
		r.Abort(5920, nil)
		return
	}

	// Initialize Action Message
	err = messages.InitQueueAction(msg, uid.String(), "org:user:delete")
	if err != nil { // Failed: Abort
		r.Abort(5920, nil)
		return
	}

	// Set User Coordinates
	msg.SetParameter("organization", fmt.Sprintf(":%x", org))
	msg.SetParameter("user", fmt.Sprintf(":%x", user))

	//Set Action Creator's Information
	actionUser := r.MustGet("action-user").(uint64)
	msg.SetParameter("action-user", fmt.Sprintf(":%x", actionUser))
	msg.SetParameter("action-user-name", r.MustGet("action-user-name"))
	msg.SetParameter("action-user-email", r.MustGet("action-user-email"))

	// Save Activation
	r.Set("queue-message", msg)
}

func CreateMessageDeleteStore(r rpf.GINProcessor, c *gin.Context) {
	// Get the Required Invitation
	org := r.MustGet("request-org").(uint64)
	store := r.MustGet("request-store").(uint64)

	// Create Action Message
	msg := &messages.ActionMessage{}

	// Create GUID (V4 see https://www.sohamkamani.com/uuid-versions-explained/)
	uid, err := uuid.NewV4()
	if err != nil {
		r.Abort(5920, nil)
		return
	}

	// Initialize Action Message
	err = messages.InitQueueAction(msg, uid.String(), "org:store:delete")
	if err != nil { // Failed: Abort
		r.Abort(5920, nil)
		return
	}

	// Set Store Coordinates
	msg.SetParameter("organization", fmt.Sprintf(":%x", org))
	msg.SetParameter("store", fmt.Sprintf(":%x", store))

	//Set Action Creator's Information
	actionUser := r.MustGet("action-user").(uint64)
	msg.SetParameter("action-user", fmt.Sprintf(":%x", actionUser))
	msg.SetParameter("action-user-name", r.MustGet("action-user-name"))
	msg.SetParameter("action-user-email", r.MustGet("action-user-email"))

	// Save Activation
	r.Set("queue-message", msg)
}

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
