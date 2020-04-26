// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package crypto

import (
	"encoding/json"

	"github.com/pkg/errors"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var (
	UnsupportedAlgorithm                = errors.New("unsupported event encryption algorithm")
	NotEncryptedForMe                   = errors.New("olm event doesn't contain ciphertext for this device")
	UnsupportedOlmMessageType           = errors.New("unsupported olm message type")
	DecryptionFailedWithMatchingSession = errors.New("decryption failed with matching session")
	DecryptionFailedForNormalMessage    = errors.New("decryption failed for normal message")
	SenderMismatch                      = errors.New("mismatched sender in olm payload")
	RecipientMismatch                   = errors.New("mismatched recipient in olm payload")
	RecipientKeyMismatch                = errors.New("mismatched recipient key in olm payload")
)

type OlmEvent struct {
	Source *event.Event `json:"-"`

	SenderKey id.SenderKey `json:"-"`

	Sender        id.UserID    `json:"sender"`
	SenderDevice  id.DeviceID  `json:"sender_device"`
	Keys          OlmEventKeys `json:"keys"`
	Recipient     id.UserID    `json:"recipient"`
	RecipientKeys OlmEventKeys `json:"recipient_keys"`

	Type    event.Type    `json:"type"`
	Content event.Content `json:"content"`
}

func (mach *OlmMachine) DecryptOlmEvent(evt *event.Event) (*OlmEvent, error) {
	content, ok := evt.Content.Parsed.(*event.EncryptedEventContent)
	if !ok {
		return nil, IncorrectEncryptedContentType
	} else if content.Algorithm != id.AlgorithmOlmV1 {
		return nil, UnsupportedAlgorithm
	}
	_, ownKey := mach.account.IdentityKeys()
	ownContent, ok := content.OlmCiphertext[string(ownKey)]
	if !ok {
		return nil, NotEncryptedForMe
	}
	return mach.decryptOlmEvent(evt, content.SenderKey, ownContent.Type, ownContent.Body)
}

type OlmEventKeys struct {
	Ed25519 id.Ed25519 `json:"ed25519"`
}

func (mach *OlmMachine) decryptOlmEvent(evt *event.Event, senderKey id.SenderKey, olmType id.OlmMsgType, ciphertext string) (*OlmEvent, error) {
	if olmType != id.OlmMsgTypePreKey && olmType != id.OlmMsgTypeMsg {
		return nil, UnsupportedOlmMessageType
	}

	plaintext, err := mach.tryDecryptOlmEvent(senderKey, olmType, ciphertext)
	if err != nil {
		if err == DecryptionFailedWithMatchingSession {
			mach.log.Warn("Found matching session yet decryption failed for sender %s with key %s", evt.Sender, senderKey)
			mach.markDeviceForUnwedging(evt.Sender, senderKey)
		}
		return nil, errors.Wrap(err, "failed to decrypt olm event")
	}

	// Decryption failed with every known session or no known sessions, let's try to create a new session.
	if plaintext == nil {
		// New sessions can only be created if it's a prekey message, we can't decrypt the message
		// if it isn't one at this point in time anymore, so return early.
		if olmType != id.OlmMsgTypePreKey {
			mach.markDeviceForUnwedging(evt.Sender, senderKey)
			return nil, DecryptionFailedForNormalMessage
		}

		session, err := mach.createInboundSession(senderKey, ciphertext)
		if err != nil {
			mach.markDeviceForUnwedging(evt.Sender, senderKey)
			return nil, errors.Wrap(err, "failed to create new session from prekey message")
		}

		plaintext, err = session.Decrypt(ciphertext, olmType)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decrypt olm event with session created from prekey message")
		}
	}

	var olmEvt OlmEvent
	err = json.Unmarshal(plaintext, &olmEvt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse olm payload")
	}
	if evt.Sender != olmEvt.Sender {
		return nil, SenderMismatch
	} else if mach.client.UserID != olmEvt.Recipient {
		return nil, RecipientMismatch
	} else if ed25519, _ := mach.account.IdentityKeys(); ed25519 != olmEvt.RecipientKeys.Ed25519 {
		return nil, RecipientKeyMismatch
	}

	err = olmEvt.Content.ParseRaw(olmEvt.Type)
	if err != nil && !event.IsUnsupportedContentType(err) {
		return nil, errors.Wrap(err, "failed to parse content of olm payload event")
	}

	olmEvt.Source = evt
	olmEvt.SenderKey = senderKey

	return &olmEvt, nil
}

func (mach *OlmMachine) tryDecryptOlmEvent(senderKey id.SenderKey, olmType id.OlmMsgType, ciphertext string) ([]byte, error) {
	sessions, err := mach.store.GetSessions(senderKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get session for %s", senderKey)
	}
	for _, session := range sessions {
		if olmType == id.OlmMsgTypePreKey {
			matches, err := session.MatchesInboundSession(ciphertext)
			if err != nil {
				return nil, errors.Wrap(err, "failed to check if ciphertext matches inbound session")
			} else if !matches {
				continue
			}
		}
		plaintext, err := session.Decrypt(ciphertext, olmType)
		if err != nil {
			if olmType == id.OlmMsgTypePreKey {
				return nil, DecryptionFailedWithMatchingSession
			}
		} else {
			return plaintext, nil
		}
	}
	return nil, nil
}

func (mach *OlmMachine) createInboundSession(senderKey id.SenderKey, ciphertext string) (*OlmSession, error) {
	session, err := mach.account.NewInboundSessionFrom(senderKey, ciphertext)
	if err != nil {
		return nil, err
	}
	mach.SaveAccount()
	err = mach.store.AddSession(senderKey, session)
	if err != nil {
		mach.log.Error("Failed to store created inbound session: %v", err)
	}
	return session, nil
}
