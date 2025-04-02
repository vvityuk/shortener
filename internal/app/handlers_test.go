package app

import (
	"net/http"
	"testing"
)

func TestHandler_GetURL(t *testing.T) {
	type fields struct {
		service *Service
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
			}
			h.GetURL(tt.args.w, tt.args.r)
		})
	}
}
