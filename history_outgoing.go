package main

import (
	"crypto/rand"
	"encoding/json"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (s *server) saveOutgoingRow(userID, chatJID, messageID, messageType, textContent, mediaLink, quotedMessageID string, sentAt time.Time, historyLimit int, message *waE2E.Message) {
	if historyLimit <= 0 {
		return
	}
	event := &events.Message{
		Info: types.MessageInfo{
			MessageSource: types.MessageSource{
				Chat:     types.JID{},
				IsFromMe: true,
			},
			ID:        types.MessageID(messageID),
			Timestamp: sentAt,
		},
		Message: message,
	}
	// Preserve the chat JID in the serialized event, matching received event data.
	if chat, err := types.ParseJID(chatJID); err == nil {
		event.Info.Chat = chat
	}
	data, err := json.Marshal(event)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal outgoing history action")
		return
	}
	if err = s.saveMessageToHistory(userID, chatJID, "me", messageID, messageType, textContent, mediaLink, quotedMessageID, string(data), sentAt); err != nil {
		log.Error().Err(err).Msg("Failed to save outgoing history action")
	}
}

// outgoingHistoryAction is the history row for a reaction, edit or revoke sent
// through the API. Rows use the same columns as received ones: reactions and
// edits keep their target in quoted_message_id, revokes in text_content.
type outgoingHistoryAction struct {
	messageType string
	textContent string
	target      string
	message     *waE2E.Message
}

// outgoingAction reads a built message the way the live event handler reads a
// received one. The stored message is unwrapped, because received events are:
// consumers read datajson's Message.protocolMessage, which BuildEdit nests
// inside editedMessage.
func outgoingAction(message *waE2E.Message) (outgoingHistoryAction, bool) {
	unwrapped := (&events.Message{RawMessage: message}).UnwrapRaw().Message
	if edit, ok := historyEdit(unwrapped); ok {
		return outgoingHistoryAction{"edit", edit.text, edit.target, unwrapped}, true
	}
	if reaction := unwrapped.GetReactionMessage(); reaction != nil {
		return outgoingHistoryAction{"reaction", reaction.GetText(), reaction.GetKey().GetID(), unwrapped}, true
	}
	if protocol := unwrapped.GetProtocolMessage(); protocol != nil && protocol.GetType() == waE2E.ProtocolMessage_REVOKE {
		return outgoingHistoryAction{"delete", protocol.GetKey().GetID(), "", unwrapped}, true
	}
	return outgoingHistoryAction{}, false
}

// saveOutgoingAction stores a reaction, edit or revoke after a successful send.
// Normally resp.ID identifies the action itself. Newsletter edits and revokes
// reuse the target's ID on the wire, so they need a separate local history ID.
func (s *server) saveOutgoingAction(userinfo Values, chat types.JID, resp whatsmeow.SendResponse, message *waE2E.Message) {
	action, ok := outgoingAction(message)
	if !ok {
		log.Warn().Str("messageID", resp.ID).Msg("Outgoing message is not a reaction, edit or revoke, skipping history")
		return
	}
	historyLimit, _ := strconv.Atoi(userinfo.Get("History"))
	messageID := resp.ID
	if chat.Server == types.NewsletterServer && (action.messageType == "edit" || action.messageType == "delete") {
		messageID = "newsletter-action-" + rand.Text()
	}
	s.saveOutgoingRow(userinfo.Get("Id"), chat.String(), messageID, action.messageType, action.textContent, "", action.target, resp.Timestamp, historyLimit, action.message)
}
