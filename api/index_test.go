package handler

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_Index(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		method     string
		expectCode int
		contains   string
	}{
		{
			name:       "Rewritten root path",
			url:        "/api/index?__path=/",
			method:     "GET",
			expectCode: 200,
			contains:   "Equivalência CUB",
		},
		{
			name:       "Rewritten double slash",
			url:        "/api/index?__path=//",
			method:     "GET",
			expectCode: 200,
			contains:   "Equivalência CUB",
		},
		{
			name:       "Direct function invocation without param",
			url:        "/api/index",
			method:     "GET",
			expectCode: 200,
			contains:   "Equivalência CUB",
		},
		{
			name:       "Direct root invocation",
			url:        "/",
			method:     "GET",
			expectCode: 200,
			contains:   "Equivalência CUB",
		},
		{
			name:       "Static CSS",
			url:        "/api/index?__path=/static/css/styles.css",
			method:     "GET",
			expectCode: 200,
			contains:   "tailwindcss",
		},
		{
			name:       "Static JS",
			url:        "/api/index?__path=/static/js/htmx.min.js",
			method:     "GET",
			expectCode: 200,
			contains:   "htmx",
		},
		{
			name:       "API Propostas",
			url:        "/api/index?__path=/api/propostas",
			method:     "GET",
			expectCode: 200,
			contains:   "Propostas Salvas",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			rec := httptest.NewRecorder()

			Handler(rec, req)

			t.Logf("Test %s: status=%d, bodyLen=%d", tt.name, rec.Code, rec.Body.Len())

			if rec.Code != tt.expectCode {
				t.Errorf("expected status %d, got %d, body: %s", tt.expectCode, rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tt.contains) {
				t.Errorf("body does not contain %q, body length: %d", tt.contains, rec.Body.Len())
			}
		})
	}
}
