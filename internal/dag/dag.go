package dag

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"vnodex/internal/model"
)

var (
	ErrEmptyHash        = errors.New("dag: event hash is required")
	ErrDuplicateEvent   = errors.New("dag: duplicate event hash")
	ErrMissingParent    = errors.New("dag: missing parent hash")
	ErrSelfParent       = errors.New("dag: event cannot reference itself as a parent")
	ErrDuplicateParent  = errors.New("dag: duplicate parent hash")
	ErrEmptyParent      = errors.New("dag: empty parent hash")
	ErrCycleDetected    = errors.New("dag: cycle detected")
)

type DAG struct {
	mu       sync.RWMutex
	nodes    map[string]model.Event
	parents  map[string]map[string]struct{}
	children map[string]map[string]struct{}
}

func New() *DAG {
	return &DAG{
		nodes:    make(map[string]model.Event),
		parents:  make(map[string]map[string]struct{}),
		children: make(map[string]map[string]struct{}),
	}
}

func NewFromEvents(events map[string]model.Event) (*DAG, error) {
	d := New()
	for _, event := range events {
		if err := d.Add(event); err != nil {
			return nil, err
		}
	}
	return d, nil
}

func (d *DAG) Add(event model.Event) error {
	if event.EventHash == "" {
		return ErrEmptyHash
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.nodes[event.EventHash]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateEvent, event.EventHash)
	}

	if _, ok := d.parents[event.EventHash]; !ok {
		d.parents[event.EventHash] = make(map[string]struct{})
	}
	if _, ok := d.children[event.EventHash]; !ok {
		d.children[event.EventHash] = make(map[string]struct{})
	}

	seenParents := make(map[string]struct{}, len(event.ParentHashes))
	for _, parent := range event.ParentHashes {
		if parent == "" {
			return ErrEmptyParent
		}
		if parent == event.EventHash {
			return ErrSelfParent
		}
		if _, ok := seenParents[parent]; ok {
			return ErrDuplicateParent
		}
		seenParents[parent] = struct{}{}

		d.parents[event.EventHash][parent] = struct{}{}
		if _, ok := d.children[parent]; !ok {
			d.children[parent] = make(map[string]struct{})
		}
		d.children[parent][event.EventHash] = struct{}{}
	}

	d.nodes[event.EventHash] = event
	return nil
}

func (d *DAG) Has(hash string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, ok := d.nodes[hash]
	return ok
}

func (d *DAG) Event(hash string) (model.Event, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	ev, ok := d.nodes[hash]
	return ev, ok
}

func (d *DAG) MissingParents(event model.Event) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	missing := make([]string, 0)
	seen := make(map[string]struct{}, len(event.ParentHashes))
	for _, parent := range event.ParentHashes {
		if parent == "" {
			continue
		}
		if _, ok := seen[parent]; ok {
			continue
		}
		seen[parent] = struct{}{}
		if _, ok := d.nodes[parent]; !ok {
			missing = append(missing, parent)
		}
	}
	sort.Strings(missing)
	return missing
}

func (d *DAG) Heads() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	heads := make([]string, 0)
	for hash := range d.nodes {
		children := d.children[hash]
		if len(children) == 0 {
			heads = append(heads, hash)
		}
	}
	sort.Strings(heads)
	return heads
}

func (d *DAG) Validate() error {
	_, err := d.TopoOrder()
	return err
}

func (d *DAG) TopoOrder() ([]model.Event, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.nodes) == 0 {
		return []model.Event{}, nil
	}

	indegree := make(map[string]int, len(d.nodes))
	for hash := range d.nodes {
		indegree[hash] = 0
	}

	for hash, event := range d.nodes {
		seen := make(map[string]struct{}, len(event.ParentHashes))
		for _, parent := range event.ParentHashes {
			if parent == "" {
				return nil, ErrEmptyParent
			}
			if parent == hash {
				return nil, ErrSelfParent
			}
			if _, ok := seen[parent]; ok {
				return nil, ErrDuplicateParent
			}
			seen[parent] = struct{}{}
			if _, ok := d.nodes[parent]; !ok {
				return nil, fmt.Errorf("%w: %s", ErrMissingParent, parent)
			}
			indegree[hash]++
		}
	}

	queue := make([]model.Event, 0)
	for hash, deg := range indegree {
		if deg == 0 {
			queue = append(queue, d.nodes[hash])
		}
	}
	sort.Slice(queue, func(i, j int) bool {
		return eventLess(queue[i], queue[j])
	})

	out := make([]model.Event, 0, len(d.nodes))
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		out = append(out, current)

		for childHash := range d.children[current.EventHash] {
			indegree[childHash]--
			if indegree[childHash] == 0 {
				queue = insertEventSorted(queue, d.nodes[childHash])
			}
		}
	}

	if len(out) != len(d.nodes) {
		return nil, ErrCycleDetected
	}

	return out, nil
}

func eventLess(a, b model.Event) bool {
	if a.LogicalOrder != b.LogicalOrder {
		return a.LogicalOrder < b.LogicalOrder
	}
	if a.TimeCreated.UnixNano() != b.TimeCreated.UnixNano() {
		return a.TimeCreated.UnixNano() < b.TimeCreated.UnixNano()
	}
	if a.EventHash != b.EventHash {
		return a.EventHash < b.EventHash
	}
	return a.EventID < b.EventID
}

func insertEventSorted(queue []model.Event, event model.Event) []model.Event {
	queue = append(queue, event)
	sort.Slice(queue, func(i, j int) bool {
		return eventLess(queue[i], queue[j])
	})
	return queue
}