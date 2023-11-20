package types

type Camera struct {
	FOV           float64
	FocalLength   float64
	FocalDistance float64
	Position      *Vec3
	AspectRation  float64

	Up, Right *Vec3
	LookAt    *Vec3
	Forward   *Vec3

	Height float64
	Width  float64
}
