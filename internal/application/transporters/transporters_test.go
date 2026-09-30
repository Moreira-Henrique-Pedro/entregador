package transporters

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/events"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
)

const (
	testInternalTopic = "delivery.internal"
	testSourceTopic   = "delivery.source"
)

// transporterCase describes one transporter: how to run it and what it must publish.
type transporterCase struct {
	name      string
	handle    func(ctx context.Context, pub pubsub.MessagePublisher[any]) error
	wantType  string
	wantKey   string
	wantData  any
	commandID func(data any) (string, bool)
	// clearID returns a copy of the command without its CommandID, for field comparison.
	clearID func(data any) any
}

func transporterCases() []transporterCase {
	return []transporterCase{
		{
			name: "create delivery",
			handle: func(ctx context.Context, pub pubsub.MessagePublisher[any]) error {
				return NewCreateDeliveryTransporter(pub, testInternalTopic, testSourceTopic).Handle(ctx, &events.CreateDelivery{
					Apartment: "101", ResidentID: "r1", PackageType: "box", Urgency: "high",
				})
			},
			wantType: commands.ProcessCreateDeliveryCommandType,
			wantKey:  "101",
			wantData: commands.ProcessCreateDeliveryCommand{Apartment: "101", ResidentID: "r1", PackageType: "box", Urgency: "high"},
			commandID: func(data any) (string, bool) {
				c, ok := data.(*commands.ProcessCreateDeliveryCommand)
				if !ok {
					return "", false
				}
				return c.CommandID, true
			},
			clearID: func(data any) any {
				c := *data.(*commands.ProcessCreateDeliveryCommand)
				c.CommandID = ""
				return c
			},
		},
		{
			name: "create resident",
			handle: func(ctx context.Context, pub pubsub.MessagePublisher[any]) error {
				return NewCreateResidentTransporter(pub, testInternalTopic, testSourceTopic).Handle(ctx, &events.CreateResident{
					Name: "Ana", Apartment: "202", Phone: "11999999999",
				})
			},
			wantType: commands.ProcessCreateResidentCommandType,
			wantKey:  "Ana",
			wantData: commands.ProcessCreateResidentCommand{Name: "Ana", Apartment: "202", Phone: "11999999999"},
			commandID: func(data any) (string, bool) {
				c, ok := data.(*commands.ProcessCreateResidentCommand)
				if !ok {
					return "", false
				}
				return c.CommandID, true
			},
			clearID: func(data any) any {
				c := *data.(*commands.ProcessCreateResidentCommand)
				c.CommandID = ""
				return c
			},
		},
		{
			name: "delete delivery",
			handle: func(ctx context.Context, pub pubsub.MessagePublisher[any]) error {
				return NewDeleteDeliveryTransporter(pub, testInternalTopic, testSourceTopic).Handle(ctx, &events.DeleteDelivery{DeliveryID: "d1"})
			},
			wantType: commands.ProcessDeleteDeliveryCommandType,
			wantKey:  "d1",
			wantData: commands.ProcessDeleteDeliveryCommand{DeliveryID: "d1"},
			commandID: func(data any) (string, bool) {
				c, ok := data.(*commands.ProcessDeleteDeliveryCommand)
				if !ok {
					return "", false
				}
				return c.CommandID, true
			},
			clearID: func(data any) any {
				c := *data.(*commands.ProcessDeleteDeliveryCommand)
				c.CommandID = ""
				return c
			},
		},
		{
			name: "delete resident",
			handle: func(ctx context.Context, pub pubsub.MessagePublisher[any]) error {
				return NewDeleteResidentTransporter(pub, testInternalTopic, testSourceTopic).Handle(ctx, &events.DeleteResident{ResidentID: "r1"})
			},
			wantType: commands.ProcessDeleteResidentCommandType,
			wantKey:  "r1",
			wantData: commands.ProcessDeleteResidentCommand{ResidentID: "r1"},
			commandID: func(data any) (string, bool) {
				c, ok := data.(*commands.ProcessDeleteResidentCommand)
				if !ok {
					return "", false
				}
				return c.CommandID, true
			},
			clearID: func(data any) any {
				c := *data.(*commands.ProcessDeleteResidentCommand)
				c.CommandID = ""
				return c
			},
		},
		{
			name: "update resident",
			handle: func(ctx context.Context, pub pubsub.MessagePublisher[any]) error {
				return NewUpdateResidentTransporter(pub, testInternalTopic, testSourceTopic).Handle(ctx, &events.UpdateResident{
					ResidentID: "r1", Name: "Ana", Apartment: "202", Phone: "11999999999",
				})
			},
			wantType: commands.ProcessUpdateResidentCommandType,
			wantKey:  "r1",
			wantData: commands.ProcessUpdateResidentCommand{ResidentID: "r1", Name: "Ana", Apartment: "202", Phone: "11999999999"},
			commandID: func(data any) (string, bool) {
				c, ok := data.(*commands.ProcessUpdateResidentCommand)
				if !ok {
					return "", false
				}
				return c.CommandID, true
			},
			clearID: func(data any) any {
				c := *data.(*commands.ProcessUpdateResidentCommand)
				c.CommandID = ""
				return c
			},
		},
	}
}

// publishedCommandID runs the transporter once and returns the published command id.
func publishedCommandID(t *testing.T, tc transporterCase, ctx context.Context) string {
	t.Helper()
	pub := &fakePublisher{}
	if err := tc.handle(ctx, pub); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if len(pub.calls) != 1 || len(pub.calls[0].messages) != 1 {
		t.Fatalf("expected exactly one published message, got %+v", pub.calls)
	}
	id, ok := tc.commandID(pub.calls[0].messages[0].Payload.Data)
	if !ok {
		t.Fatalf("unexpected payload type %T", pub.calls[0].messages[0].Payload.Data)
	}
	return id
}

func TestTransportersPublishCommand(t *testing.T) {
	for _, tc := range transporterCases() {
		t.Run(tc.name, func(t *testing.T) {
			pub := &fakePublisher{}
			if err := tc.handle(context.Background(), pub); err != nil {
				t.Fatalf("Handle() error = %v", err)
			}

			if len(pub.calls) != 1 {
				t.Fatalf("Publish called %d times, want 1", len(pub.calls))
			}
			call := pub.calls[0]
			if call.topic != testInternalTopic {
				t.Errorf("topic = %q, want %q", call.topic, testInternalTopic)
			}
			if len(call.messages) != 1 {
				t.Fatalf("published %d messages, want 1", len(call.messages))
			}
			msg := call.messages[0]
			if msg.Headers.EventType != tc.wantType {
				t.Errorf("EventType = %q, want %q", msg.Headers.EventType, tc.wantType)
			}
			if msg.Headers.Key != tc.wantKey {
				t.Errorf("Key = %q, want %q", msg.Headers.Key, tc.wantKey)
			}
			if msg.Headers.Source != testSourceTopic {
				t.Errorf("Source = %q, want %q", msg.Headers.Source, testSourceTopic)
			}
			if msg.Headers.OriginalTopic != nil {
				t.Errorf("OriginalTopic = %q, want nil", *msg.Headers.OriginalTopic)
			}

			id, ok := tc.commandID(msg.Payload.Data)
			if !ok {
				t.Fatalf("payload type = %T, want %T", msg.Payload.Data, tc.wantData)
			}
			if id == "" {
				t.Error("CommandID is empty")
			}
			if got := tc.clearID(msg.Payload.Data); !reflect.DeepEqual(got, tc.wantData) {
				t.Errorf("command = %+v, want %+v", got, tc.wantData)
			}
		})
	}
}

func TestTransportersCommandID(t *testing.T) {
	for _, tc := range transporterCases() {
		t.Run(tc.name+"/deterministic with source message id", func(t *testing.T) {
			ctx := pubsub.ContextWithSourceMessageID(context.Background(), "topic-0-42")
			first := publishedCommandID(t, tc, ctx)
			redelivered := publishedCommandID(t, tc, ctx)
			if first != redelivered {
				t.Errorf("redelivery produced a different CommandID: %q != %q", first, redelivered)
			}
			if first != newCommandID(ctx) {
				t.Errorf("CommandID = %q, want %q", first, newCommandID(ctx))
			}
			other := publishedCommandID(t, tc, pubsub.ContextWithSourceMessageID(context.Background(), "topic-0-43"))
			if other == first {
				t.Error("different source messages produced the same CommandID")
			}
		})
		t.Run(tc.name+"/random without source message id", func(t *testing.T) {
			first := publishedCommandID(t, tc, context.Background())
			second := publishedCommandID(t, tc, context.Background())
			if first == "" || second == "" {
				t.Fatal("CommandID is empty")
			}
			if first == second {
				t.Error("CommandID without source message id must not repeat")
			}
		})
	}
}

func TestTransportersPublishError(t *testing.T) {
	errBroker := errors.New("broker down")
	for _, tc := range transporterCases() {
		t.Run(tc.name, func(t *testing.T) {
			pub := &fakePublisher{err: errBroker}
			err := tc.handle(context.Background(), pub)
			if !errors.Is(err, errBroker) {
				t.Fatalf("Handle() error = %v, want wrapping %v", err, errBroker)
			}
			if len(pub.calls) != 1 {
				t.Errorf("Publish called %d times, want 1", len(pub.calls))
			}
		})
	}
}
