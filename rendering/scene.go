package rendering

import "github.com/Rilord/go-rt/types"

type Scene struct {
	Pixel00Loc   *types.Vec3
	PixelDeltaUV *types.UV
	Camera       *types.Camera
}

func NewScene(pixel00Loc *types.Vec3, pixelDeltaUV *types.UV, camera *types.Camera) *Scene {
	return &Scene{
		Pixel00Loc:   pixel00Loc,
		PixelDeltaUV: pixelDeltaUV,
		Camera:       camera,
	}
}
