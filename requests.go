package mautrix

import (
	"encoding/json"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// ReqRegister is the JSON request for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-register
type ReqRegister struct {
	Username                 string      `json:"username,omitempty"`
	BindEmail                bool        `json:"bind_email,omitempty"`
	Password                 string      `json:"password,omitempty"`
	DeviceID                 id.DeviceID `json:"device_id,omitempty"`
	InitialDeviceDisplayName string      `json:"initial_device_display_name"`
	Auth                     interface{} `json:"auth,omitempty"`
}

type UserIdentifier struct {
	Type string `json:"type"`

	User string `json:"user,omitempty"`

	Medium  string `json:"medium,omitempty"`
	Address string `json:"address,omitempty"`

	Country string `json:"country,omitempty"`
	Phone   string `json:"phone,omitempty"`
}

// ReqLogin is the JSON request for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-login
type ReqLogin struct {
	Type                     string         `json:"type"`
	Identifier               UserIdentifier `json:"identifier"`
	Password                 string         `json:"password,omitempty"`
	Token                    string         `json:"token,omitempty"`
	DeviceID                 id.DeviceID    `json:"device_id,omitempty"`
	InitialDeviceDisplayName string         `json:"initial_device_display_name,omitempty"`
}

// ReqCreateRoom is the JSON request for https://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-createroom
type ReqCreateRoom struct {
	Visibility      string                 `json:"visibility,omitempty"`
	RoomAliasName   string                 `json:"room_alias_name,omitempty"`
	Name            string                 `json:"name,omitempty"`
	Topic           string                 `json:"topic,omitempty"`
	Invite          []id.UserID            `json:"invite,omitempty"`
	Invite3PID      []ReqInvite3PID        `json:"invite_3pid,omitempty"`
	CreationContent map[string]interface{} `json:"creation_content,omitempty"`
	InitialState    []*event.Event         `json:"initial_state,omitempty"`
	Preset          string                 `json:"preset,omitempty"`
	IsDirect        bool                   `json:"is_direct,omitempty"`
}

// ReqRedact is the JSON request for http://matrix.org/docs/spec/client_server/r0.2.0.html#put-matrix-client-r0-rooms-roomid-redact-eventid-txnid
type ReqRedact struct {
	Reason string `json:"reason,omitempty"`
	TxnID  string `json:"-"`
}

type ReqMembers struct {
	At            string           `json:"at"`
	Membership    event.Membership `json:"membership,omitempty"`
	NotMembership event.Membership `json:"not_membership,omitempty"`
}

// ReqInvite3PID is the JSON request for https://matrix.org/docs/spec/client_server/r0.2.0.html#id57
// It is also a JSON object used in https://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-createroom
type ReqInvite3PID struct {
	IDServer string `json:"id_server"`
	Medium   string `json:"medium"`
	Address  string `json:"address"`
}

// ReqInviteUser is the JSON request for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-invite
type ReqInviteUser struct {
	UserID id.UserID `json:"user_id"`
}

// ReqKickUser is the JSON request for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-kick
type ReqKickUser struct {
	Reason string    `json:"reason,omitempty"`
	UserID id.UserID `json:"user_id"`
}

// ReqBanUser is the JSON request for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-ban
type ReqBanUser struct {
	Reason string    `json:"reason,omitempty"`
	UserID id.UserID `json:"user_id"`
}

// ReqUnbanUser is the JSON request for http://matrix.org/docs/spec/client_server/r0.2.0.html#post-matrix-client-r0-rooms-roomid-unban
type ReqUnbanUser struct {
	UserID id.UserID `json:"user_id"`
}

// ReqTyping is the JSON request for https://matrix.org/docs/spec/client_server/r0.2.0.html#put-matrix-client-r0-rooms-roomid-typing-userid
type ReqTyping struct {
	Typing  bool  `json:"typing"`
	Timeout int64 `json:"timeout,omitempty"`
}

type ReqPresence struct {
	Presence string `json:"presence"`
}

type ReqAliasCreate struct {
	RoomID id.RoomID `json:"room_id"`
}

type OneTimeKey struct {
	Key        string                 `json:"key"`
	IsSigned   bool                   `json:"-"`
	Signatures Signatures             `json:"signatures,omitempty"`
	Unsigned   map[string]interface{} `json:"unsigned,omitempty"`
}

type serializableOTK OneTimeKey

func (otk *OneTimeKey) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, (*serializableOTK)(otk))
	if err != nil {
		var key string
		err := json.Unmarshal(data, &key)
		if err != nil {
			return err
		}
		otk.Key = key
		otk.Signatures = nil
		otk.Unsigned = nil
		otk.IsSigned = false
	} else {
		otk.IsSigned = true
	}
	return nil
}

func (otk *OneTimeKey) MarshalJSON() ([]byte, error) {
	if !otk.IsSigned {
		return json.Marshal(otk.Key)
	} else {
		return json.Marshal((*serializableOTK)(otk))
	}
}

type ReqUploadKeys struct {
	DeviceKeys  *DeviceKeys             `json:"device_keys,omitempty"`
	OneTimeKeys map[id.KeyID]OneTimeKey `json:"one_time_keys"`
}

type DeviceKeys struct {
	UserID     id.UserID                 `json:"user_id"`
	DeviceID   id.DeviceID               `json:"device_id"`
	Algorithms []id.Algorithm            `json:"algorithms"`
	Keys       map[id.DeviceKeyID]string `json:"keys"`
	Signatures Signatures                `json:"signatures"`
	Unsigned   map[string]interface{}    `json:"unsigned,omitempty"`
}

type Signatures map[id.UserID]map[id.DeviceKeyID]string

type ReqQueryKeys struct {
	DeviceKeys map[id.UserID][]id.DeviceID `json:"device_keys"`

	Timeout int64  `json:"timeout,omitempty"`
	Token   string `json:"token,omitempty"`
}

type ReqClaimKeys struct {
	OneTimeKeys map[id.UserID]map[id.DeviceID]string `json:"one_time_keys"`

	Timeout int64 `json:"timeout,omitempty"`
}
