# Deploy no Google Cloud Run

Topologia de produção: **um serviço Cloud Run** (a API) + **Pub/Sub** (notificações) + **MongoDB Atlas** (banco). Não há worker separado. O Pub/Sub entrega as mensagens por push na própria API, então o serviço pode **escalar a zero** quando ninguém está usando.

```
front ──HTTPS──▶ Cloud Run: entregador-api ──▶ MongoDB Atlas
                    │  ▲
          publish   │  │ push (OIDC)
                    ▼  │
              Pub/Sub: delivery-notifications ──(após N falhas)──▶ delivery-notifications-dlq
```

> Os comandos abaixo usam o `gcloud` (instale o [Google Cloud CLI](https://cloud.google.com/sdk/docs/install) e rode `gcloud auth login`). Os nomes de recursos são sugestões.

## 0. Variáveis

```bash
export PROJECT_ID=<seu-projeto>
export REGION=southamerica-east1   # São Paulo
export SERVICE=entregador-api

export API_SA=entregador-api@$PROJECT_ID.iam.gserviceaccount.com
export PUSH_SA=entregador-pubsub-push@$PROJECT_ID.iam.gserviceaccount.com

gcloud config set project $PROJECT_ID
export PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
export PUBSUB_AGENT=service-$PROJECT_NUMBER@gcp-sa-pubsub.iam.gserviceaccount.com
```

## 1. APIs e service accounts

```bash
gcloud services enable run.googleapis.com pubsub.googleapis.com \
  artifactregistry.googleapis.com cloudbuild.googleapis.com secretmanager.googleapis.com

# identidade do serviço (publica no Pub/Sub, lê os secrets)
gcloud iam service-accounts create entregador-api --display-name="Entregador API"

# identidade que o Pub/Sub usa para assinar o push
gcloud iam service-accounts create entregador-pubsub-push --display-name="Pub/Sub push para a Entregador API"

# o Pub/Sub precisa gerar tokens OIDC com essa service account
gcloud iam service-accounts add-iam-policy-binding $PUSH_SA \
  --member=serviceAccount:$PUBSUB_AGENT --role=roles/iam.serviceAccountTokenCreator
```

### Firebase Auth

1. No [console do Firebase](https://console.firebase.google.com), **adicione o Firebase ao projeto GCP** (`$PROJECT_ID`). Não é um projeto novo.
2. Em **Authentication → Sign-in method**, habilite **E-mail/senha**. Não habilite telefone (SMS é cobrado por mensagem).
3. Dê à API permissão para criar usuários e definir papéis (`POST /v1/users`). Validar tokens não precisa de permissão.

```bash
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member=serviceAccount:$API_SA --role=roles/firebaseauth.admin
```

4. Crie o primeiro admin, da sua máquina, com as suas credenciais (`gcloud auth application-default login`) e **sem** o emulador:

```bash
FIREBASE_AUTH_EMULATOR_HOST= GCP_PROJECT_ID=$PROJECT_ID FIREBASE_PROJECT_ID=$PROJECT_ID \
  ENV_FILE=/dev/null make create-admin EMAIL=sindico@condominio.com NAME="Síndico" PASSWORD='<senha forte>'
```

O front usa a configuração web do Firebase (Project settings → Your apps) para fazer login e envia o ID token no header `Authorization: Bearer`.

## 2. MongoDB Atlas

1. Crie um cluster no [MongoDB Atlas](https://www.mongodb.com/atlas). O plano gratuito serve para começar; escolha uma região próxima (São Paulo, se disponível).
2. Crie um usuário de banco com senha forte.
3. Em **Network Access**, libere o acesso. O Cloud Run não tem IP de saída fixo, então as opções são:
   - `0.0.0.0/0` (qualquer IP), protegido só pelo usuário e senha. É o caminho simples e de custo zero;
   - IP fixo via Serverless VPC Access + Cloud NAT. É mais seguro, mas tem custo mensal.
4. Copie a connection string (`mongodb+srv://...`).

## 3. Secrets

```bash
printf '%s' 'mongodb+srv://<user>:<senha>@<cluster>/?retryWrites=true&w=majority' \
  | gcloud secrets create mongodb-uri --data-file=-
printf '%s' '<twilio auth token>' | gcloud secrets create twilio-auth-token --data-file=-

for s in mongodb-uri twilio-auth-token; do
  gcloud secrets add-iam-policy-binding $s \
    --member=serviceAccount:$API_SA --role=roles/secretmanager.secretAccessor
done
```

## 4. Tópicos

```bash
gcloud pubsub topics create delivery-notifications
gcloud pubsub topics create delivery-notifications-dlq

# guarda o que cair no dead-letter, para inspecionar/reprocessar
gcloud pubsub subscriptions create delivery-notifications-dlq-sub \
  --topic=delivery-notifications-dlq --message-retention-duration=7d

# a API publica as notificações
gcloud pubsub topics add-iam-policy-binding delivery-notifications \
  --member=serviceAccount:$API_SA --role=roles/pubsub.publisher

# o Pub/Sub move as mensagens que esgotaram as tentativas para o dead-letter
gcloud pubsub topics add-iam-policy-binding delivery-notifications-dlq \
  --member=serviceAccount:$PUBSUB_AGENT --role=roles/pubsub.publisher
```

## 5. Deploy da API

O `--source .` envia o código para o Cloud Build, que gera a imagem com o `Dockerfile` da raiz (`APP=api` por padrão) e publica no Artifact Registry.

Crie um `env.cloudrun.yaml` (fora do git: ele está no `.gitignore`):

```yaml
ENVIRONMENT: production
APP_NAME: entregador-api
LOG_LEVEL: info
MONGODB_DATABASE: delivery
GCP_PROJECT_ID: <seu-projeto>
PUBSUB_NOTIFICATIONS_TOPIC: delivery-notifications
PUBSUB_PUSH_SERVICE_ACCOUNT: entregador-pubsub-push@<seu-projeto>.iam.gserviceaccount.com
PUBSUB_PUSH_AUDIENCE: pending
FIREBASE_PROJECT_ID: <seu-projeto>
NOTIFIER_PROVIDER: twilio
TWILIO_ACCOUNT_SID: AC...
TWILIO_WHATSAPP_FROM: "+55..."
TWILIO_CONTENT_SID_DELIVERY_ARRIVED: HX...
TWILIO_CONTENT_SID_DELIVERY_PICKED_UP: HX...
```

```bash
gcloud run deploy $SERVICE \
  --source . \
  --region $REGION \
  --service-account $API_SA \
  --allow-unauthenticated \
  --memory 256Mi --cpu 1 \
  --min-instances 0 --max-instances 3 \
  --timeout 60 \
  --env-vars-file env.cloudrun.yaml \
  --set-secrets "MONGODB_URI=mongodb-uri:latest,TWILIO_AUTH_TOKEN=twilio-auth-token:latest"
```

- `PORT` é injetado pelo Cloud Run, e a API o usa no lugar de `HTTP_PORT`.
- `PUBSUB_PUSH_AUDIENCE=pending` é provisório: a URL do serviço só existe depois do primeiro deploy. O próximo passo corrige.

```bash
export URL=$(gcloud run services describe $SERVICE --region $REGION --format='value(status.url)')
export PUSH_ENDPOINT=$URL/internal/pubsub/notifications

gcloud run services update $SERVICE --region $REGION \
  --update-env-vars PUBSUB_PUSH_AUDIENCE=$PUSH_ENDPOINT

curl $URL/health   # 200
```

## 6. Push subscription

Aqui ficam o retry e o dead-letter: no Pub/Sub, eles são configuração da subscription, não código.

```bash
gcloud pubsub subscriptions create delivery-notifications-push \
  --topic=delivery-notifications \
  --push-endpoint=$PUSH_ENDPOINT \
  --push-auth-service-account=$PUSH_SA \
  --push-auth-token-audience=$PUSH_ENDPOINT \
  --ack-deadline=60 \
  --min-retry-delay=10s --max-retry-delay=600s \
  --dead-letter-topic=delivery-notifications-dlq \
  --max-delivery-attempts=10

# o Pub/Sub precisa confirmar (ack) na subscription de origem ao mover para o dead-letter
gcloud pubsub subscriptions add-iam-policy-binding delivery-notifications-push \
  --member=serviceAccount:$PUBSUB_AGENT --role=roles/pubsub.subscriber
```

| Opção | Por quê |
|-------|---------|
| `--ack-deadline=60` | Dá tempo para o Twilio responder (timeout de 10s) e para o cold start. |
| `--min/max-retry-delay` | Backoff exponencial entre 10s e 10min, para não martelar o Twilio quando ele estiver fora. |
| `--max-delivery-attempts=10` | Depois de 10 falhas, a mensagem vai para `delivery-notifications-dlq`. |

## 7. Testar em produção

```bash
curl -X POST $URL/v1/residents -H 'Content-Type: application/json' \
  -d '{"name":"Teste","apartment":"999","phone":"<seu celular>"}'
curl -X POST $URL/v1/deliveries -H 'Content-Type: application/json' \
  -d '{"apartment":"999","package_type":"teste"}'
```

Logs:

```bash
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=$SERVICE" \
  --limit 50 --format='value(jsonPayload.msg, jsonPayload.error)'
```

Procure `Pub/Sub message processed successfully`. Se aparecer `Rejected Pub/Sub push request`, o `PUBSUB_PUSH_AUDIENCE` ou o `PUBSUB_PUSH_SERVICE_ACCOUNT` não bate com a subscription.

O que caiu no dead-letter:

```bash
gcloud pubsub subscriptions pull delivery-notifications-dlq-sub --limit 10 --auto-ack
```

## Novas versões

```bash
gcloud run deploy $SERVICE --source . --region $REGION
```

As variáveis, secrets e a service account são mantidas entre deploys.

## Custos e limites

- **Cloud Run** cobra por uso (CPU e memória enquanto atende requests) e tem franquia gratuita mensal. Com `--min-instances 0`, o custo fica em zero quando não há uso. O preço é um *cold start* de alguns centésimos de segundo, mais a conexão com o Atlas, na primeira requisição depois de um tempo parado. `--min-instances 1` elimina isso, mas passa a cobrar a instância parada.
- **Pub/Sub**: a franquia gratuita mensal cobre com folga o volume de notificações de alguns condomínios.
- **Artifact Registry / Cloud Build**: as imagens e builds de cada deploy têm franquia gratuita. Apague imagens antigas se o armazenamento crescer.
- O maior custo tende a ser o **WhatsApp (Meta + Twilio)**, cobrado por mensagem.

Confira os valores atuais nas páginas de preço do Google Cloud, do MongoDB Atlas e do Twilio.

> ⚠️ **Antes de colocar dados reais:** a API ainda **não tem autenticação**. Com `--allow-unauthenticated`, qualquer pessoa com a URL lista moradores, com nome e telefone. Planeje a autenticação da portaria antes de usar com moradores de verdade (LGPD).
