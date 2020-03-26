package runner_test

import (
	"github.com/b2wdigital/restQL-golang/internal/domain"
	"github.com/b2wdigital/restQL-golang/internal/runner"
	"reflect"
	"testing"
	"time"
)

func TestNewDoneResource(t *testing.T) {
	tests := []struct {
		name     string
		request  domain.HttpRequest
		response domain.HttpResponse
		options  runner.DoneResourceOptions
		expected domain.DoneResource
	}{
		{
			"should create done resource for successful execution",
			domain.HttpRequest{},
			domain.HttpResponse{StatusCode: 200, Body: nil},
			runner.DoneResourceOptions{},
			domain.DoneResource{Details: domain.Details{Status: 200, Success: true, IgnoreErrors: false}, Result: nil},
		},
		{
			"should create done resource for failed execution",
			domain.HttpRequest{},
			domain.HttpResponse{StatusCode: 400, Body: nil},
			runner.DoneResourceOptions{},
			domain.DoneResource{Details: domain.Details{Status: 400, Success: false, IgnoreErrors: false}, Result: nil},
		},
		{
			"should create done resource with debug",
			domain.HttpRequest{
				Schema:  "http",
				Uri:     "hero.io/api",
				Query:   map[string]interface{}{"id": "123456"},
				Headers: map[string]string{"X-TID": "12345abdef"},
			},
			domain.HttpResponse{
				Url:        "http://hero.io/api",
				StatusCode: 200,
				Body:       nil,
				Headers:    map[string]string{"Content-Type": "application/json"},
				Duration:   100 * time.Millisecond,
			},
			runner.DoneResourceOptions{Debugging: true},
			domain.DoneResource{
				Details: domain.Details{
					Status:       200,
					Success:      true,
					IgnoreErrors: false,
					Debug: &domain.Debugging{
						Url:             "http://hero.io/api",
						RequestHeaders:  map[string]string{"X-TID": "12345abdef"},
						ResponseHeaders: map[string]string{"Content-Type": "application/json"},
						Params:          map[string]interface{}{"id": "123456"},
						ResponseTime:    100,
					},
				},
				Result: nil,
			},
		},
		{
			"should create done resource with ignore errors",
			domain.HttpRequest{},
			domain.HttpResponse{StatusCode: 200, Body: nil},
			runner.DoneResourceOptions{IgnoreErrors: true},
			domain.DoneResource{Details: domain.Details{Status: 200, Success: true, IgnoreErrors: true}, Result: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runner.NewDoneResource(tt.request, tt.response, tt.options)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NewDoneResource = %#+v, want = %#+v", got, tt.expected)
			}
		})
	}

}

func TestNewTimeoutResponse(t *testing.T) {
	timeoutErr := domain.ErrRequestTimeout

	tests := []struct {
		name     string
		request  domain.HttpRequest
		response domain.HttpResponse
		options  runner.DoneResourceOptions
		expected domain.DoneResource
	}{
		{
			"should create response for time outed execution",
			domain.HttpRequest{},
			domain.HttpResponse{},
			runner.DoneResourceOptions{},
			domain.DoneResource{
				Details: domain.Details{Status: 408, Success: false, IgnoreErrors: false},
				Result:  timeoutErr.Error(),
			},
		},
		{
			"should create response for time outed execution with debug",
			domain.HttpRequest{
				Schema:  "http",
				Uri:     "hero.io/api",
				Query:   map[string]interface{}{"id": "123456"},
				Headers: map[string]string{"X-TID": "12345abdef"},
			},
			domain.HttpResponse{
				Url:      "http://hero.io/api",
				Duration: 100 * time.Millisecond,
			},
			runner.DoneResourceOptions{Debugging: true},
			domain.DoneResource{
				Details: domain.Details{
					Status:       408,
					Success:      false,
					IgnoreErrors: false,
					Debug: &domain.Debugging{
						Url:            "http://hero.io/api",
						RequestHeaders: map[string]string{"X-TID": "12345abdef"},
						Params:         map[string]interface{}{"id": "123456"},
						ResponseTime:   100,
					},
				},
				Result: timeoutErr.Error(),
			},
		},
		{
			"should create response for time outed execution with debug",
			domain.HttpRequest{},
			domain.HttpResponse{},
			runner.DoneResourceOptions{IgnoreErrors: true},
			domain.DoneResource{
				Details: domain.Details{Status: 408, Success: false, IgnoreErrors: true},
				Result:  timeoutErr.Error(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runner.NewTimeoutResponse(timeoutErr, tt.request, tt.response, tt.options)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NewTimeoutResponse = %#+v, want = %#+v", got, tt.expected)
			}
		})
	}
}

func TestNewEmptyChainedResponse(t *testing.T) {
	t.Run("should create response for single empty chained param", func(t *testing.T) {
		params := []string{"id"}
		options := runner.DoneResourceOptions{}

		expected := domain.DoneResource{
			Details: domain.Details{Status: 400, Success: false, IgnoreErrors: false},
			Result:  "The request was skipped due to missing { :id } param value",
		}

		got := runner.NewEmptyChainedResponse(params, options)

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("NewEmptyChainedResponse = %#+v, want = %#+v", got, expected)
		}
	})

	t.Run("should create response for multiple empty chained param", func(t *testing.T) {
		params := []string{"id", "name", "city"}
		options := runner.DoneResourceOptions{}

		expected := domain.DoneResource{
			Details: domain.Details{Status: 400, Success: false, IgnoreErrors: false},
			Result:  "The request was skipped due to missing { :id :name :city } param value",
		}

		got := runner.NewEmptyChainedResponse(params, options)

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("NewEmptyChainedResponse = %#+v, want = %#+v", got, expected)
		}
	})

	t.Run("should create response for empty chained statement with ignore errors", func(t *testing.T) {
		params := []string{"id"}
		options := runner.DoneResourceOptions{IgnoreErrors: true}

		expected := domain.DoneResource{
			Details: domain.Details{Status: 400, Success: false, IgnoreErrors: true},
			Result:  "The request was skipped due to missing { :id } param value",
		}

		got := runner.NewEmptyChainedResponse(params, options)

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("NewEmptyChainedResponse = %#+v, want = %#+v", got, expected)
		}
	})
}

func TestGetEmptyChainedParams(t *testing.T) {
	tests := []struct {
		name      string
		statement domain.Statement
		expected  []string
	}{
		{
			"should return nothing if there is no empty chained param",
			domain.Statement{With: map[string]interface{}{"id": "12345"}},
			nil,
		},
		{
			"should return name of empty chained param",
			domain.Statement{With: map[string]interface{}{"id": "12345", "name": runner.EmptyChained}},
			[]string{"name"},
		},
		{
			"should return name of empty chained param inside list",
			domain.Statement{With: map[string]interface{}{"id": "12345", "name": []interface{}{runner.EmptyChained}}},
			[]string{"name"},
		},
		{
			"should return name of empty chained param inside map",
			domain.Statement{With: map[string]interface{}{"id": "12345", "name": map[string]interface{}{"first": runner.EmptyChained}}},
			[]string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runner.GetEmptyChainedParams(tt.statement)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetEmptyChainedParams = %#+v, want = %#+v", got, tt.expected)
			}
		})
	}
}