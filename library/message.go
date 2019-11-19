package library

import (
	. "github.com/shimmeringbee/unpi"
	"reflect"
)

type Library struct {
	identityToType map[Identity]reflect.Type
	typeToIdentity map[reflect.Type]Identity
}

type Identity struct {
	MessageType MessageType
	Subsystem   Subsystem
	CommandID   uint8
}

func New() *Library {
	return &Library{
		identityToType: make(map[Identity]reflect.Type),
		typeToIdentity: make(map[reflect.Type]Identity),
	}
}

func (cl *Library) Add(messageType MessageType, subsystem Subsystem, commandID uint8, v interface{}) {
	t := reflect.TypeOf(v)

	identity := Identity{
		MessageType: messageType,
		Subsystem:   subsystem,
		CommandID:   commandID,
	}

	cl.identityToType[identity] = t
	cl.typeToIdentity[t] = identity
}

func (cl *Library) GetByIdentifier(messageType MessageType, subsystem Subsystem, commandID uint8) (reflect.Type, bool) {
	identity := Identity{
		MessageType: messageType,
		Subsystem:   subsystem,
		CommandID:   commandID,
	}

	t, found := cl.identityToType[identity]
	return t, found
}

func (cl *Library) GetByObject(v interface{}) (Identity, bool) {
	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	identity, found := cl.typeToIdentity[t]
	return identity, found
}
