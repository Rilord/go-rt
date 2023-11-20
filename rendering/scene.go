package rendering

import (
	"io"

	"github.com/Rilord/go-rt/accelerators"
	"github.com/Rilord/go-rt/types"
)

type Scene struct {
	TopLevel *accelerators.BVH
	Meshes   []*types.Mesh
	Cameras  []*types.Camera
}

func LoadScene(r *Renderer, io *io.ByteReader) *Scene {
}
