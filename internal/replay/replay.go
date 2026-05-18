package replay

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"vnodex/internal/record"
)

var (
	ErrMissingEntityID = errors.New("replay: missing entity id")
	ErrUnknownOp       = errors.New("replay: unknown operation")
)

type EntityState struct {
	Position      [3]float64 `json:"position"`
	Velocity      [3]float64 `json:"velocity"`
	RegionID      string     `json:"region_id"`
	LastRecordID  string     `json:"last_record_id"`
	UpdatedAtNano int64      `json:"updated_at_nano"`
}

type RegionState struct {
	EntityCount   int    `json:"entity_count"`
	LastRecordID  string `json:"last_record_id"`
	UpdatedAtNano int64  `json:"updated_at_nano"`
}

type State struct {
	Entities map[string]EntityState `json:"entities"`
	Regions  map[string]RegionState `json:"regions"`
}

type Engine struct {
	mu    sync.RWMutex
	state State
}

type OrderedSource interface {
	TopologicalOrder() ([]record.Record, error)
}

func NewEngine() *Engine {
	return &Engine{
		state: State{
			Entities: make(map[string]EntityState),
			Regions:  make(map[string]RegionState),
		},
	}
}

func (e *Engine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.state = State{
		Entities: make(map[string]EntityState),
		Regions:  make(map[string]RegionState),
	}
}

func (e *Engine) Snapshot() State {
	e.mu.RLock()
	defer e.mu.RUnlock()

	out := State{
		Entities: make(map[string]EntityState, len(e.state.Entities)),
		Regions:  make(map[string]RegionState, len(e.state.Regions)),
	}
	for k, v := range e.state.Entities {
		out.Entities[k] = v
	}
	for k, v := range e.state.Regions {
		out.Regions[k] = v
	}
	return out
}

func (e *Engine) Apply(r record.Record) error {
	if strings.TrimSpace(r.EntityID) == "" {
		return ErrMissingEntityID
	}
	if err := r.Validate(); err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	current := e.state.Entities[r.EntityID]
	current.RegionID = r.RegionID
	current.LastRecordID = r.ID
	current.UpdatedAtNano = r.TimestampUnixNano

	op := strings.ToLower(strings.TrimSpace(r.Operation))
	switch op {
	case "set":
		current.Position = r.Vector
	case "move", "add":
		current.Position[0] += r.Vector[0]
		current.Position[1] += r.Vector[1]
		current.Position[2] += r.Vector[2]
	case "subtract", "sub":
		current.Position[0] -= r.Vector[0]
		current.Position[1] -= r.Vector[1]
		current.Position[2] -= r.Vector[2]
	case "scale":
		factor := scalarFromVector(r.Vector)
		current.Position[0] *= factor
		current.Position[1] *= factor
		current.Position[2] *= factor
	case "normalize":
		current.Position = normalize(current.Position)
	case "constrain":
		current.Position = constrain(current.Position, r.Vector)
	case "nullify", "zero":
		current.Position = [3]float64{}
	case "project":
		current.Position = project(current.Position, r.Vector)
	case "rotate":
		current.Position = rotateZ(current.Position, r.Vector[2])
	default:
		return fmt.Errorf("%w: %s", ErrUnknownOp, op)
	}

	e.state.Entities[r.EntityID] = current

	if r.RegionID != "" {
		region := e.state.Regions[r.RegionID]
		region.LastRecordID = r.ID
		region.UpdatedAtNano = r.TimestampUnixNano
		if op == "set" || op == "move" || op == "add" || op == "subtract" || op == "scale" || op == "normalize" || op == "constrain" || op == "nullify" || op == "zero" || op == "project" || op == "rotate" {
			region.EntityCount = countEntitiesInRegion(e.state.Entities, r.RegionID)
		}
		e.state.Regions[r.RegionID] = region
	}

	return nil
}

func (e *Engine) Replay(records []record.Record) (State, error) {
	e.Reset()
	for _, r := range records {
		if err := e.Apply(r); err != nil {
			return State{}, err
		}
	}
	return e.Snapshot(), nil
}

func (e *Engine) ReplayOrderedSource(src OrderedSource) (State, error) {
	ordered, err := src.TopologicalOrder()
	if err != nil {
		return State{}, err
	}
	return e.Replay(ordered)
}

func (s State) Hash() (string, error) {
	wire := struct {
		Entities []entityWire `json:"entities"`
		Regions  []regionWire `json:"regions"`
	}{
		Entities: make([]entityWire, 0, len(s.Entities)),
		Regions:  make([]regionWire, 0, len(s.Regions)),
	}

	entityIDs := make([]string, 0, len(s.Entities))
	for id := range s.Entities {
		entityIDs = append(entityIDs, id)
	}
	sort.Strings(entityIDs)

	for _, id := range entityIDs {
		st := s.Entities[id]
		wire.Entities = append(wire.Entities, entityWire{
			ID:            id,
			Position:      st.Position,
			Velocity:      st.Velocity,
			RegionID:      st.RegionID,
			LastRecordID:  st.LastRecordID,
			UpdatedAtNano: st.UpdatedAtNano,
		})
	}

	regionIDs := make([]string, 0, len(s.Regions))
	for id := range s.Regions {
		regionIDs = append(regionIDs, id)
	}
	sort.Strings(regionIDs)

	for _, id := range regionIDs {
		st := s.Regions[id]
		wire.Regions = append(wire.Regions, regionWire{
			ID:            id,
			EntityCount:   st.EntityCount,
			LastRecordID:  st.LastRecordID,
			UpdatedAtNano: st.UpdatedAtNano,
		})
	}

	b, err := json.Marshal(wire)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

type entityWire struct {
	ID            string     `json:"id"`
	Position      [3]float64 `json:"position"`
	Velocity      [3]float64 `json:"velocity"`
	RegionID      string     `json:"region_id"`
	LastRecordID  string     `json:"last_record_id"`
	UpdatedAtNano int64      `json:"updated_at_nano"`
}

type regionWire struct {
	ID            string `json:"id"`
	EntityCount   int    `json:"entity_count"`
	LastRecordID  string `json:"last_record_id"`
	UpdatedAtNano int64  `json:"updated_at_nano"`
}

func scalarFromVector(v [3]float64) float64 {
	if v[0] != 0 {
		return v[0]
	}
	if v[1] != 0 {
		return v[1]
	}
	if v[2] != 0 {
		return v[2]
	}
	return 1
}

func normalize(v [3]float64) [3]float64 {
	mag := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
	if mag == 0 {
		return v
	}
	return [3]float64{v[0] / mag, v[1] / mag, v[2] / mag}
}

func constrain(v, limit [3]float64) [3]float64 {
	out := v
	for i := 0; i < 3; i++ {
		max := math.Abs(limit[i])
		if out[i] > max {
			out[i] = max
		}
		if out[i] < -max {
			out[i] = -max
		}
	}
	return out
}

func project(a, b [3]float64) [3]float64 {
	denom := b[0]*b[0] + b[1]*b[1] + b[2]*b[2]
	if denom == 0 {
		return [3]float64{}
	}
	scale := (a[0]*b[0] + a[1]*b[1] + a[2]*b[2]) / denom
	return [3]float64{b[0] * scale, b[1] * scale, b[2] * scale}
}

func rotateZ(v [3]float64, radians float64) [3]float64 {
	if radians == 0 {
		return v
	}
	c := math.Cos(radians)
	s := math.Sin(radians)
	return [3]float64{
		(v[0] * c) - (v[1] * s),
		(v[0] * s) + (v[1] * c),
		v[2],
	}
}

func countEntitiesInRegion(entities map[string]EntityState, regionID string) int {
	count := 0
	for _, st := range entities {
		if st.RegionID == regionID {
			count++
		}
	}
	return count
}
