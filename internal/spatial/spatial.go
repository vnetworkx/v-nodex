package spatial

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
)

func ClampVector(v [3]float64, min, max float64) [3]float64 {
	out := v
	for i := 0; i < 3; i++ {
		if out[i] < min {
			out[i] = min
		}
		if out[i] > max {
			out[i] = max
		}
	}
	return out
}

func Distance(a, b [3]float64) float64 {
	dx := a[0] - b[0]
	dy := a[1] - b[1]
	dz := a[2] - b[2]
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func CellKey(v [3]float64, cellSize float64) string {
	if cellSize <= 0 {
		cellSize = 1
	}
	x := int(math.Floor(v[0] / cellSize))
	y := int(math.Floor(v[1] / cellSize))
	z := int(math.Floor(v[2] / cellSize))
	return fmt.Sprintf("%d:%d:%d", x, y, z)
}

func OctreePath(v [3]float64, depth int, worldScale float64) string {
	if depth <= 0 {
		return ""
	}
	if worldScale <= 0 {
		worldScale = 1
	}

	min := [3]float64{-worldScale, -worldScale, -worldScale}
	max := [3]float64{worldScale, worldScale, worldScale}

	var b strings.Builder
	for i := 0; i < depth; i++ {
		mid := [3]float64{
			(min[0] + max[0]) / 2,
			(min[1] + max[1]) / 2,
			(min[2] + max[2]) / 2,
		}

		octant := byte(0)
		if v[0] >= mid[0] {
			octant |= 1
			min[0] = mid[0]
		} else {
			max[0] = mid[0]
		}

		if v[1] >= mid[1] {
			octant |= 2
			min[1] = mid[1]
		} else {
			max[1] = mid[1]
		}

		if v[2] >= mid[2] {
			octant |= 4
			min[2] = mid[2]
		} else {
			max[2] = mid[2]
		}

		b.WriteByte('0' + octant)
	}

	return b.String()
}

func RegionID(v [3]float64, depth int, worldScale float64) string {
	return fmt.Sprintf("r/%d/%s", depth, OctreePath(v, depth, worldScale))
}

func RouteKey(entityID, regionID string) string {
	sum := sha256.Sum256([]byte(entityID + "|" + regionID))
	return hex.EncodeToString(sum[:])
}

func Quantize(v [3]float64, step float64) [3]int64 {
	if step <= 0 {
		step = 1
	}
	return [3]int64{
		int64(math.Round(v[0] / step)),
		int64(math.Round(v[1] / step)),
		int64(math.Round(v[2] / step)),
	}
}
