package queue

// cSpell:ignore amqp, otype

import (
	"github.com/gin-gonic/gin"

	"github.com/objectvault/api-services/orm/action"
	rpf "github.com/objectvault/goginrpf"

	"github.com/objectvault/queue-interface/messages"
)

func CreateActionMessage(r rpf.GINProcessor, c *gin.Context) {
	// Get the Required Invitation
	oa := r.MustGet("action").(*action.Action)

	// Create Email Message
	msg, err := messages.NewQueueActionWithGUID(oa.GUID(), oa.Type())
	if err != nil { // Failed: Abort
		r.Abort(5920, nil)
		return
	}

	// Initialize Action Message
	// msg.SetParameters(oa.Parameters().Map())
	// msg.SetProperties(oa.Properties().Map())

	// Save Activation
	r.Set("queue-message", msg)
}
