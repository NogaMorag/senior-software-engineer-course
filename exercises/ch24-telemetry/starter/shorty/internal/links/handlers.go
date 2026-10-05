package links

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"example.com/ch24-telemetry/internal/platform"
)

// Handler serves the shortener's HTTP API.
type Handler struct {
	store   Store
	baseURL string // public address of the service, without a trailing slash
	maxBody int64
	log     *slog.Logger
	publish PublisherFunc // nil: created links are not announced
}

// PublisherFunc announces a created link to the rest of the system (chapter
// 24: a Kafka record that linkworker consumes).
type PublisherFunc func(ctx context.Context, l Link) error

func NewHandler(store Store, baseURL string, maxBody int64, log *slog.Logger) *Handler {
	return &Handler{store: store, baseURL: baseURL, maxBody: maxBody, log: log}
}

// WithPublisher makes create announce every new link through p.
func (h *Handler) WithPublisher(p PublisherFunc) *Handler {
	h.publish = p
	return h
}

// Register adds the routes to mux. ServeMux answers 405 with an Allow header
// when the path matches but the method does not, and 404 when nothing matches.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/links", h.create)
	mux.HandleFunc("GET /v1/links/{code}", h.stats)
	mux.HandleFunc("GET /{code}", h.resolve)
}

type createRequest struct {
	URL  string `json:"url"`
	Code string `json:"code,omitempty"` // optional custom code
}

type linkResponse struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	ShortURL  string    `json:"short_url"`
	Hits      int64     `json:"hits"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) response(l Link) linkResponse {
	return linkResponse{Code: l.Code, URL: l.URL, ShortURL: h.baseURL + "/" + l.Code, Hits: l.Hits, CreatedAt: l.CreatedAt}
}

func badRequest(err error) error {
	return &platform.HTTPError{Status: http.StatusBadRequest, Message: err.Error()}
}

// create handles POST /v1/links with {"url": "...", "code": "optional"}.
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	req, err := platform.DecodeJSON[createRequest](w, r, h.maxBody)
	if err != nil {
		platform.RespondError(w, r, h.log, err)
		return
	}
	if err := checkURL(req.URL); err != nil {
		platform.RespondError(w, r, h.log, badRequest(err))
		return
	}
	link := Link{Code: req.Code, URL: req.URL, CreatedAt: time.Now().UTC()}
	if link.Code != "" {
		if err := checkCode(link.Code); err != nil {
			platform.RespondError(w, r, h.log, badRequest(err))
			return
		}
		err = h.store.Create(r.Context(), link)
	} else {
		// A random code can collide with an existing one; try a few times.
		for range 3 {
			link.Code = newCode()
			if err = h.store.Create(r.Context(), link); !errors.Is(err, ErrExists) {
				break
			}
		}
	}
	switch {
	case errors.Is(err, ErrExists):
		platform.RespondError(w, r, h.log, &platform.HTTPError{Status: http.StatusConflict, Message: "code already in use"})
		return
	case err != nil:
		platform.RespondError(w, r, h.log, err)
		return
	}
	if h.publish != nil {
		// Naive on purpose: the link is stored but its event may be lost. The
		// transactional outbox of chapter 13 removes this dual write.
		if err := h.publish(r.Context(), link); err != nil {
			h.log.ErrorContext(r.Context(), "publish failed", slog.String("code", link.Code), slog.Any("error", err))
			platform.WriteError(w, http.StatusServiceUnavailable, "could not announce the new link")
			return
		}
	}
	h.log.InfoContext(r.Context(), "link created", slog.String("code", link.Code))
	w.Header().Set("Location", "/v1/links/"+link.Code)
	platform.WriteJSON(w, http.StatusCreated, h.response(link))
}

// resolve handles GET /{code}: a redirect to the stored URL.
func (h *Handler) resolve(w http.ResponseWriter, r *http.Request) {
	link, err := h.store.Resolve(r.Context(), r.PathValue("code"))
	if errors.Is(err, ErrNotFound) {
		platform.WriteError(w, http.StatusNotFound, "no such link")
		return
	}
	if err != nil {
		platform.RespondError(w, r, h.log, err)
		return
	}
	http.Redirect(w, r, link.URL, http.StatusFound)
}

// stats handles GET /v1/links/{code}: the link and its hit count.
func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	link, err := h.store.Get(r.Context(), r.PathValue("code"))
	if errors.Is(err, ErrNotFound) {
		platform.WriteError(w, http.StatusNotFound, "no such link")
		return
	}
	if err != nil {
		platform.RespondError(w, r, h.log, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, h.response(link))
}
