package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/makeitshort/backend/internal/shortid"
)

const maxURLLength = 2048
const maxAliasLength = 32

var aliasPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func (s *Server) handleShorten() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeError(w, http.StatusBadRequest, "invalid request payload")
			return
		}

		req.URL = strings.TrimSpace(req.URL)
		req.Alias = strings.TrimSpace(req.Alias)
		if req.URL == "" {
			s.writeError(w, http.StatusBadRequest, "url is required")
			return
		}
		if len(req.URL) > maxURLLength {
			s.writeError(w, http.StatusBadRequest, "url exceeds maximum length of 2048 characters")
			return
		}
		parsedURL, err := url.ParseRequestURI(req.URL)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			s.writeError(w, http.StatusBadRequest, "invalid url format, must start with http:// or https://")
			return
		}

		if req.Alias != "" {
			if len(req.Alias) > maxAliasLength || !aliasPattern.MatchString(req.Alias) {
				s.writeError(w, http.StatusBadRequest, "alias must be 1-32 characters using letters, numbers, _ or -")
				return
			}
		}

		if req.ExpiresIn < 0 || (req.ExpiresIn > 0 && (req.ExpiresIn < 300 || req.ExpiresIn > 30*24*60*60)) {
			s.writeError(w, http.StatusBadRequest, "expires_in must be 0 or between 300 seconds and 30 days")
			return
		}

		id := req.Alias
		if id == "" {
			id = shortid.GenerateBase62()
		}

		ctx := r.Context()
		expiresAt := interface{}(nil)
		var expiresAtString *string
		if req.ExpiresIn > 0 {
			t := time.Now().UTC().Add(time.Duration(req.ExpiresIn) * time.Second)
			formatted := t.Format(time.RFC3339)
			expiresAt = t.Format(time.RFC3339)
			expiresAtString = &formatted
		}

		var results []interface{}
		err = s.supabase.DB.From("links").Insert(map[string]interface{}{
			"id":           id,
			"original_url": req.URL,
			"expires_at":   expiresAt,
		}).Execute(&results)
		if err != nil {
			s.logger.Error("failed to insert link into database", "error", err)
			if req.Alias != "" {
				s.writeError(w, http.StatusConflict, "alias is already in use")
				return
			}
			s.writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		ttl := 48 * time.Hour
		if req.ExpiresIn > 0 {
			ttl = time.Duration(req.ExpiresIn) * time.Second
		}
		if err := s.redis.Set(ctx, "link:"+id, req.URL, ttl).Err(); err != nil {
			s.logger.Error("failed to cache link in redis", "error", err)
		}

		baseURL := strings.TrimRight(s.baseURL, "/")
		resp := ShortenResponse{ID: id, ShortURL: baseURL + "/" + id, ExpiresAt: expiresAtString}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func (s *Server) handleStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			s.writeError(w, http.StatusBadRequest, "id is required")
			return
		}

		var rows []struct {
			OriginalURL   string     `json:"original_url"`
			ClickCount    int64      `json:"click_count"`
			LastClickedAt *time.Time `json:"last_clicked_at"`
			ExpiresAt     *time.Time `json:"expires_at"`
		}
		if err := s.supabase.DB.From("links").Select("original_url,click_count,last_clicked_at,expires_at").Eq("id", id).Execute(&rows); err != nil {
			s.logger.Error("failed to query link stats", "error", err, "id", id)
			s.writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if len(rows) == 0 {
			s.writeError(w, http.StatusNotFound, "link not found")
			return
		}

		row := rows[0]
		resp := StatsResponse{ID: id, OriginalURL: row.OriginalURL, ClickCount: row.ClickCount}
		if row.LastClickedAt != nil {
			v := row.LastClickedAt.UTC().Format(time.RFC3339)
			resp.LastClickedAt = &v
		}
		if row.ExpiresAt != nil {
			v := row.ExpiresAt.UTC().Format(time.RFC3339)
			resp.ExpiresAt = &v
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func (s *Server) handleRedirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			s.writeError(w, http.StatusNotFound, "link not found")
			return
		}

		ctx := r.Context()
		now := time.Now().UTC()

		// Redis provides the fast path. Metadata is checked in PostgreSQL when
		// a redirect needs expiry/click analytics updates.
		originalURL, redisErr := s.redis.Get(ctx, "link:"+id).Result()

		var row struct {
			OriginalURL string     `json:"original_url"`
			ExpiresAt   *time.Time `json:"expires_at"`
			ClickCount  int64      `json:"click_count"`
		}

		err := s.supabase.DB.From("links").Select("original_url,expires_at,click_count").Eq("id", id).Execute(&[]struct {
			OriginalURL string     `json:"original_url"`
			ExpiresAt   *time.Time `json:"expires_at"`
			ClickCount  int64      `json:"click_count"`
		}{row})
		if err != nil {
			s.logger.Error("failed to query link", "error", err, "id", id)
			s.writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		var row struct {
			OriginalURL string
			ExpiresAt   *time.Time
			ClickCount  int64
		}
		if len(rows) > 0 {
			row.OriginalURL = rows[0].OriginalURL
			row.ExpiresAt = rows[0].ExpiresAt
			row.ClickCount = rows[0].ClickCount
		}

		if row.OriginalURL == "" {
			if redisErr == nil && originalURL != "" {
				row.OriginalURL = originalURL
			} else {
				s.writeError(w, http.StatusNotFound, "link not found")
				return
			}
		}

		if row.ExpiresAt != nil && !now.Before(*row.ExpiresAt) {
			_ = s.redis.Del(ctx, "link:"+id).Err()
			s.writeError(w, http.StatusGone, "this short link has expired")
			return
		}

		// Persist lightweight analytics for every successful redirect.
		updated := map[string]interface{}{
			"click_count":     row.ClickCount + 1,
			"last_clicked_at": now.Format(time.RFC3339),
		}
		var updateResult []interface{}
		if err := s.supabase.DB.From("links").Update(updated).Eq("id", id).Execute(&updateResult); err != nil {
			s.logger.Error("failed to update click analytics", "error", err, "id", id)
		}

		if redisErr != nil || originalURL == "" {
			ttl := 48 * time.Hour
			if row.ExpiresAt != nil {
				ttl = time.Until(*row.ExpiresAt)
				if ttl <= 0 {
					s.writeError(w, http.StatusGone, "this short link has expired")
					return
				}
			}
			go func(cacheID, target string, cacheTTL time.Duration) {
				if err := s.redis.Set(context.Background(), "link:"+cacheID, target, cacheTTL).Err(); err != nil {
					s.logger.Error("failed to refresh redis cache", "error", err)
				}
			}(id, row.OriginalURL, ttl)
		}

		http.Redirect(w, r, row.OriginalURL, http.StatusFound)
	}
}
