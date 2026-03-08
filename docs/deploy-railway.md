# Deploy FleetML on Railway

Total time: ~15 minutes. Cost: ~$5-20/month for light usage.

## Prerequisites

- Railway account ([railway.app](https://railway.app))
- GitHub repo with FleetML code pushed
- Dodo Payments account ([dodopayments.com](https://dodopayments.com)) for billing
- S3-compatible storage (AWS S3, Cloudflare R2, or Backblaze B2)

## Step 1: Create Railway Project

1. Go to [railway.app/new](https://railway.app/new)
2. Click **"Empty Project"**
3. Name it `fleetml`

## Step 2: Add PostgreSQL

1. In your project, click **"+ New"** → **"Database"** → **"PostgreSQL"**
2. Railway auto-provisions it. Note the `DATABASE_URL` in the Variables tab.

## Step 3: Deploy Server

1. Click **"+ New"** → **"GitHub Repo"** → select your `fleetML` repo
2. In the service settings:
   - **Source:** Root Directory = `/` (leave as default)
   - **Build:** Custom Dockerfile Path = `server/Dockerfile`
   - **Deploy:** Port = `8080`
3. Add environment variables (Settings → Variables):

| Variable | Value |
|----------|-------|
| `DATABASE_URL` | `${{Postgres.DATABASE_URL}}` (click "Add Reference") |
| `JWT_SECRET` | Generate: `openssl rand -hex 32` |
| `S3_ENDPOINT` | Your S3 endpoint (e.g., `https://s3.amazonaws.com`) |
| `S3_ACCESS_KEY` | Your S3 access key |
| `S3_SECRET_KEY` | Your S3 secret key |
| `S3_BUCKET` | `fleetml-models` |
| `COMPILER_URL` | `http://compiler.railway.internal:8081` |
| `DODO_API_KEY` | From Dodo dashboard |
| `DODO_WEBHOOK_KEY` | From Dodo dashboard |
| `DODO_ENVIRONMENT` | `live` |
| `DODO_STARTER_PRODUCT_ID` | From Dodo (see Step 6) |
| `DODO_PRO_PRODUCT_ID` | From Dodo (see Step 6) |
| `BILLING_SUCCESS_URL` | `https://<dashboard-domain>/dashboard/billing?success=true` |

4. Click **"Generate Domain"** under Settings → Networking
5. Note the generated URL (e.g., `fleetml-server-production.up.railway.app`)

## Step 4: Deploy Dashboard

1. Click **"+ New"** → **"GitHub Repo"** → same repo
2. Settings:
   - **Build:** Custom Dockerfile Path = `dashboard/Dockerfile.railway`
   - **Deploy:** Port = `3000`
3. Add build argument:

| Variable | Value |
|----------|-------|
| `VITE_API_URL` | `https://<server-domain>` (the URL from Step 3) |

4. Click **"Generate Domain"**
5. Note the URL — this is your app URL for customers

## Step 5: Deploy Compiler

1. Click **"+ New"** → **"GitHub Repo"** → same repo
2. Settings:
   - **Build:** Custom Dockerfile Path = `compiler/Dockerfile.mock`
   - **Build:** Root Directory = `/compiler`
   - **Deploy:** Port = `8081`
3. **No public domain needed** — the server accesses it via Railway's private network
4. Service name must be `compiler` (for the internal URL to work)

## Step 6: Set Up Dodo Payments

1. Go to [dodopayments.com](https://dodopayments.com) → Dashboard
2. **Create products:**
   - Product 1: Name = "FleetML Starter", Type = Subscription, Price = $49/month
   - Product 2: Name = "FleetML Pro", Type = Subscription, Price = $199/month
3. Copy the product IDs and paste into Railway server env vars
4. Go to **Developer** → **Webhooks**:
   - URL: `https://<server-domain>/api/v1/webhooks/dodo`
   - Events: All subscription events
5. Copy the webhook signing secret → paste as `DODO_WEBHOOK_KEY` in Railway

## Step 7: Update Server Env Vars

Now that you have the dashboard URL, go back to the server service and update:

| Variable | Value |
|----------|-------|
| `BILLING_SUCCESS_URL` | `https://<dashboard-domain>/dashboard/billing?success=true` |

## Step 8: Set Up S3 Storage

**Option A: AWS S3**
1. Create an S3 bucket named `fleetml-models`
2. Create an IAM user with S3 access
3. Use the access key/secret in Railway env vars

**Option B: Cloudflare R2** (cheaper, no egress fees)
1. Create an R2 bucket in Cloudflare dashboard
2. Create an R2 API token
3. Set `S3_ENDPOINT` to your R2 endpoint URL

## Step 9: Verify

1. Visit `https://<server-domain>/api/v1/health` — should return `{"status":"ok"}`
2. Visit `https://<dashboard-domain>` — should show the landing page
3. Click "Sign Up" → create account → you're in the dashboard

## Step 10: Connect CLI

Tell customers to run:

```bash
fleetml init --cloud
```

Or configure manually:

```bash
mkdir -p ~/.fleetml
cat > ~/.fleetml/config.yaml <<EOF
server:
  address: https://<server-domain>
  api_key: <their-jwt-token>
mode: cloud
EOF
```

## Architecture on Railway

```
┌─────────────────────────────────────────────┐
│  Railway Project                             │
│                                              │
│  ┌──────────┐  ┌──────────┐  ┌───────────┐ │
│  │  Server   │  │Dashboard │  │ Compiler  │ │
│  │  :8080    │  │  :3000   │  │  :8081    │ │
│  │  public   │  │  public  │  │ internal  │ │
│  └────┬──────┘  └──────────┘  └───────────┘ │
│       │                                      │
│  ┌────┴──────┐                               │
│  │PostgreSQL │                               │
│  │  managed  │                               │
│  └───────────┘                               │
└─────────────────────────────────────────────┘
         │
    gRPC + HTTPS
         │
   ┌─────┴─────┐
   │Edge Agents │  (on customer devices)
   └───────────┘
```

## Costs

| Service | Estimated Monthly Cost |
|---------|----------------------|
| Server | $5-10 (usage-based) |
| Dashboard | $2-5 |
| Compiler | $2-5 |
| PostgreSQL | $5-10 |
| **Total** | **~$15-30/month** |

Railway charges per-resource usage (CPU, memory, network). At low traffic the cost is minimal. Scales automatically as usage grows.

## Custom Domain (Optional)

1. In Railway, go to each public service → Settings → Networking
2. Click "Custom Domain"
3. Add your domain (e.g., `api.fleetml.dev`, `app.fleetml.dev`)
4. Add the CNAME record in your DNS provider
5. Railway auto-provisions SSL certificates
