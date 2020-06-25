package domain_test

import (
	"github.com/b2wdigital/restQL-golang/internal/domain"
	"github.com/b2wdigital/restQL-golang/test"
	"testing"
)

func TestMappingsPathWithParams(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		params   map[string]interface{}
		expected string
	}{
		{
			"should do nothing if there is no path param",
			"http://hero.api/hero",
			nil,
			"/hero",
		},
		{
			"should replace single path param",
			"http://hero.api/hero/:id",
			map[string]interface{}{"id": "12345"},
			"/hero/12345",
		},
		{
			"should replace multiple path param",
			"http://hero.api/hero/:id/:name",
			map[string]interface{}{"id": "12345", "name": "batman"},
			"/hero/12345/batman",
		},
		{
			"should replace multiple interspersed path param",
			"http://hero.api/hero/:id/info/:name",
			map[string]interface{}{"id": "12345", "name": "batman"},
			"/hero/12345/info/batman",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping, err := domain.NewMapping("test-resource", tt.url)
			test.VerifyError(t, err)

			got := mapping.PathWithParams(tt.params)
			test.Equal(t, got, tt.expected)
		})
	}
}
