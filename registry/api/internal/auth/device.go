package auth

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rithul/hivemind/registry/api/internal/db"
)

type DeviceRequest struct {
	DeviceCode string    `json:"device_code"`
	UserCode   string    `json:"user_code"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Status     string    `json:"status"` // "pending", "approved", "denied"
	Token      string    `json:"token,omitempty"`
}

type DeviceAuthManager struct {
	requests sync.Map
	queries  *db.Queries
	baseURL  string
	count    int64 // atomic count of active device requests
}

const maxDeviceRequests = 10000

func NewDeviceAuthManager(q *db.Queries, baseURL string) *DeviceAuthManager {
	m := &DeviceAuthManager{
		queries: q,
		baseURL: baseURL,
	}
	go m.cleanupLoop()
	return m
}

func (m *DeviceAuthManager) cleanupLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		m.requests.Range(func(key, value interface{}) bool {
			req := value.(*DeviceRequest)
			if now.After(req.ExpiresAt) {
				m.requests.Delete(key)
				atomic.AddInt64(&m.count, -1)
			}
			return true
		})
	}
}

func randomUserCode() string {
	b := make([]byte, 2)
	rand.Read(b)
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	prefix := string([]byte{chars[b[0]%26], chars[b[1]%26], chars[(b[0]+b[1])%26], chars[(b[0]*b[1])%26]})

	digits := make([]byte, 2)
	rand.Read(digits)
	num := (int(digits[0])<<8 | int(digits[1])) % 10000
	return fmt.Sprintf("%s-%04d", prefix, num)
}

func (m *DeviceAuthManager) RequestDeviceCode(w http.ResponseWriter, r *http.Request) {
	// H12: Reject new device requests if store is at capacity to prevent OOM.
	if atomic.LoadInt64(&m.count) >= maxDeviceRequests {
		http.Error(w, "too many pending device requests, try again later", http.StatusServiceUnavailable)
		return
	}

	deviceCode := uuid.New().String()
	userCode := randomUserCode()

	req := &DeviceRequest{
		DeviceCode: deviceCode,
		UserCode:   userCode,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(5 * time.Minute),
		Status:     "pending",
	}

	m.requests.Store(deviceCode, req)
	atomic.AddInt64(&m.count, 1)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"device_code":      deviceCode,
		"user_code":        userCode,
		"verification_uri": m.baseURL + "/activate",
		"expires_in":       300,
		"interval":         5,
	})
}

func (m *DeviceAuthManager) PollDeviceCode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DeviceCode string `json:"device_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	val, ok := m.requests.Load(body.DeviceCode)
	if !ok {
		w.WriteHeader(http.StatusGone)
		json.NewEncoder(w).Encode(map[string]string{"status": "expired"})
		return
	}

	req := val.(*DeviceRequest)
	if time.Now().After(req.ExpiresAt) {
		m.requests.Delete(body.DeviceCode)
		atomic.AddInt64(&m.count, -1)
		w.WriteHeader(http.StatusGone)
		json.NewEncoder(w).Encode(map[string]string{"status": "expired"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	switch req.Status {
	case "pending":
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "pending"})
	case "denied":
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "denied"})
	case "approved":
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"token": req.Token})
		// one-time use
		m.requests.Delete(body.DeviceCode)
		atomic.AddInt64(&m.count, -1)
	}
}

func (m *DeviceAuthManager) ApproveDevice(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		UserCode string `json:"user_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var found *DeviceRequest
	var foundKey string
	m.requests.Range(func(key, value interface{}) bool {
		req := value.(*DeviceRequest)
		if req.UserCode == body.UserCode && time.Now().Before(req.ExpiresAt) {
			found = req
			foundKey = key.(string)
			return false // break
		}
		return true
	})

	if found == nil {
		http.Error(w, "invalid or expired user code", http.StatusNotFound)
		return
	}

	// Create API key
	raw, hashHex, prefix, err := GenerateAPIKey()
	if err != nil {
		http.Error(w, "internal error generating key", http.StatusInternalServerError)
		return
	}

	var orgID pgtype.UUID
	_, err = m.queries.CreateAPIKey(r.Context(), db.CreateAPIKeyParams{
		UserID:    userID,
		OrgID:     orgID,
		Name:      "CLI login",
		KeyHash:   hashHex,
		KeyPrefix: prefix,
		Scopes:    []string{"publish", "read"},
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(90 * 24 * time.Hour), Valid: true},
	})
	if err != nil {
		http.Error(w, "internal error generating key", http.StatusInternalServerError)
		return
	}

	found.Status = "approved"
	found.Token = raw
	m.requests.Store(foundKey, found)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
