package types

type Ray struct {
	Orig, Dir *Vec3
}

func NewRay(orig, dir *Vec3) *Ray {
	return &Ray{
		Orig: orig,
		Dir:  dir,
	}
}

func (r *Ray) At(t float64) *Vec3 {
	return Vec3Add(r.Orig, r.Dir.Scale(t))
}
