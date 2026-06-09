## Overview

Azure App Service is a fully managed PaaS platform for hosting web apps, REST APIs, and Function Apps. It abstracts the OS and infrastructure; you configure runtime, scale, and deployment. Every App Service runs on an App Service Plan (which determines the VM SKU, region, and scaling tier).

## Common Commands

### App Service Plan and Web App

```bash
# Create resource group and App Service Plan
az group create --name my-rg --location eastus

az appservice plan create \
  --name my-plan \
  --resource-group my-rg \
  --sku S1 \
  --is-linux

# Create web app (Linux)
az webapp create \
  --resource-group my-rg \
  --plan my-plan \
  --name my-app \
  --runtime "NODE:20-lts"

# Create web app with container image
az webapp create \
  --resource-group my-rg \
  --plan my-plan \
  --name my-app \
  --deployment-container-image-name mcr.microsoft.com/appsvc/node:20-lts

# List web apps
az webapp list --resource-group my-rg --output table

# Delete web app
az webapp delete --resource-group my-rg --name my-app
```

### Deployment

```bash
# Deploy local ZIP (build output)
az webapp deploy \
  --resource-group my-rg --name my-app \
  --src-path ./dist.zip --type zip

# Deploy from GitHub repo (manual trigger)
az webapp deployment source config \
  --resource-group my-rg --name my-app \
  --repo-url https://github.com/myorg/myrepo \
  --branch main --manual-integration

# Deploy container image
az webapp config container set \
  --resource-group my-rg --name my-app \
  --docker-custom-image-name myregistry.azurecr.io/myapp:latest \
  --docker-registry-server-url https://myregistry.azurecr.io

# View deployment logs
az webapp log deployment show --resource-group my-rg --name my-app

# Restart app
az webapp restart --resource-group my-rg --name my-app
```

### Deployment Slots

Deployment slots are live environments within an App Service app for staging, testing, and safe swaps.

```bash
# Create a staging slot
az webapp deployment slot create \
  --resource-group my-rg --name my-app \
  --slot staging

# Deploy to staging slot
az webapp deploy \
  --resource-group my-rg --name my-app \
  --slot staging \
  --src-path ./dist.zip --type zip

# Preview staging slot URL
# URL pattern: https://my-app-staging.azurewebsites.net

# Swap staging to production
az webapp deployment slot swap \
  --resource-group my-rg --name my-app \
  --slot staging --target-slot production

# List slots
az webapp deployment slot list \
  --resource-group my-rg --name my-app --output table

# Delete staging slot
az webapp deployment slot delete \
  --resource-group my-rg --name my-app --slot staging
```

### Configuration and App Settings

```bash
# Set environment variables (app settings)
az webapp config appsettings set \
  --resource-group my-rg --name my-app \
  --settings NODE_ENV=production DATABASE_URL="postgresql://..."

# Get current app settings
az webapp config appsettings list --resource-group my-rg --name my-app

# Set connection strings
az webapp config connection-string set \
  --resource-group my-rg --name my-app \
  --name MyDB \
  --connection-string "Server=myserver;Database=mydb;" \
  --connection-string-type SQLServer

# Configure always-on (prevents cold starts; requires at least S1)
az webapp config set --resource-group my-rg --name my-app --always-on true

# Set custom domain
az webapp config hostname add \
  --resource-group my-rg --webapp-name my-app \
  --hostname myapp.example.com

# Bind SSL certificate
az webapp config ssl bind \
  --resource-group my-rg --name my-app \
  --certificate-thumbprint <thumbprint> --ssl-type SNI
```

### Autoscale

```bash
# Create autoscale settings for the App Service Plan
az monitor autoscale create \
  --resource-group my-rg \
  --resource my-plan \
  --resource-type Microsoft.Web/serverfarms \
  --name my-autoscale \
  --min-count 1 --max-count 10 --count 2

# Add scale-out rule (CPU > 70% for 5 min)
az monitor autoscale rule create \
  --resource-group my-rg \
  --autoscale-name my-autoscale \
  --scale out 2 \
  --condition "CpuPercentage > 70 avg 5m"

# Add scale-in rule (CPU < 30% for 5 min)
az monitor autoscale rule create \
  --resource-group my-rg \
  --autoscale-name my-autoscale \
  --scale in 1 \
  --condition "CpuPercentage < 30 avg 5m"
```

### Continuous Deployment from GitHub

```bash
# Create GitHub Actions deployment credential
az ad sp create-for-rbac \
  --name my-app-deploy \
  --role contributor \
  --scopes $(az webapp show -g my-rg -n my-app --query id -o tsv) \
  --json-auth

# Add the output JSON as AZURE_CREDENTIALS secret in GitHub
# Then use azure/webapps-deploy action in GitHub Actions workflow
```

### Logs and Diagnostics

```bash
# Enable application logging to filesystem
az webapp log config \
  --resource-group my-rg --name my-app \
  --application-logging filesystem --level information

# Stream live logs
az webapp log tail --resource-group my-rg --name my-app

# Download logs
az webapp log download --resource-group my-rg --name my-app --log-file ./app-logs.zip
```

## Patterns

### Safe Production Deployment with Slots

1. Deploy new code to `staging` slot
2. Run smoke tests against `staging` URL
3. Perform slot swap: `az webapp deployment slot swap`
4. Monitor production for 15-30 min
5. If issues arise, swap back immediately (swap is instant)

### Zero-Config Container Deployment

1. Build and push image to Azure Container Registry
2. Enable managed identity on App Service: `az webapp identity assign`
3. Grant `AcrPull` role on ACR to the managed identity
4. Set container image: `az webapp config container set`
5. Enable continuous deployment webhook for auto-redeploy on new image push

## Gotchas

- **Slot settings vs non-slot settings** — mark settings as "slot settings" to prevent them from being swapped; database connection strings should be slot settings so staging always points to staging DB
- **Cold start on Free/Shared tiers** — apps idle down after inactivity; upgrade to Basic+ and enable `--always-on` for production
- **Kudu SCM endpoint** — `https://my-app.scm.azurewebsites.net` gives access to the Kudu console, file system, and debug console; useful for diagnosing deployment failures
- **Linux vs Windows** — once created, the OS cannot be changed; choose Linux for containers and most modern runtimes, Windows for legacy .NET Framework apps
- **App Service Plan scale-up vs scale-out** — scale-up changes the VM size (requires brief restart); scale-out adds more instances (no downtime); autoscale handles scale-out automatically
