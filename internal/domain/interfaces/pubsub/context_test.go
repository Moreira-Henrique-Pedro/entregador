package pubsub

import (
	"context"
	"testing"
)

type otherKey string

func TestSourceMessageIDFromContext(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		wantID string
		wantOK bool
	}{
		{name: "missing", ctx: context.Background(), wantID: "", wantOK: false},
		{name: "empty", ctx: ContextWithSourceMessageID(context.Background(), ""), wantID: "", wantOK: false},
		{name: "present", ctx: ContextWithSourceMessageID(context.Background(), "topic-0-42"), wantID: "topic-0-42", wantOK: true},
		{
			name:   "same string under another key type is ignored",
			ctx:    context.WithValue(context.Background(), otherKey("source_message_id"), "topic-0-42"),
			wantID: "", wantOK: false,
		},
		{
			name:   "innermost value wins",
			ctx:    ContextWithSourceMessageID(ContextWithSourceMessageID(context.Background(), "outer"), "inner"),
			wantID: "inner", wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := SourceMessageIDFromContext(tt.ctx)
			if id != tt.wantID || ok != tt.wantOK {
				t.Errorf("SourceMessageIDFromContext() = (%q, %v), want (%q, %v)", id, ok, tt.wantID, tt.wantOK)
			}
		})
	}
}
