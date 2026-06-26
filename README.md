# SmartPost - Telegram Kanallarni Boshqarish Tizimi

B2B SaaS platform for Telegram channel management. Create, edit, and schedule posts with media, inline buttons, and video notes.

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Telegram Bot Token (from @BotFather)

### Setup

1. **Environment faylini sozlash:**
```bash
cp .env.example .env
# .env faylini tahrirlang: BOT_TOKEN ni o'rnating
```

2. **Docker Compose bilan ishga tushirish:**
```bash
docker-compose up --build
```

3. **Botni kanalga admin qilish:**
   - Telegramda botga `/start` yuboring
   - Botni kanalingizga admin sifatida qo'shing
   - Bot avtomatik kanalni bog'laydi

### Development (Lokal)

**Backend:**
```bash
cd backend
go run ./cmd/smartpost
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

## 📁 Project Structure

```
├── backend/               # Go API + Bot
│   ├── cmd/smartpost/     # Entry point
│   └── internal/
│       ├── api/           # REST API (TWA)
│       ├── bot/           # Telegram Bot + FSM
│       ├── config/        # Config loader
│       ├── database/      # PostgreSQL
│       ├── models/        # Data models
│       ├── scheduler/     # Asynq worker
│       └── telegram/      # Sender + Rate limiter
├── frontend/              # SolidJS TWA
│   └── src/
│       ├── api/           # API client
│       ├── components/    # UI components
│       ├── pages/         # Pages
│       └── stores/        # State management
├── nginx/                 # Reverse proxy
├── docker-compose.yml
└── .env.example
```

## 🤖 Bot Commands

| Command | Description |
|---------|------------|
| `/start` | Ro'yxatdan o'tish |
| `/newpost` | Yangi post yaratish |
| `/mychannels` | Kanallar ro'yxati |
| `/cancel` | Jarayonni bekor qilish |
| `/help` | Yordam |

## 🔘 Inline Button Format

```
Tugma matni - https://url.com - green
Ikkinchi tugma - https://url.com - red
```

**Ranglar:** green 🟢, red 🔴, blue 🔵, yellow 🟡, orange 🟠, purple 🟣

## 🏗️ Tech Stack

- **Backend:** Go 1.22+, go-telegram/bot, pgx, chi
- **Database:** PostgreSQL 16
- **Queue:** Redis 7 + Asynq
- **Frontend:** SolidJS + Vite
- **Infrastructure:** Docker, Nginx

## 📝 License

MIT
