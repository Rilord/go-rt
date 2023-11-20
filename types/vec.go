package types

import "math"

type Vec3 struct {
	X, Y, Z float64
}

func NewVec3(x, y, z float64) *Vec3 {
	return &Vec3{
		X: x,
		Y: y,
		Z: z,
	}
}

func (v *Vec3) Add(a *Vec3) *Vec3 {
	return &Vec3{
		X: v.X + a.X,
		Y: v.Y + a.Y,
		Z: v.Z + a.Z,
	}
}

func (v *Vec3) Sub(a *Vec3) *Vec3 {
	return &Vec3{
		X: v.X - a.X,
		Y: v.Y - a.Y,
		Z: v.Z - a.Z,
	}
}
func (v *Vec3) AddScalar(c float64) *Vec3 {
	return &Vec3{
		X: v.X + c,
		Y: v.Y + c,
		Z: v.Z + c,
	}
}

func (v *Vec3) Scale(c float64) *Vec3 {
	return &Vec3{
		X: v.X * c,
		Y: v.Y * c,
		Z: v.Z * c,
	}
}

func Vec3Add(v1, v2 *Vec3) *Vec3 {
	return &Vec3{
		X: v1.X + v2.X,
		Y: v1.Y + v2.Y,
		Z: v1.Z + v2.Z,
	}
}

func Vec3Sub(v1, v2 *Vec3) *Vec3 {
	return &Vec3{
		X: v1.X - v2.X,
		Y: v1.Y - v2.Y,
		Z: v1.Z - v2.Z,
	}
}

func Vec3Dot(v1, v2 *Vec3) float64 {
	return v1.X*v2.X + v1.Y*v2.Y + v1.Z*v2.Z
}

func (v *Vec3) LenSquared() float64 {
	return Vec3Dot(v, v)
}

func (v *Vec3) Len() float64 {
	return math.Sqrt(v.LenSquared())
}

func (v *Vec3) Unit() *Vec3 {
	return v.Scale(1 / v.Len())
}

type Coord Vec3

type UV struct {
	U, V *Vec3
}
