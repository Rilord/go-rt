package rendering

import (
	"context"
	"errors"
	"image"

	"github.com/Rilord/go-rt/types"
)

type RenderTile struct {
	X, Y int
}

type RenderState struct {
	*RenderTile
}

func NewRenderState(tile *RenderTile) *RenderState {
	return &RenderState{
		RenderTile: tile,
	}

}

func NewRenderer(scene *Scene, state *RenderState) *Renderer {
	return &Renderer{
		Scene: scene,
		State: state,
	}
}

type Renderer struct {
	*Scene
	State *RenderState
}

func RenderPixelFn(parentCtx context.Context, ctx context.Context, data ...interface{}) (error, []interface{}) {

	var renderer *Renderer
	var img *image.RGBA

	if len(data) != 2 {
		return errors.New("Incorrect number of args"), nil
	}

	renderer, ok := data[0].(*Renderer)

	if !ok {
		return errors.New("Couldn't cast renderer type"), nil
	}

	img, ok = data[1].(*image.RGBA)

	if !ok {
		return errors.New("Couldn't cast image type"), nil
	}

	i, j := renderer.State.RenderTile.X, renderer.State.RenderTile.Y

	pixelCenter := renderer.Pixel00Loc.
		Add(
			renderer.PixelDeltaUV.U.Scale(float64(i)),
		).
		Add(
			renderer.PixelDeltaUV.V.Scale(float64(j)),
		)

	rayDir := pixelCenter.Sub(renderer.Camera.Center)

	r := types.NewRay(renderer.Camera.Center, rayDir)

	pixelColor := types.NewColor(0, 0, 0).FromRay(r)
	img.Set(i, j, pixelColor.ToStdColor())

	pixelColors := make([]interface{}, 0)
	pixelColors = append(pixelColors, pixelColor)

	return nil, pixelColors
}
