package types

import "github.com/Rilord/go-rt/accelerators"

type Mesh struct {
	Vertices  []*Vec3
	Normals   []*Vec3
	TexCoords []*Vec3
	Polygons  []*Polygon
	BVH       *accelerators.BVH
	Name      string
	RayOffset float64
}
