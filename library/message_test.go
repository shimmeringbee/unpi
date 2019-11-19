package library

import (
	. "github.com/shimmeringbee/unpi"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestMessageLibrary(t *testing.T) {
	t.Run("verifies that the message library returns false if message not found", func(t *testing.T) {
		ml := NewLibrary()

		_, found := ml.GetByIdentifier(AREQ, SYS, 0xff)
		assert.False(t, found)

		type UnknownStruct struct{}

		_, found = ml.GetByObject(UnknownStruct{})
		assert.False(t, found)
	})

	t.Run("verifies that registered messages are available", func(t *testing.T) {
		ml := NewLibrary()

		type KnownStruct struct{}

		commandId := byte(0x80)
		ml.Add(AREQ, SYS, commandId, KnownStruct{})

		expectedType := reflect.TypeOf(KnownStruct{})
		actualType, found := ml.GetByIdentifier(AREQ, SYS, commandId)

		assert.True(t, found)
		assert.Equal(t, expectedType, actualType)

		expectedIdentity := Identity{MessageType: AREQ, Subsystem: SYS, CommandID: commandId}
		actualIdentity, found := ml.GetByObject(KnownStruct{})

		assert.True(t, found)
		assert.Equal(t, expectedIdentity, actualIdentity)
	})
}
