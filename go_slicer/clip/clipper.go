package clip

import (
	"GoSlicer/go_slicer/data"
	"GoSlicer/util"
	clipper "github.com/ctessum/go.clipper"
)

type Clip interface {
	// GenerateLayerParts partitions the whole layer into several partition parts
	GenerateLayerParts(l data.Layer) (data.PartitionedLayer, bool)
	InsetLayer(layer data.PartitionedLayer, offset util.Micrometer, insetCount int) [][]data.Paths
	Inset(part data.LayerPart, offset util.Micrometer, insetCount int) []data.Paths
}

// clipperClip implements Clip using the external clipper library
type clipperClip struct {
}

func NewClip() Clip {
	return clipperClip{}
}

type layerPart struct {
	outline data.Path
	holes   data.Paths
}

func (l layerPart) Outline() data.Path {
	return l.outline
}

func (l layerPart) Holes() data.Paths {
	return l.holes
}

type partitionedLayer struct {
	parts []data.LayerPart
}

func (p partitionedLayer) LayerParts() []data.LayerPart {
	return p.parts
}

func clipperPoint(p util.MicroPoint) *clipper.IntPoint {
	return &clipper.IntPoint{
		X: clipper.CInt(p.X()),
		Y: clipper.CInt(p.Y()),
	}
}

func clipperPaths(p data.Paths) clipper.Paths {
	var result clipper.Paths
	for _, path := range p {
		var newPath clipper.Path
		for _, point := range path {
			newPath = append(newPath, clipperPoint(point))
		}
		result = append(result, newPath)
	}

	return result
}

func microPoint(p *clipper.IntPoint) util.MicroPoint {
	return util.NewMicroPoint(util.Micrometer(p.X), util.Micrometer(p.Y))
}

func microPath(p clipper.Path) data.Path {
	var result data.Path
	for _, point := range p {
		result = append(result, microPoint(point))
	}
	return result
}

func microPaths(p clipper.Paths, simplify bool) data.Paths {
	var result data.Paths

	for _, path := range p {
		microPath := microPath(path)

		if simplify {
			microPath = microPath.Simplify(-1, -1)
		}

		result = append(result,
			microPath.Simplify(-1, -1),
		)
	}

	return result
}

func (c clipperClip) GenerateLayerParts(l data.Layer) (data.PartitionedLayer, bool) {
	polyList := clipper.Paths{}
	// convert all polygons to clipper polygons
	for _, layerPolygon := range l.Polygons() {
		var path = clipper.Path{}

		prev := 0
		// convert all points of this polygons
		for j, layerPoint := range layerPolygon {
			// ignore first as the next check would fail otherwise
			if j == 0 {
				path = append(path, clipperPoint(layerPolygon[0]))
				continue
			}

			// filter too near points
			// check this always with the previous point
			if layerPoint.Sub(layerPolygon[prev]).ShorterThan(100) {
				continue
			}

			path = append(path, clipperPoint(layerPoint))
			prev = j
		}

		polyList = append(polyList, path)
	}

	layer := partitionedLayer{}

	clip := clipper.NewClipper(clipper.IoNone)
	clip.AddPaths(polyList, clipper.PtSubject, true)
	resultPolys, ok := clip.Execute2(clipper.CtUnion, clipper.PftEvenOdd, clipper.PftEvenOdd)
	if !ok {
		return nil, false
	}

	polysForNextRound := []*clipper.PolyNode{}

	for _, c := range resultPolys.Childs() {
		polysForNextRound = append(polysForNextRound, c)
	}
	for {
		if polysForNextRound == nil {
			break
		}
		thisRound := polysForNextRound
		polysForNextRound = nil

		for _, p := range thisRound {

			part := layerPart{
				outline: microPath(p.Contour()),
			}
			for _, child := range p.Childs() {
				part.holes = append(part.holes, microPath(child.Contour()))
				for _, c := range child.Childs() {
					polysForNextRound = append(polysForNextRound, c)
				}
			}
			layer.parts = append(layer.parts, &part)
		}
	}
	return layer, true
}

func (c clipperClip) InsetLayer(layer data.PartitionedLayer, offset util.Micrometer, insetCount int) [][]data.Paths {
	var result [][]data.Paths
	for _, part := range layer.LayerParts() {
		result = append(result, c.Inset(part, offset, insetCount))
	}

	return result
}

func (c clipperClip) Inset(part data.LayerPart, offset util.Micrometer, insetCount int) []data.Paths {
	var insets []data.Paths

	// save which insets are already finished and don't process them further (for more performance)
	insetsFinished := false
	o := clipper.NewClipperOffset()

	for i := 0; i < insetCount; i++ {
		// insets for the outline
		if !insetsFinished {
			o.Clear()
			o.AddPaths(clipperPaths(data.Paths{part.Outline()}), clipper.JtSquare, clipper.EtClosedPolygon)
			o.AddPaths(clipperPaths(part.Holes()), clipper.JtSquare, clipper.EtClosedPolygon)

			o.MiterLimit = 2
			newInset := o.Execute(float64(-int(offset)*i) - float64(offset/2))
			if len(newInset) <= 0 {
				insetsFinished = true
			} else {
				insets = append(insets, microPaths(newInset, true))
			}
		}
	}

	return insets
}

func (c clipperClip) GetLinearFill(poly data.Path, lineWidth util.Micrometer) {
	/*paths := clipperPaths(data.Paths{poly})
	bounds := clipper.GetBounds(paths)
	cl := clipper.NewClipper(clipper.IoNone)

	cl.AddPath(paths[0], clipper.PtSubject, true)
	cl.

	currentPos := clipper.NewIntPoint(bounds, 0)
	for {

	}*/
}
