package model

import (
	"errors"
	"math"
)

type Vector []float64

func ZeroVector(dim int) Vector {
	if dim < 0 {
		dim = 0
	}
	return make(Vector, dim)
}

func (v Vector) Clone() Vector {
	out := make(Vector, len(v))
	copy(out, v)
	return out
}

func (v Vector) Dimension() int { return len(v) }

func (v Vector) IsZero() bool {
	for _, x := range v {
		if x != 0 {
			return false
		}
	}
	return true
}

func (v Vector) Magnitude() float64 {
	var sum float64
	for _, x := range v {
		sum += x
	}
	return sum
}

func (v Vector) EuclideanNorm() float64 {
	var sum float64
	for _, x := range v {
		sum += x * x
	}
	return math.Sqrt(sum)
}

func (v Vector) Direction() Vector {
	total := v.Magnitude()
	if total == 0 {
		return ZeroVector(len(v))
	}
	out := make(Vector, len(v))
	for i, x := range v {
		out[i] = x / total
	}
	return out
}

func (v Vector) Add(b Vector) Vector {
	dim := max(len(v), len(b))
	out := make(Vector, dim)
	for i := 0; i < dim; i++ {
		if i < len(v) {
			out[i] += v[i]
		}
		if i < len(b) {
			out[i] += b[i]
		}
	}
	return out
}

func (v Vector) Sub(b Vector) Vector {
	dim := max(len(v), len(b))
	out := make(Vector, dim)
	for i := 0; i < dim; i++ {
		if i < len(v) {
			out[i] += v[i]
		}
		if i < len(b) {
			out[i] -= b[i]
		}
	}
	return out
}

func (v Vector) Scale(k float64) Vector {
	out := make(Vector, len(v))
	for i, x := range v {
		out[i] = x * k
	}
	return out
}

func (v Vector) Normalize() (Vector, error) {
	if v.IsZero() {
		return nil, errors.New("cannot normalize zero vector")
	}
	return v.Direction(), nil
}

func (v Vector) ProjectOnto(axis Vector) (Vector, error) {
	norm := axis.EuclideanNorm()
	if norm == 0 {
		return nil, errors.New("cannot project onto zero vector")
	}
	var dot float64
	dim := max(len(v), len(axis))
	for i := 0; i < dim; i++ {
		if i < len(v) && i < len(axis) {
			dot += v[i] * axis[i]
		}
	}
	scale := dot / (norm * norm)
	return axis.Scale(scale), nil
}

func (v Vector) Rotate2D(theta float64) (Vector, error) {
	if len(v) < 2 {
		return nil, errors.New("rotate2d requires at least two components")
	}
	out := v.Clone()
	cosT, sinT := math.Cos(theta), math.Sin(theta)
	x, y := v[0], v[1]
	out[0] = x*cosT - y*sinT
	out[1] = x*sinT + y*cosT
	return out, nil
}

func (v Vector) Constrain(minV, maxV Vector) Vector {
	dim := max(len(v), max(len(minV), len(maxV)))
	out := make(Vector, dim)
	for i := 0; i < dim; i++ {
		val := 0.0
		if i < len(v) {
			val = v[i]
		}
		lo := math.Inf(-1)
		hi := math.Inf(1)
		if i < len(minV) {
			lo = minV[i]
		}
		if i < len(maxV) {
			hi = maxV[i]
		}
		if val < lo {
			val = lo
		}
		if val > hi {
			val = hi
		}
		out[i] = val
	}
	return out
}

func (v Vector) Nullify() Vector { return ZeroVector(len(v)) }

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
