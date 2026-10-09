package task

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"sync"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Repo struct {
	mu    sync.RWMutex
	seq   int64
	items map[int64]*Task
	path  string
}

func NewRepo(path string) (*Repo, error) {
	r := &Repo{items: make(map[int64]*Task), path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return r, nil
	}
	if err != nil {
		return nil, err
	}
	var tasks []*Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	for _, t := range tasks {
		if t == nil || t.ID <= 0 {
			return nil, errors.New("invalid task in JSON file")
		}
		r.items[t.ID] = t
		if t.ID > r.seq {
			r.seq = t.ID
		}
	}
	return r, nil
}
func (r *Repo) List() []*Task {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Task, 0, len(r.items))
	for _, t := range r.items {
		copy := *t
		out = append(out, &copy)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (r *Repo) Get(id int64) (*Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *t
	return &copy, nil
}

// Вызывается только при удержании блокировки на запись.
func (r *Repo) save() error {
	tasks := make([]*Task, 0, len(r.items))
	for _, t := range r.items {
		tasks = append(tasks, t)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(r.path+".tmp", data, 0644); err != nil {
		return err
	}
	return os.Rename(r.path+".tmp", r.path)
}
func (r *Repo) Create(title string) (*Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	now := time.Now()
	t := &Task{ID: r.seq, Title: title, Done: false, CreatedAt: now, UpdatedAt: now}
	r.items[t.ID] = t
	if err := r.save(); err != nil {
		delete(r.items, t.ID)
		r.seq--
		return nil, err
	}
	copy := *t
	return &copy, nil
}
func (r *Repo) Update(id int64, title string, done bool) (*Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	old := *t
	t.Title = title
	t.Done = done
	t.UpdatedAt = time.Now()
	if err := r.save(); err != nil {
		*t = old
		return nil, err
	}
	copy := *t
	return &copy, nil
}
func (r *Repo) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.items[id]
	if !ok {
		return ErrNotFound
	}
	delete(r.items, id)
	if err := r.save(); err != nil {
		r.items[id] = old
		return err
	}
	return nil
}
