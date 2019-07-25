package strato

import (
	"context"
	"strato/proto"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	goodKey = "exists"
	badKey  = "does-not-exist"
)

var (
	ctx = context.Background()

	goodLoc = &proto.Location{
		Key: goodKey,
	}

	badLoc = &proto.Location{
		Key: badKey,
	}

	goodContent = []byte("here is some test value content")

	goodVal = &proto.Value{
		Content: goodContent,
	}

	goodReq = &proto.PutRequest{
		Location: goodLoc,
		Value:    goodVal,
	}
)

func TestServer(t *testing.T) {
	is := assert.New(t)

	srv, err := NewServer(goodServerCfg)
	is.NoError(err)
	is.NotNil(srv)

	t.Run("KV", func(t *testing.T) {
		empty, err := srv.Put(ctx, goodReq)
		is.NoError(err)
		is.NotNil(empty)

		fetched, err := srv.Get(ctx, goodLoc)
		is.NoError(err)
		is.NotNil(fetched)
		is.Equal(fetched.Value.Content, goodVal.Content)

		empty, err = srv.Delete(ctx, goodLoc)
		is.NoError(err)
		is.NotNil(empty)

		fetched, err = srv.Get(ctx, badLoc)
		stat, ok := status.FromError(err)
		is.True(ok)
		is.Equal(stat.Code(), codes.NotFound)
		is.Equal(stat.Message(), NotFound(&Location{Key: badKey}).Error())
		is.Nil(fetched)
	})

	t.Run("Shutdown", func(t *testing.T) {
		srv.ShutDown()
	})
}
