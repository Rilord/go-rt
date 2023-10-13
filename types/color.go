package types

import (
	"image/color"
)

type Color Vec3

func NewColor(r, g, b float64) *Color {
	return &Color{
		X: r,
		Y: g,
		Z: b,
	}
}

func (c *Color) FromRay(r *Ray) *Color {
	unitDir := r.Dir.Unit()
	a := (unitDir.Y + 1.0) * 0.5
	return (*Color)(
		NewVec3(1.0, 1.0, 1.0).
			Scale(1.0 - a).
			Add(
				NewVec3(0.5, 0.7, 1.0).
					Scale(a),
			),
	)
}

func (c *Color) ToStdColor() color.RGBA {
	return color.RGBA{uint8(c.X * 255.0), uint8(c.Y * 255.0), uint8(c.Z * 255.0), 0xff}
}
