package clubscore

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	ErrDuplicateID = errors.New("activity id already exists")
	ErrNotFound    = errors.New("activity not found")
)

type Activity struct {
	ID           string  `yaml:"id"`
	Name         string  `yaml:"name"`
	Leader       string  `yaml:"leader"`
	Score        float64 `yaml:"score"`
	Participants int     `yaml:"participants"`
}

type ActivityNode struct {
	Activity Activity
	Prev     *ActivityNode
	Next     *ActivityNode
}

type ActivityList struct {
	mu    sync.RWMutex
	head  *ActivityNode
	tail  *ActivityNode
	index map[string]*ActivityNode
}

type archive struct {
	Activities []Activity `yaml:"activities"`
}

func NewList() *ActivityList {
	return &ActivityList{index: make(map[string]*ActivityNode)}
}

func validateActivity(activity Activity) error {
	if activity.ID == "" {
		return errors.New("activity id is required")
	}
	if activity.Name == "" {
		return errors.New("activity name is required")
	}
	if activity.Leader == "" {
		return errors.New("activity leader is required")
	}
	if activity.Score < 0 || activity.Score > 100 {
		return errors.New("activity score must be between 0 and 100")
	}
	if activity.Participants < 0 {
		return errors.New("activity participants cannot be negative")
	}
	return nil
}

func (l *ActivityList) Append(activity Activity) error {
	if err := validateActivity(activity); err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, exists := l.index[activity.ID]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateID, activity.ID)
	}
	node := &ActivityNode{Activity: activity}
	l.index[activity.ID] = node
	if l.tail == nil {
		l.head = node
		l.tail = node
		return nil
	}
	node.Prev = l.tail
	l.tail.Next = node
	l.tail = node
	return nil
}

func (l *ActivityList) Delete(id string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	node, exists := l.index[id]
	if !exists {
		return fmt.Errorf("%w: %s", ErrNotFound, id)
	}
	if node.Prev == nil {
		l.head = node.Next
	} else {
		node.Prev.Next = node.Next
	}
	if node.Next == nil {
		l.tail = node.Prev
	} else {
		node.Next.Prev = node.Prev
	}
	delete(l.index, id)
	node.Prev = nil
	node.Next = nil
	return nil
}

func (l *ActivityList) Modify(id string, updated Activity) error {
	if err := validateActivity(updated); err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	node, exists := l.index[id]
	if !exists {
		return fmt.Errorf("%w: %s", ErrNotFound, id)
	}
	if updated.ID != id {
		if _, used := l.index[updated.ID]; used {
			return fmt.Errorf("%w: %s", ErrDuplicateID, updated.ID)
		}
		delete(l.index, id)
		l.index[updated.ID] = node
	}
	node.Activity = updated
	return nil
}

func (l *ActivityList) Query(id string) (Activity, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	node, exists := l.index[id]
	if !exists {
		return Activity{}, false
	}
	return node.Activity, true
}

func (l *ActivityList) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.index)
}

func (l *ActivityList) ToSlice() []Activity {
	l.mu.RLock()
	defer l.mu.RUnlock()
	activities := make([]Activity, 0, len(l.index))
	for node := l.head; node != nil; node = node.Next {
		activities = append(activities, node.Activity)
	}
	return activities
}

func (l *ActivityList) SortByID() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sortLocked(func(left, right Activity) bool {
		return left.ID < right.ID
	})
}

func (l *ActivityList) SortByScore() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sortLocked(func(left, right Activity) bool {
		if left.Score == right.Score {
			return left.ID < right.ID
		}
		return left.Score > right.Score
	})
}

func (l *ActivityList) sortLocked(less func(Activity, Activity) bool) {
	nodes := make([]*ActivityNode, 0, len(l.index))
	for node := l.head; node != nil; node = node.Next {
		nodes = append(nodes, node)
	}
	sort.SliceStable(nodes, func(i, j int) bool {
		return less(nodes[i].Activity, nodes[j].Activity)
	})
	if len(nodes) == 0 {
		l.head = nil
		l.tail = nil
		return
	}
	l.head = nodes[0]
	l.head.Prev = nil
	for i := 1; i < len(nodes); i++ {
		nodes[i-1].Next = nodes[i]
		nodes[i].Prev = nodes[i-1]
	}
	l.tail = nodes[len(nodes)-1]
	l.tail.Next = nil
}

func (l *ActivityList) Save(w io.Writer) error {
	if w == nil {
		return errors.New("archive writer is nil")
	}
	data, err := yaml.Marshal(archive{Activities: l.ToSlice()})
	if err != nil {
		return fmt.Errorf("encode archive: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write archive: %w", err)
	}
	return nil
}

func (l *ActivityList) SaveFile(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open archive for writing: %w", err)
	}
	defer file.Close()
	return l.Save(file)
}

func Load(r io.Reader) (*ActivityList, error) {
	if r == nil {
		return nil, errors.New("archive reader is nil")
	}
	var saved archive
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&saved); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("decode archive: %w", err)
		}
	}
	// Always initialize the index map, even for an empty archive. A nil map
	// would panic on the first Append (assignment to entry in nil map) once a
	// valid empty archive has been loaded.
	list := &ActivityList{index: make(map[string]*ActivityNode, len(saved.Activities))}
	for _, activity := range saved.Activities {
		if err := list.Append(activity); err != nil {
			return nil, fmt.Errorf("load activity %s: %w", activity.ID, err)
		}
	}
	return list, nil
}

func LoadFile(path string) (*ActivityList, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()
	return Load(file)
}
