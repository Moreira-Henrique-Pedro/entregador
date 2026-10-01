package pubsub

import (
	"context"
	"encoding/json"
	"testing"

	gcppubsub "cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"cloud.google.com/go/pubsub/v2/pstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/messages"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	pkgPubsub "github.com/Moreira-Henrique-Pedro/entregador/pkg/pubsub"
)

const topic = "projects/test/topics/delivery-notifications"

func newFakeClient(t *testing.T) (*pstest.Server, *gcppubsub.Client) {
	t.Helper()
	ctx := context.Background()

	srv := pstest.NewServer()
	t.Cleanup(func() { _ = srv.Close() })

	conn, err := grpc.NewClient(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	client, err := gcppubsub.NewClient(ctx, "test", option.WithGRPCConn(conn))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	return srv, client
}

func TestNotificationScheduler_Schedule(t *testing.T) {
	srv, client := newFakeClient(t)
	_, err := client.TopicAdminClient.CreateTopic(context.Background(), &pubsubpb.Topic{Name: topic})
	require.NoError(t, err)

	scheduler := NewNotificationScheduler(client, topic)
	defer scheduler.Stop()

	require.NoError(t, scheduler.Schedule(context.Background(), "d1", domain.NotificationTypeDeliveryArrived))
	require.NoError(t, scheduler.Schedule(context.Background(), "d1", domain.NotificationTypeDeliveryArrived))

	published := srv.Messages()
	require.Len(t, published, 2)

	message := published[0]
	assert.Equal(t, messages.NotifyDeliveryType, message.Attributes[pkgPubsub.EventTypeHeader])
	assert.Equal(t, "d1", message.Attributes[pkgPubsub.KeyHeader])

	var command messages.NotifyDelivery
	require.NoError(t, json.Unmarshal(message.Data, &command))
	assert.Equal(t, "d1", command.DeliveryID)
	assert.Equal(t, domain.NotificationTypeDeliveryArrived, command.NotificationType)
	assert.NotEmpty(t, command.CommandID)

	var second messages.NotifyDelivery
	require.NoError(t, json.Unmarshal(published[1].Data, &second))
	assert.Equal(t, command.CommandID, second.CommandID, "same notification, same command id")
}

func TestNotificationScheduler_UnknownTopicFails(t *testing.T) {
	_, client := newFakeClient(t)

	scheduler := NewNotificationScheduler(client, "projects/test/topics/missing")
	defer scheduler.Stop()

	err := scheduler.Schedule(context.Background(), "d1", domain.NotificationTypeDeliveryArrived)

	require.Error(t, err)
}
