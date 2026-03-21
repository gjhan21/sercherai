# Linux Deployment Guide

This project provides a one-command deployment script for Linux servers:

- script: `scripts/deploy_linux_server.sh`
- templates:
  - `deploy/linux/backend.env.example`
  - `deploy/linux/sercherai-backend.service.template`
  - `deploy/linux/sercherai.nginx.conf.template`

## 1) Prerequisites on server

Install runtime/build tools first:

```bash
sudo apt update
sudo apt install -y git golang-go nodejs npm mysql-client nginx
```

If your Linux distro is not Ubuntu/Debian, install equivalent packages.

## 2) Pull code

```bash
sudo mkdir -p /opt/sercherai
sudo chown -R "$USER":"$USER" /opt/sercherai
git clone https://github.com/gjhan21/sercherai.git /opt/sercherai
cd /opt/sercherai
```

## 3) One-command deployment

Default command (migrate DB + build backend/admin/client + systemd + nginx):

```bash
cd /opt/sercherai
MYSQL_HOST=127.0.0.1 \
MYSQL_PORT=3306 \
MYSQL_USER=sercherai \
MYSQL_PWD='wbdE4xkwew2TaNJL' \
MYSQL_DB=sercherai \
BACKEND_PORT=18080 \
CLIENT_PORT=80 \
ADMIN_PORT=8081 \
SERVICE_USER="$USER" \
./scripts/deploy_linux_server.sh
```

Or rely on built-in defaults (already baked into deploy package):

```bash
cd /opt/sercherai
SERVICE_USER="$USER" ./scripts/deploy_linux_server.sh
```

Optional: seed demo data for first-time verification.

```bash
cd /opt/sercherai
RUN_SEED=true SERVICE_USER="$USER" ./scripts/deploy_linux_server.sh
```

## 3.1) Split mode: database and app deployed separately

For environments where DB migration and app release must be isolated (recommended on BT panel servers), use:

DB only:

```bash
cd /opt/sercherai
MYSQL_HOST=127.0.0.1 \
MYSQL_PORT=3306 \
MYSQL_USER=sercherai \
MYSQL_PWD='wbdE4xkwew2TaNJL' \
MYSQL_DB=sercherai \
SERVICE_USER=www \
SERVICE_GROUP=www \
./scripts/deploy_linux_db.sh
```

App only:

```bash
cd /opt/sercherai
SKIP_NGINX=true \
APP_DIR=/www/server/sercherai \
WWW_DIR=/www/wwwroot/sercherai \
SERVICE_USER=www \
SERVICE_GROUP=www \
./scripts/deploy_linux_app.sh
```

## 4) Edit production secrets

After first deploy, update:

- `/etc/sercherai/backend.env`

At minimum:

- `JWT_SECRET`
- `PAYMENT_SIGNING_SECRET`
- `ATTACHMENT_SIGNING_SECRET`
- `TUSHARE_TOKEN` (if market/news sync is required)

Then restart backend:

```bash
sudo systemctl restart sercherai-backend
```

## 5) Verification

```bash
curl -fsS http://127.0.0.1:18080/healthz
sudo systemctl status sercherai-backend --no-pager
sudo nginx -t
```

Access URLs:

- Client: `http://<server-ip>/`
- Admin: `http://<server-ip>:8081/`

## 6) Upgrade flow

```bash
cd /opt/sercherai
git pull origin main
SERVICE_USER="$USER" ./scripts/deploy_linux_server.sh
```

This will re-run idempotent migrations, rebuild binaries/static assets, and restart services.
