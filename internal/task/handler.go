package task

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Handler struct{ repo *Repo }

func NewHandler(repo *Repo) *Handler { return &Handler{repo: repo} }
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	return r
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	tasks := h.repo.List()
	query := r.URL.Query()
	if query.Has("done") {
		raw := query.Get("done")
		if raw != "true" && raw != "false" {
			httpError(w, 400, "done must be true or false")
			return
		}
		done := raw == "true"
		filtered := tasks[:0]
		for _, t := range tasks {
			if t.Done == done {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}
	if query.Has("page") || query.Has("limit") {
		page, limit := 1, 10
		var err error
		if query.Has("page") {
			page, err = strconv.Atoi(query.Get("page"))
			if err != nil || page <= 0 {
				httpError(w, 400, "page must be positive")
				return
			}
		}
		if query.Has("limit") {
			limit, err = strconv.Atoi(query.Get("limit"))
			if err != nil || limit <= 0 {
				httpError(w, 400, "limit must be positive")
				return
			}
		}
		// Проверка до умножения защищает от переполнения и выхода за границы.
		if len(tasks) == 0 || page-1 > (len(tasks)-1)/limit {
			tasks = tasks[:0]
		} else {
			start := (page - 1) * limit
			end := len(tasks)
			if limit < len(tasks)-start {
				end = start + limit
			}
			tasks = tasks[start:end]
		}
	}
	writeJSON(w, http.StatusOK, tasks)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	t, err := h.repo.Get(id)
	if err != nil {
		repoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

type createReq struct {
	Title string `json:"title"`
}

func validTitle(w http.ResponseWriter, title string) bool {
	length := utf8.RuneCountInString(title)
	if length < 3 || length > 100 {
		httpError(w, 400, "title must contain 3 to 100 characters")
		return false
	}
	return true
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, 400, "invalid json")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if !validTitle(w, req.Title) {
		return
	}
	t, err := h.repo.Create(req.Title)
	if err != nil {
		repoError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

type updateReq struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req updateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, 400, "invalid json")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if !validTitle(w, req.Title) {
		return
	}
	t, err := h.repo.Update(id, req.Title, req.Done)
	if err != nil {
		repoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.repo.Delete(id); err != nil {
		repoError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		httpError(w, 400, "invalid id")
		return 0, false
	}
	return id, true
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func httpError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
func repoError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		httpError(w, 404, "task not found")
	} else {
		httpError(w, 500, "cannot save tasks")
	}
}
