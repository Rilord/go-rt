package rendering

import (
	"github.com/Rilord/go-rt/pkg/concurrent"
	"github.com/Rilord/go-rt/types"
)

type RenderWorker struct {
	concurrent.Worker

	*types.Camera
	*Renderer

	curTile *RenderTile
}
