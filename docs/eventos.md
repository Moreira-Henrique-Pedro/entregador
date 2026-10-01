# Eventos da aplicação

Guia de quais mensagens publicar no Kafka para operar o Entregador: em que tópico, com quais headers e com qual body.

## Como funciona

```
 você publica                    entregador (transporter)                entregador (writer)
┌─────────────────────────────┐        ┌──────────────────────┐        ┌──────────────────────────┐
│ resident-management.events  │──────▶ │ valida e converte em │──────▶ │ delivery-internal.commands│──▶ MongoDB
│ delivery-management.events  │        │ comando interno      │        │ grava / notifica          │──▶ WhatsApp
└─────────────────────────────┘        └──────────────────────┘        └──────────────────────────┘
```

- Você publica só **eventos externos** (`CreateResident`, `CreateDelivery`, …) nos tópicos `*-management.events`.
- A aplicação converte cada evento num **comando interno** (`ProcessCreateResident`, …) no tópico `delivery-internal.commands`. É esse consumer que grava no MongoDB e envia as notificações.
- Os IDs (`resident_id`, `delivery_id`) **são gerados pela aplicação**. Para descobrir um ID, consulte a API HTTP (veja [Consultando o resultado](#consultando-o-resultado)).

## Formato da mensagem

### Headers

| Header      | Obrigatório | Descrição |
|-------------|-------------|-----------|
| `EventType` | **sim**     | Nome do evento, por exemplo `CreateResident`. Sem ele (ou com um valor desconhecido) a mensagem é ignorada sem erro. |
| `Key`       | não         | Chave de negócio, só aparece nos logs. |
| `Source`    | não         | Quem publicou, só aparece nos logs. |

Os nomes dos headers diferenciam maiúsculas de minúsculas. A key da mensagem no Kafka não é usada.

### Body

JSON com o payload dentro de `data`:

```json
{
  "data": { ...campos do evento... }
}
```

> Um objeto sem o envelope `data` também é aceito, mas prefira o formato acima.

---

## Moradores — tópico `resident-management.events`

### Cadastrar morador — `CreateResident`

Headers:

```
EventType: CreateResident
```

Body:

```json
{
  "data": {
    "name": "Ana Souza",
    "apartment": "101",
    "phone": "11999998888"
  }
}
```

| Campo       | Obrigatório | Descrição |
|-------------|-------------|-----------|
| `name`      | sim         | Nome do morador. |
| `apartment` | sim         | Número do apartamento. |
| `phone`     | não         | Telefone para o WhatsApp. Sem código do país, o `NOTIFIER_DEFAULT_COUNTRY_CODE` (padrão `55`) é adicionado. Morador sem telefone não recebe notificação. |

O que acontece:
- O morador é criado com um `resident_id` gerado.
- O **primeiro** morador do apartamento vira `resident-primary`, e os seguintes, `resident-secondary`. O principal é quem recebe as notificações das entregas feitas para o "Outro".
- Se o apartamento ainda não tiver, também é criado o morador **"Outro"** (`resident_id` = `other-<apartamento>`, tipo `other`). Ele recebe as entregas sem destinatário definido. Cada apartamento tem um único "Outro": se ele já existir, não é criado de novo.

Tipos de morador (campo `type` na API):

| `type`               | Quem é |
|----------------------|--------|
| `resident-primary`   | Morador principal. Só existe um por apartamento. |
| `resident-secondary` | Demais moradores do apartamento. |
| `other`              | Destinatário genérico do apartamento, criado automaticamente. |

Se o principal sair do apartamento (por remoção ou mudança de apartamento), o morador mais antigo que restar é promovido a principal.

Status do morador (campo `status`):

| `status`  | Quando |
|-----------|--------|
| `created` | Status inicial de todo morador, inclusive o "Outro". |
| `deleted` | Morador removido por `DeleteResident`. |

### Atualizar morador — `UpdateResident`

Headers:

```
EventType: UpdateResident
```

Body:

```json
{
  "data": {
    "resident_id": "<resident_id>",
    "name": "Ana Souza Lima",
    "apartment": "102",
    "phone": "11977776666"
  }
}
```

| Campo         | Obrigatório | Descrição |
|---------------|-------------|-----------|
| `resident_id` | sim         | ID do morador (consulte na API). |
| `name`        | não         | Novo nome. |
| `apartment`   | não         | Novo apartamento. |
| `phone`       | não         | Novo telefone. |

A atualização é parcial: campo vazio ou omitido não é alterado.
- Se `apartment` mudar:
  - o morador "Outro" do novo apartamento é criado, caso não exista;
  - um principal que muda de apartamento chega ao novo como `resident-secondary`, e o morador mais antigo do apartamento antigo vira principal;
  - se o novo apartamento não tiver moradores, quem chega vira o principal dele.
- `resident_id` inexistente ou do tipo "Outro": o evento é ignorado, com um aviso no log.

### Remover morador — `DeleteResident`

Headers:

```
EventType: DeleteResident
```

Body:

```json
{
  "data": {
    "resident_id": "<resident_id>"
  }
}
```

A remoção é lógica: o morador continua no MongoDB com `status: deleted` e `deleteat` preenchido, mas some das consultas da API. Se o morador removido era o principal, o morador mais antigo que restar no apartamento vira principal.
- `resident_id` inexistente ou do tipo "Outro": o evento é ignorado, com um aviso no log.

---

## Entregas — tópico `delivery-management.events`

### Registrar entrega — `CreateDelivery`

Headers:

```
EventType: CreateDelivery
```

Body:

```json
{
  "data": {
    "apartment": "101",
    "resident_id": "<resident_id>",
    "package_type": "caixa",
    "urgency": "alta"
  }
}
```

| Campo          | Obrigatório | Descrição |
|----------------|-------------|-----------|
| `apartment`    | sim         | Apartamento de destino. Sem ele o evento é descartado. O apartamento precisa ter pelo menos um morador cadastrado (o "Outro" não conta). |
| `resident_id`  | não         | Destinatário. Se vazio, se não existir ou se o morador não for desse apartamento, a entrega vai para o morador "Outro" do apartamento. |
| `package_type` | não         | Tipo do pacote, usado na mensagem. Se vazio, a mensagem usa "encomenda". |
| `urgency`      | não         | Texto livre que entra na mensagem de chegada. |

O que acontece:
1. Se o apartamento não tiver nenhum morador cadastrado, a entrega é **recusada** com o erro `não existe um morador cadastrado para este apartamento`. A mensagem vai direto para a DLQ, sem retry, e nada é gravado.
2. A entrega é criada com status `pending` e um `delivery_id` gerado.
3. É disparada a notificação **delivery_arrived**, conforme o destinatário:
   - morador definido: só ele recebe;
   - morador "Outro": só o **morador principal** (`resident-primary`) recebe. Se ele não tiver telefone, ninguém é notificado.

### Retirar entrega — `DeleteDelivery`

Marca a entrega como retirada.

Headers:

```
EventType: DeleteDelivery
```

Body:

```json
{
  "data": {
    "delivery_id": "<delivery_id>"
  }
}
```

O que acontece:
1. A entrega passa para o status `deleted` (retirada).
2. É disparada a notificação **delivery_picked_up**, com a mesma regra de destinatários do `CreateDelivery`.
- `delivery_id` inexistente: o evento é ignorado, com um aviso no log.

---

## Comandos internos — tópico `delivery-internal.commands`

A própria aplicação publica estes comandos. **Não é preciso publicá-los** no uso normal. Eles estão aqui para ajudar a ler logs e depurar.

| `EventType`             | Publicado por            | Body (`data`) |
|-------------------------|--------------------------|---------------|
| `ProcessCreateResident` | `CreateResident`         | `command_id`, `name`, `apartment`, `phone` |
| `ProcessUpdateResident` | `UpdateResident`         | `command_id`, `resident_id`, `name`, `apartment`, `phone` |
| `ProcessDeleteResident` | `DeleteResident`         | `command_id`, `resident_id` |
| `ProcessCreateDelivery` | `CreateDelivery`         | `command_id`, `apartment`, `resident_id`, `package_type`, `urgency` |
| `ProcessDeleteDelivery` | `DeleteDelivery`         | `command_id`, `delivery_id` |
| `ProcessNotifyDelivery` | writers de criação/retirada de entrega | `command_id`, `delivery_id`, `notification_type` (`delivery_arrived` ou `delivery_picked_up`) |

Sobre o `command_id`:
- É derivado do tópico, partição e offset da mensagem original. Assim, uma mesma mensagem reentregue gera o mesmo comando e não duplica dados.
- Ele vira o `resident_id` / `delivery_id` do registro criado.

## Erros e DLQ

- Se o processamento falhar, a mensagem é reprocessada com backoff exponencial, conforme o bloco `retry` do arquivo em [config/subscriber/deployments/](../config/subscriber/deployments/).
- Esgotadas as tentativas, ou em erro permanente, a mensagem vai para o tópico `delivery-subscriber.dlq` (`DLQ_TOPIC`). Erros permanentes não passam pelo retry. Exemplos: body que não é JSON válido, ou entrega para apartamento sem morador.
- O body da mensagem na DLQ traz o motivo em `data.error` e a mensagem original em `data.raw_payload_string`. O header `OriginalTopic` indica de qual tópico ela veio.

---

## Testando localmente

### 1. Subir tudo

```bash
make up
```

Isso sobe infraestrutura e aplicação:

| Serviço (compose)            | O que é |
|------------------------------|---------|
| `kafka`                      | Kafka, acessível do host em `localhost:9094` |
| `mongodb`                    | MongoDB em `localhost:27017`, sem autenticação |
| `kafka-ui`                   | Kafka UI em http://localhost:8080 |
| `kafka-init-topics`          | cria os tópicos e encerra |
| `resident-events-consumer`   | consome `resident-management.events` |
| `delivery-events-consumer`   | consome `delivery-management.events` |
| `internal-commands-consumer` | consome `delivery-internal.commands` (grava no Mongo e notifica) |
| `api`                        | API HTTP de consulta em http://localhost:8081 |

Os containers da aplicação:
- leem o seu `.env`;
- sobrescrevem só `DELIVERY_BROKER_HOSTS` (`kafka:9092`) e `MONGODB_URI` (`mongodb://mongodb:27017`), porque, de dentro da rede do compose, `localhost` não alcança os outros containers.

Para não depender do Twilio, use `NOTIFIER_PROVIDER=log` no `.env`. A notificação sai só no log.

### 2. Acompanhar os logs

```bash
# todos os serviços da aplicação
docker compose logs -f resident-events-consumer delivery-events-consumer internal-commands-consumer api

# só quem grava e notifica
docker compose logs -f internal-commands-consumer
```

Depois de mudar o código, recompile e suba de novo com `docker compose up --build -d`.

Para rodar um consumer fora do Docker (para depurar, por exemplo):
- pare o container correspondente;
- rode `go run ./cmd/entregador -config=<arquivo em config/subscriber/deployments/>`, com o `.env` apontando para `localhost`.

### 3. Publicar um evento

**Pela Kafka UI:**
1. Abra http://localhost:8080 → *Topics* → escolha o tópico → *Produce Message*.
2. Cole o body em **Value**.
3. Em **Headers**, informe, por exemplo, `{"EventType": "CreateResident"}`.

**Pelo terminal:** use o console producer do container. Header e body ficam na mesma linha, separados por **TAB**:

```bash
docker compose exec -T kafka /opt/kafka/bin/kafka-console-producer.sh \
  --bootstrap-server kafka:9092 \
  --topic resident-management.events \
  --property parse.headers=true <<'EOF'
EventType:CreateResident	{"data":{"name":"Ana Souza","apartment":"101","phone":"11999998888"}}
EOF
```

Mais de um header na mesma mensagem: separe com vírgula, por exemplo `EventType:CreateResident,Source:teste`.

### Consultando o resultado

```bash
# moradores de um apartamento (ou ?phone=11999998888)
curl 'http://localhost:8081/v1/residents?apartment=101'

# entregas de um apartamento (status opcional: pending | deleted)
curl 'http://localhost:8081/v1/deliveries?apartment=101&status=pending'
```

### Roteiro do primeiro teste

1. `CreateResident` em `resident-management.events`.
2. `GET /v1/residents?apartment=101`: aparecem a Ana e o morador "Outro". Anote o `resident_id` da Ana.
3. `CreateDelivery` em `delivery-management.events` com esse `resident_id`. No log do consumer de comandos internos aparece a notificação de chegada.
4. `GET /v1/deliveries?apartment=101&status=pending`: anote o `delivery_id`.
5. `DeleteDelivery` com esse `delivery_id`. No log aparece a notificação de retirada, e a entrega passa a aparecer com `status=deleted`.
