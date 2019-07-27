package strato

import (
	"testing"

	"github.com/lucperkins/strato/proto"

	"github.com/stretchr/testify/assert"
)

func TestKVTypes(t *testing.T) {
	is := assert.New(t)

	t.Run("Location", func(t *testing.T) {
		key := "some-key"

		loc := &Location{
			Key: key,
		}

		is.Equal(loc.String(), "Location<key: some-key>")
		is.Equal(loc.Proto(), &proto.Location{Key: key})
		is.Equal(loc.Key, key)
	})

	t.Run("Value", func(t *testing.T) {
		content := []byte("some test content")

		val := &Value{
			Content: content,
		}

		is.Equal(val.String(), `Value<content: "some test content">`)
		is.Equal(val.Proto(), &proto.Value{Content: content})
		is.Equal(val.Content, content)
	})
}
