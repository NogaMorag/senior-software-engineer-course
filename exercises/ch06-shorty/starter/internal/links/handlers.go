package links

import (
	"log/slog"
	"net/http"
	"time"

	"example.com/ch06-shorty/internal/platform"
)

// Handler serves the shortener's HTTP API.
type Handler struct {
	store   Store
	baseURL string // public address of the service, without a trailing slash
	maxBody int64
	log     *slog.Logger
}

func NewHandler(store Store, baseURL string, maxBody int64, log *slog.Logger) *Handler {
	return &Handler{store: store, baseURL: baseURL, maxBody: maxBody, log: log}
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
	// TODO 2a (day 11, section 6.2):
	//   - decode with platform.DecodeJSON[createRequest](w, r, h.maxBody);
	//     on error, platform.RespondError and return;
	//   - checkURL(req.URL), and checkCode(req.Code) when a code was given:
	//     failures go out as badRequest(err);
	//   - with no code, try newCode() up to 3 times while Create says ErrExists;
	//   - ErrExists at the end is a 409 *platform.HTTPError "code already in use";
	//   - log "link created" with InfoContext, set Location: /v1/links/<code>,
	//     and platform.WriteJSON(w, http.StatusCreated, h.response(link)).
	panic("TODO 2a: create")
}

// resolve handles GET /{code}: a redirect to the stored URL.
func (h *Handler) resolve(w http.ResponseWriter, r *http.Request) {
	// TODO 2b (day 11): h.store.Resolve(r.Context(), r.PathValue("code"));
	// ErrNotFound is a 404 "no such link"; success is http.Redirect with 302.
	panic("TODO 2b: resolve")
}

// stats handles GET /v1/links/{code}: the link and its hit count.
func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	// TODO 2c (day 11): like resolve, but h.store.Get and a 200 JSON response.
	panic("TODO 2c: stats")
}
