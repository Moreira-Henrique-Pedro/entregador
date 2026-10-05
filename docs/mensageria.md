# Mensageria (notificações assíncronas)

A mensageria é usada só **internamente**, para enviar as notificações de WhatsApp fora do request do front. Moradores e entregas são cadastrados pela [API HTTP](api.md). **Não é preciso publicar nada** no uso normal.

A fila é o **Google Cloud Pub/Sub**, tanto em produção (Cloud Run) quanto localmente (emulador). Há um processo só: a API publica a notificação e também a recebe de volta por push.

## Como funciona

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
| `GCP_PROJECT_ID` | — | Projeto do Google Cloud. Obrigatório. |
| `PUBSUB_NOTIFICATIONS_TOPIC` | `delivery-notifications` | Tópico das notificações. |
| `PUBSUB_PUSH_VERIFY_TOKEN` | `true` | Valida o token OIDC do push. |
| `PUBSUB_PUSH_AUDIENCE` | — | Audience do token. Use a URL do endpoint de push. |
| `PUBSUB_PUSH_SERVICE_ACCOUNT` | — | Service account configurada na push subscription. |
| `PUBSUB_EMULATOR_HOST` | — | Só local: aponta o client para o emulador (ex.: `localhost:8085`). |

As credenciais vêm do *Application Default Credentials*: no Cloud Run, a service account do serviço; localmente, o emulador não exige credencial.

Retry, backoff e dead-letter **não ficam no código**: são configuração da subscription. Veja [deploy-cloud-run.md](deploy-cloud-run.md).

---

## Testando localmente

### Subir o ambiente

```bash
make up
```

| Serviço (compose) | O que é |
|-------------------|---------|
| `mongodb`         | MongoDB em `localhost:27017`, sem autenticação |
| `pubsub`          | Emulador do Pub/Sub em `localhost:8085` |
| `pubsub-init`     | Cria o tópico e a push subscription apontando para a `api`, e encerra |
| `pubsub-ui`       | UI do emulador ([NeoScript/pubsub-emulator-ui](https://github.com/NeoScript/pubsub-emulator-ui)) em http://localhost:7200 |
| `api`             | API HTTP em http://localhost:8081 |

- Localmente, as variáveis vêm do `.env.test` (`NOTIFIER_PROVIDER=log`: a notificação sai só no log, sem Twilio). Para usar outro arquivo: `make up ENV_FILE=.env`.
- O compose força os valores de rede e do emulador (`MONGODB_URI`, `PUBSUB_EMULATOR_HOST`, `GCP_PROJECT_ID`, `PUBSUB_PUSH_VERIFY_TOKEN=false`).
- Para enviar WhatsApp de verdade localmente, descomente o bloco do Twilio no `.env.test`.

### Ver as mensagens na UI

1. Abra http://localhost:7200 e troque o host para `http://localhost:8085` (a UI roda no navegador e chama o emulador direto).
2. Adicione o projeto `entregador-local`.
3. Crie uma **pull** subscription no tópico `delivery-notifications` (ex.: `debug-tap`). Não use a `delivery-notifications-push`: ela entrega para a API.
4. Registre uma entrega e faça o pull na `debug-tap`.

A subscription só recebe o que for publicado depois de ser criada, e some quando o emulador reinicia (ele não guarda estado).

Para rodar a API **fora do Docker** (para depurar, por exemplo):

```bash
make infra   # MongoDB + emulador, com o push apontando para http://host.docker.internal:8081
make api     # use os valores locais do .env.example
```

> Diferenças do emulador: ele não tem backoff nem dead-letter, e reentrega uma mensagem com erro a cada ~1s até ela ser confirmada. Se ficar preso num loop, `make down` limpa tudo, porque o emulador guarda os dados só em memória.

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

Logs: `docker compose logs -f api`.

### Reenviar uma notificação manualmente

Útil para reprocessar algo que caiu no dead-letter. Se a notificação já tiver sido enviada, ela só é pulada.

No emulador (na GCP, use `gcloud pubsub topics publish`):

```bash
curl -X POST localhost:8085/v1/projects/entregador-local/topics/delivery-notifications:publish \
  -H 'Content-Type: application/json' \
  -d "{\"messages\":[{\"attributes\":{\"EventType\":\"NotifyDelivery\"},
       \"data\":\"$(printf '{"delivery_id":"<delivery_id>","notification_type":"delivery_arrived"}' | base64 -w0)\"}]}"
```
