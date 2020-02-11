package webhook

import (
	"context"
	"strconv"
	"testing"

	"github.com/google/go-github/v29/github"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		input interface{}
	}{
		{func(e *github.PullRequestEvent) {}},
		{func(i *Installation, e *github.IssueEvent) {}},
		{func(ctx context.Context, e *github.PullRequestEvent) {}},
		{func(ctx context.Context, i *Installation, e *github.IssueEvent) {}},

		{func(e *github.PullRequestEvent) error { return nil }},
		{func(i *Installation, e *github.IssueEvent) error { return nil }},
		{func(ctx context.Context, e *github.PullRequestEvent) error { return nil }},
		{func(ctx context.Context, i *Installation, e *github.IssueEvent) error { return nil }},
	}

	for i, tt := range tests {
		input := tt.input
		t.Run(t.Name()+":"+strconv.Itoa(i), func(t *testing.T) {
			err := Validate(input)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestValidateErrorCases(t *testing.T) {
	tests := []struct {
		input interface{}
	}{
		{func() {}},
		{func(ctx context.Context) {}},
		{func(ctx context.Context, a, b, c string) {}},
		{func(ctx context.Context, a, b, c string) string { return "" }},
	}

	for i, tt := range tests {
		input := tt.input
		t.Run(t.Name()+":"+strconv.Itoa(i), func(t *testing.T) {
			err := Validate(input)
			if err == nil {
				t.Error("error is nil")
			}
		})
	}
}
