# Mensageria (Kafka)

O Kafka é usado só **internamente**, para enviar as notificações de forma assíncrona. Moradores e entregas são cadastrados pela [API HTTP](api.md). **Não é preciso publicar nada no Kafka** no uso normal.

## Como funciona

```
 front                     api (cmd/api)                                worker (cmd/worker)
┌───────────┐  HTTP   ┌─────────────────────────────┐   Kafka    ┌────────────────────────────┐
│ POST      │───────▶ │ grava a entrega no MongoDB  │──────────▶ │ NotifyDelivery             │──▶ WhatsApp
│ /v1/deliv │ ◀────── │ agenda NotifyDelivery       │  delivery- │ busca entrega + moradores  │
└───────────┘  201    └─────────────────────────────┘  internal. │ envia e marca notified_at  │
                                                       commands  └────────────────────────────┘
```

- A API grava no MongoDB e responde na hora. Ela só **agenda** a notificação, publicando o comando `NotifyDelivery` no tópico `delivery-internal.commands` (`INTERNAL_COMMANDS_TOPIC`).
- O worker consome esse tópico, decide quem deve receber e envia a notificação pelo `NOTIFIER_PROVIDER` (`twilio` ou `log`).
- A notificação é idempotente. Cada entrega guarda `arrivalnotifiedat` / `pickupnotifiedat`, então um comando reprocessado não manda a mensagem duas vezes.

## Comando `NotifyDelivery` — tópico `delivery-internal.commands`

Headers:

| Header      | Valor |
|-------------|-------|
| `EventType` | `NotifyDelivery`. Mensagem sem ele, ou com outro valor, é ignorada sem erro. |
| `Key`       | `delivery_id`, só aparece nos logs. |
| `Source`    | `APP_NAME` de quem publicou, só aparece nos logs. |

Body:

```json
{
  "data": {
    "command_id": "<uuid>",
    "delivery_id": "<delivery_id>",
    "notification_type": "delivery_arrived"
  }
}
```

| Campo               | Descrição |
|---------------------|-----------|
| `command_id`        | Derivado de `notification_type` + `delivery_id`. A mesma notificação sempre gera o mesmo ID, o que ajuda a rastrear nos logs. |
| `delivery_id`       | Entrega a notificar. Se não existir, o comando é ignorado com um aviso no log. |
| `notification_type` | `delivery_arrived` (publicado no `POST /v1/deliveries`) ou `delivery_picked_up` (publicado no `DELETE /v1/deliveries/{id}`). Outro valor é descartado. |

Quem recebe:
- entrega para um morador definido: só ele;
- entrega para o morador "Outro": só o **morador principal** (`resident-primary`) do apartamento;
- morador sem telefone não recebe. Telefone recusado pelo Twilio é ignorado com um aviso no log.

## Erros e DLQ

- Se o processamento falhar (MongoDB ou Twilio fora, por exemplo), o worker tenta de novo com backoff exponencial, conforme o bloco `retry` de [config/subscriber/deployments/](../config/subscriber/deployments/).
- Esgotadas as tentativas, ou em erro permanente (body que não é JSON válido), a mensagem vai para o tópico `delivery-subscriber.dlq` (`DLQ_TOPIC`).
- O body da mensagem na DLQ traz o motivo em `data.error` e a mensagem original em `data.raw_payload_string`. O header `OriginalTopic` indica de qual tópico ela veio.

---

## Testando localmente

### 1. Subir tudo

```bash
make up
```

| Serviço (compose)   | O que é |
|---------------------|---------|
| `kafka`             | Kafka, acessível do host em `localhost:9094` |
| `mongodb`           | MongoDB em `localhost:27017`, sem autenticação |
| `kafka-ui`          | Kafka UI em http://localhost:8080 |
| `kafka-init-topics` | cria os tópicos e encerra |
| `api`               | API HTTP em http://localhost:8081 |
| `worker`            | consome `delivery-internal.commands` e envia as notificações |

Os containers da aplicação:
- leem o seu `.env`;
- sobrescrevem só `DELIVERY_BROKER_HOSTS` (`kafka:9092`) e `MONGODB_URI` (`mongodb://mongodb:27017`), porque, de dentro da rede do compose, `localhost` não alcança os outros containers.

Para não depender do Twilio, use `NOTIFIER_PROVIDER=log` no `.env`. A notificação sai só no log do worker.

### 2. Acompanhar os logs

```bash
docker compose logs -f api worker
```

Depois de mudar o código, recompile e suba de novo com `docker compose up --build -d`.

Para rodar fora do Docker (para depurar, por exemplo), pare o container correspondente e use `make api` ou `make worker`, com o `.env` apontando para `localhost`.

### 3. Roteiro do primeiro teste

```bash
# 1. cadastre um morador e anote o resident_id
curl -X POST http://localhost:8081/v1/residents \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ana Souza","apartment":"101","phone":"11999998888"}'

# 2. registre uma entrega e anote o delivery_id
#    -> no log do worker aparece a notificação de chegada
curl -X POST http://localhost:8081/v1/deliveries \
  -H 'Content-Type: application/json' \
  -d '{"apartment":"101","resident_id":"<resident_id>","package_type":"caixa"}'

# 3. confira a entrega pendente
curl 'http://localhost:8081/v1/deliveries?apartment=101&status=pending'

# 4. retire a entrega
#    -> no log do worker aparece a notificação de retirada
curl -X DELETE http://localhost:8081/v1/deliveries/<delivery_id>
```

### Reenviar uma notificação manualmente

Útil para reprocessar algo que caiu na DLQ. Header e body ficam na mesma linha, separados por **TAB**:

```bash
docker compose exec -T kafka /opt/kafka/bin/kafka-console-producer.sh \
  --bootstrap-server kafka:9092 \
  --topic delivery-internal.commands \
  --property parse.headers=true <<'EOF'
EventType:NotifyDelivery	{"data":{"delivery_id":"<delivery_id>","notification_type":"delivery_arrived"}}
EOF
```

Se a notificação já tiver sido enviada, o worker só registra no log que a pulou.
