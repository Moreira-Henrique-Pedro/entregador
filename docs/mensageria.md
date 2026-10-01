# Mensageria (notificações assíncronas)

A mensageria é usada só **internamente**, para enviar as notificações de WhatsApp fora do request do front. Moradores e entregas são cadastrados pela [API HTTP](api.md). **Não é preciso publicar nada** no uso normal.

São dois adapters, escolhidos por `MESSAGING_PROVIDER`:

| `MESSAGING_PROVIDER` | Quando usar | Processos |
|----------------------|-------------|-----------|
| `pubsub` (padrão)    | Produção no Cloud Run e desenvolvimento local | **Um só**: a API publica e também recebe o push do Pub/Sub |
| `kafka`              | Se um dia precisar de Kafka (alto volume, ecossistema Kafka) | Dois: API (`cmd/api`) + worker (`cmd/worker`) |

Os casos de uso são os mesmos nos dois modos. Só muda o adapter que implementa a porta `out.NotificationScheduler` e o adapter que entrega a mensagem ao caso de uso `NotifyDelivery`.

## Pub/Sub (padrão)

```
 front                         api (Cloud Run)                                    Google Pub/Sub
┌───────────┐  POST       ┌──────────────────────────────┐   publish    ┌──────────────────────────────┐
│ portaria  │───────────▶ │ RegisterDelivery             │────────────▶ │ tópico delivery-notifications │
│           │ ◀────────── │  grava no MongoDB            │              │                              │
└───────────┘  201        │  agenda NotifyDelivery       │              │ push subscription            │
                          │                              │  POST (OIDC) │  retry com backoff           │
                          │ /internal/pubsub/            │ ◀─────────── │  dead-letter após N tentativas│
                          │   notifications              │              └──────────────────────────────┘
                          │  NotifyDelivery ──▶ WhatsApp │
                          └──────────────────────────────┘
```

1. A API grava a entrega no MongoDB e publica o comando `NotifyDelivery` no tópico `delivery-notifications` (`PUBSUB_NOTIFICATIONS_TOPIC`). Ela espera o Pub/Sub confirmar o recebimento antes de responder.
2. A **push subscription** faz um `POST` em `/internal/pubsub/notifications`, na própria API.
3. A API valida o token, envia o WhatsApp e responde:

| Resposta | Quando | O que o Pub/Sub faz |
|----------|--------|---------------------|
| `204`    | Notificação enviada, ou já enviada antes (idempotente), ou `EventType` desconhecido | Confirma (ack) a mensagem |
| `500`    | Falha temporária (MongoDB, Twilio) | Tenta de novo com backoff |
| `400`    | Mensagem malformada | Tenta de novo até `max-delivery-attempts` e manda para o **dead-letter topic** |
| `401`    | Token ausente ou inválido | Tenta de novo (indica configuração errada da subscription) |

### Mensagem

Atributos:

| Atributo    | Valor |
|-------------|-------|
| `EventType` | `NotifyDelivery` |
| `Key`       | `delivery_id`, só aparece nos logs |
| `Source`    | `APP_NAME` de quem publicou |

`data` (JSON, sem envelope):

```json
{
  "command_id": "<uuid>",
  "delivery_id": "<delivery_id>",
  "notification_type": "delivery_arrived"
}
```

| Campo               | Descrição |
|---------------------|-----------|
| `command_id`        | Derivado de `notification_type` + `delivery_id`: a mesma notificação sempre gera o mesmo ID. Ajuda a rastrear nos logs. |
| `delivery_id`       | Entrega a notificar. Se não existir, a mensagem é confirmada com um aviso no log. |
| `notification_type` | `delivery_arrived` (publicado no `POST /v1/deliveries`) ou `delivery_picked_up` (publicado no `DELETE /v1/deliveries/{id}`). Outro valor é descartado. |

Quem recebe:
- entrega para um morador definido: só ele;
- entrega para o morador "Outro": só o **morador principal** (`resident-primary`) do apartamento;
- morador sem telefone não recebe. Telefone recusado pelo Twilio é ignorado com um aviso no log.

### Idempotência

O Pub/Sub entrega **pelo menos uma vez**, então a mesma mensagem pode chegar duas vezes. Isso não gera mensagem duplicada: cada entrega guarda `arrivalnotifiedat` / `pickupnotifiedat`, e o `NotifyDelivery` pula a notificação que já foi enviada.

### Segurança do endpoint de push

O `/internal/pubsub/notifications` fica na mesma URL pública da API. Por isso, com `PUBSUB_PUSH_VERIFY_TOKEN=true` (o padrão), toda requisição precisa trazer o token OIDC que o Pub/Sub assina com a service account da subscription. A API confere três coisas:
- a assinatura do Google e a validade do token;
- o `aud`, que deve ser igual a `PUBSUB_PUSH_AUDIENCE`;
- o `email`, que deve ser igual a `PUBSUB_PUSH_SERVICE_ACCOUNT` (e verificado).

Sem `PUBSUB_PUSH_AUDIENCE` e `PUBSUB_PUSH_SERVICE_ACCOUNT`, a API nem sobe. Desligue a verificação (`false`) **só** com o emulador, que não assina o push.

### Configuração

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `MESSAGING_PROVIDER` | `pubsub` | `pubsub` ou `kafka`. |
| `GCP_PROJECT_ID` | — | Projeto do Google Cloud. Obrigatório no modo `pubsub`. |
| `PUBSUB_NOTIFICATIONS_TOPIC` | `delivery-notifications` | Tópico das notificações. |
| `PUBSUB_PUSH_VERIFY_TOKEN` | `true` | Valida o token OIDC do push. |
| `PUBSUB_PUSH_AUDIENCE` | — | Audience do token. Use a URL do endpoint de push. |
| `PUBSUB_PUSH_SERVICE_ACCOUNT` | — | Service account configurada na push subscription. |
| `PUBSUB_EMULATOR_HOST` | — | Só local: aponta o client para o emulador (ex.: `localhost:8085`). |

As credenciais vêm do *Application Default Credentials*: no Cloud Run, a service account do serviço; localmente, o emulador não exige credencial.

Retry, backoff e dead-letter **não ficam no código**: são configuração da subscription. Veja [deploy-cloud-run.md](deploy-cloud-run.md).

## Kafka (alternativo)

Com `MESSAGING_PROVIDER=kafka`, a API publica o mesmo comando `NotifyDelivery` no tópico `delivery-internal.commands` (`INTERNAL_COMMANDS_TOPIC`). O **worker** (`cmd/worker`) o consome. Nesse modo, o próprio código cuida de:
- retry com backoff exponencial, conforme o bloco `retry` de [config/subscriber/deployments/](../config/subscriber/deployments/);
- envio para a DLQ `delivery-subscriber.dlq` (`DLQ_TOPIC`) quando as tentativas se esgotam, ou em erro permanente (body que não é JSON). A mensagem na DLQ traz o motivo em `data.error`, a original em `data.raw_payload_string` e o header `OriginalTopic`.

No Kafka, a mensagem usa headers `EventType`/`Key`/`Source` e o body com envelope `{"data": {...}}`.

---

## Testando localmente

### Modo Pub/Sub (padrão)

```bash
make up
```

| Serviço (compose) | O que é |
|-------------------|---------|
| `mongodb`         | MongoDB em `localhost:27017`, sem autenticação |
| `pubsub`          | Emulador do Pub/Sub em `localhost:8085` |
| `pubsub-init`     | Cria o tópico e a push subscription apontando para a `api`, e encerra |
| `api`             | API HTTP em http://localhost:8081 |

- Os containers leem o seu `.env`, mas o compose força os valores de rede e do emulador (`MONGODB_URI`, `PUBSUB_EMULATOR_HOST`, `GCP_PROJECT_ID`, `PUBSUB_PUSH_VERIFY_TOKEN=false`).
- Para não depender do Twilio, use `NOTIFIER_PROVIDER=log` no `.env`. A notificação sai só no log.

Para rodar a API **fora do Docker** (para depurar, por exemplo):

```bash
make infra   # MongoDB + emulador, com o push apontando para http://host.docker.internal:8081
make api     # use os valores locais do .env.example
```

> Diferenças do emulador: ele não tem backoff nem dead-letter, e reentrega uma mensagem com erro a cada ~1s até ela ser confirmada. Se ficar preso num loop, `make down` limpa tudo, porque o emulador guarda os dados só em memória.

### Modo Kafka

```bash
# com MESSAGING_PROVIDER=kafka no .env
make up-kafka
```

Sobe também `kafka` (`localhost:9094`), `kafka-ui` (http://localhost:8080), `kafka-init-topics` e `worker`.

### Roteiro do primeiro teste

```bash
# 1. cadastre um morador e anote o resident_id
curl -X POST http://localhost:8081/v1/residents \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ana Souza","apartment":"101","phone":"11999998888"}'

# 2. registre uma entrega e anote o delivery_id
#    -> no log aparece "Notification sent" com a mensagem de chegada
curl -X POST http://localhost:8081/v1/deliveries \
  -H 'Content-Type: application/json' \
  -d '{"apartment":"101","resident_id":"<resident_id>","package_type":"caixa"}'

# 3. retire a entrega
#    -> no log aparece a mensagem de retirada
curl -X DELETE http://localhost:8081/v1/deliveries/<delivery_id>
```

Logs: `docker compose logs -f api` (no modo Kafka, `docker compose logs -f api worker`).

### Reenviar uma notificação manualmente

Útil para reprocessar algo que caiu no dead-letter. Se a notificação já tiver sido enviada, ela só é pulada.

**Pub/Sub** (emulador; na GCP, use `gcloud pubsub topics publish`):

```bash
curl -X POST localhost:8085/v1/projects/entregador-local/topics/delivery-notifications:publish \
  -H 'Content-Type: application/json' \
  -d "{\"messages\":[{\"attributes\":{\"EventType\":\"NotifyDelivery\"},
       \"data\":\"$(printf '{"delivery_id":"<delivery_id>","notification_type":"delivery_arrived"}' | base64 -w0)\"}]}"
```

**Kafka** (header e body separados por **TAB**):

```bash
docker compose exec -T kafka /opt/kafka/bin/kafka-console-producer.sh \
  --bootstrap-server kafka:9092 \
  --topic delivery-internal.commands \
  --property parse.headers=true <<'EOF'
EventType:NotifyDelivery	{"data":{"delivery_id":"<delivery_id>","notification_type":"delivery_arrived"}}
EOF
```
