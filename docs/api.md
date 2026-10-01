# API HTTP

A API roda em http://localhost:8081 (`HTTP_PORT`; no Cloud Run, a porta vem de `PORT`). Todas as respostas são JSON. Em caso de erro, o body é `{"error": "<motivo>"}`.

## Moradores

O cadastro de moradores é síncrono: a resposta só volta depois que o MongoDB foi atualizado.

### Tipos e status

| `type`               | Quem é |
|----------------------|--------|
| `resident-primary`   | Morador principal. Só existe um por apartamento. Recebe as notificações das entregas feitas para o "Outro". |
| `resident-secondary` | Demais moradores do apartamento. |
| `other`              | Destinatário genérico do apartamento (`resident_id` = `other-<apartamento>`), criado automaticamente. Não pode ser alterado nem removido. |

| `status`  | Quando |
|-----------|--------|
| `created` | Status inicial de todo morador, inclusive o "Outro". |
| `deleted` | Morador removido. Não aparece mais nas consultas. |

Se o principal sair do apartamento (por remoção ou mudança de apartamento), o morador mais antigo que restar é promovido a principal.

### Formato do morador

```json
{
  "resident_id": "6b1f0c1e-2c4f-4a3e-9d0b-6a2f7e8c9d10",
  "name": "Ana Souza",
  "apartment": "101",
  "phone": "11999998888",
  "type": "resident-primary",
  "status": "created",
  "created_at": "2026-10-01T10:00:00Z",
  "updated_at": "2026-10-01T10:00:00Z"
}
```

### Cadastrar — `POST /v1/residents`

```bash
curl -X POST http://localhost:8081/v1/residents \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ana Souza","apartment":"101","phone":"11999998888"}'
```

| Campo       | Obrigatório | Descrição |
|-------------|-------------|-----------|
| `name`      | sim         | Nome do morador. |
| `apartment` | sim         | Número do apartamento. |
| `phone`     | não         | Telefone para o WhatsApp. Sem código do país, o `NOTIFIER_DEFAULT_COUNTRY_CODE` (padrão `55`) é adicionado. Morador sem telefone não recebe notificação. |

- O `resident_id` é gerado pela aplicação.
- O **primeiro** morador do apartamento vira `resident-primary`, e os seguintes, `resident-secondary`.
- Se o apartamento ainda não tiver, também é criado o morador **"Outro"**.

| Status | Quando |
|--------|--------|
| `201`  | Criado. O body é o morador. |
| `400`  | JSON inválido, campo desconhecido ou campo obrigatório ausente. |

### Consultar por apartamento — `GET /v1/residents?apartment=<apartamento>`

```bash
curl 'http://localhost:8081/v1/residents?apartment=101'
```

Retorna a lista de moradores ativos do apartamento, incluindo o "Outro". Lista vazia se não houver nenhum. Também aceita `?phone=<telefone>` no lugar de `apartment`.

| Status | Quando |
|--------|--------|
| `200`  | Lista de moradores. |
| `400`  | Sem `apartment` nem `phone`, ou com os dois. |

### Atualizar — `PATCH /v1/residents/{resident_id}`

```bash
curl -X PATCH http://localhost:8081/v1/residents/<resident_id> \
  -H 'Content-Type: application/json' \
  -d '{"apartment":"102"}'
```

Body com qualquer combinação de `name`, `apartment` e `phone`. A atualização é parcial: campo vazio ou omitido não é alterado.

Se `apartment` mudar:
- o morador "Outro" do novo apartamento é criado, caso não exista;
- um principal que muda de apartamento chega ao novo como `resident-secondary`, e o morador mais antigo do apartamento antigo vira principal;
- se o novo apartamento não tiver moradores, quem chega vira o principal dele.

| Status | Quando |
|--------|--------|
| `200`  | Atualizado. O body é o morador. |
| `400`  | JSON inválido, campo desconhecido ou nenhum campo informado. |
| `404`  | Morador não existe (ou já foi removido). |
| `409`  | Morador do tipo "Outro". |

### Remover — `DELETE /v1/residents/{resident_id}`

```bash
curl -X DELETE http://localhost:8081/v1/residents/<resident_id>
```

A remoção é lógica: o morador continua no MongoDB com `status: deleted` e `deleteat` preenchido, mas some das consultas. Se ele era o principal, o morador mais antigo que restar no apartamento vira principal.

| Status | Quando |
|--------|--------|
| `204`  | Removido. Sem body. |
| `404`  | Morador não existe (ou já foi removido). |
| `409`  | Morador do tipo "Outro". |

## Entregas

O registro e a retirada são **síncronos**: a resposta só volta depois que o MongoDB foi atualizado. A **notificação** por WhatsApp é **assíncrona**: a API publica um comando na fila (Pub/Sub, ou Kafka no modo alternativo) e a mensagem é enviada logo depois, fora do request. Veja [mensageria.md](mensageria.md). Por isso a resposta não espera o Twilio, e uma falha nele não afeta o cadastro.

### Formato da entrega

```json
{
  "delivery_id": "0b6c3c1e-8f2a-4d4e-9b7a-1c2d3e4f5a6b",
  "apartment": "101",
  "resident_id": "6b1f0c1e-2c4f-4a3e-9d0b-6a2f7e8c9d10",
  "package_type": "caixa",
  "urgency": "alta",
  "status": "pending",
  "created_at": "2026-10-01T10:00:00Z",
  "updated_at": "2026-10-01T10:00:00Z"
}
```

`deleted_at` só aparece em entregas retiradas (`status: deleted`).

### Registrar — `POST /v1/deliveries`

```bash
curl -X POST http://localhost:8081/v1/deliveries \
  -H 'Content-Type: application/json' \
  -d '{"apartment":"101","resident_id":"<resident_id>","package_type":"caixa","urgency":"alta"}'
```

| Campo          | Obrigatório | Descrição |
|----------------|-------------|-----------|
| `apartment`    | sim         | Apartamento de destino. Precisa ter pelo menos um morador cadastrado (o "Outro" não conta). |
| `resident_id`  | não         | Destinatário. Se vazio, se não existir ou se o morador não for desse apartamento, a entrega vai para o morador "Outro". |
| `package_type` | não         | Tipo do pacote, usado na mensagem. Se vazio, a mensagem usa "encomenda". |
| `urgency`      | não         | Texto livre que entra na mensagem de chegada. |

O que acontece:
1. A entrega é gravada com status `pending` e um `delivery_id` gerado pela aplicação.
2. É agendada a notificação **delivery_arrived**:
   - morador definido: só ele recebe;
   - morador "Outro": só o **morador principal** (`resident-primary`) recebe. Se ele não tiver telefone, ninguém é notificado.

> Se a fila (Pub/Sub ou Kafka) estiver fora do ar, a entrega **é registrada mesmo assim** (`201`), mas a notificação de chegada não é enviada. O erro aparece no log da API (`Failed to schedule arrival notification`). Responder erro aqui faria o front tentar de novo e duplicar a entrega.

| Status | Quando |
|--------|--------|
| `201`  | Registrada. O body é a entrega. |
| `400`  | JSON inválido, campo desconhecido ou sem `apartment`. |
| `422`  | O apartamento não tem nenhum morador cadastrado. |

### Retirar — `DELETE /v1/deliveries/{delivery_id}`

```bash
curl -X DELETE http://localhost:8081/v1/deliveries/<delivery_id>
```

Marca a entrega como retirada (`status: deleted`) e agenda a notificação **delivery_picked_up**, com a mesma regra de destinatários do registro.

A chamada é **idempotente**: retirar de novo uma entrega já retirada responde `204`. Se a notificação de retirada ainda não tiver sido enviada, ela é agendada outra vez. Então, se der `500` (por exemplo, com a fila fora do ar), é seguro tentar de novo.

| Status | Quando |
|--------|--------|
| `204`  | Retirada. Sem body. |
| `404`  | Entrega não existe. |
| `500`  | Falha ao gravar ou ao agendar a notificação. Pode tentar de novo. |

### Consultar por apartamento — `GET /v1/deliveries?apartment=<apartamento>`

```bash
curl 'http://localhost:8081/v1/deliveries?apartment=101&status=pending'
```

`status` é opcional: `pending` ou `deleted` (retirada).

| Status | Quando |
|--------|--------|
| `200`  | Lista de entregas. |
| `400`  | Sem `apartment`, ou `status` inválido. |

## Rota interna

`POST /internal/pubsub/notifications` recebe o push da subscription do Pub/Sub (só existe com `MESSAGING_PROVIDER=pubsub`). **Não é para o front**: ela exige o token OIDC assinado pelo Google. Veja [mensageria.md](mensageria.md).
