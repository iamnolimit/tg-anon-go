# 🚀 Quick Start Guide

## Minimal Setup (2 langkah!)

### 1. Setup Environment Variables

Buat file `.env`:

```env
BOT_TOKEN=your_bot_token_from_botfather
DATABASE_URL=postgresql://user:password@host/database?sslmode=require
```

**That's it!** Redis URL sudah ada default-nya, tidak perlu set lagi! 🎉

### 2. Run Bot

```bash
go run main.go
```

Bot akan otomatis:

- ✅ Connect ke default Redis instance
- ✅ Initialize database
- ✅ Start matcher
- ✅ Ready untuk matching!

---

## Full Setup (dengan custom config)

### 1. Copy Environment Template

```bash
cp .env.example .env
```

### 2. Edit `.env`

```env
# Required
BOT_TOKEN=7819067921:AAEaYrdKdwvLnPXJ2cMPTGNNBOERGzymBb4
DATABASE_URL=postgresql://neondb_owner:npg_xxx@host/neondb?sslmode=require

# Optional - Bot sudah punya default Redis
# REDIS_URL=redis://default:password@host:port

# Admin
OWNER_IDS=1259894923
LOG_GROUP_ID=-1002339919418

# Settings
USE_POLLING=true
MAX_WARNINGS=3
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run Bot

```bash
go run main.go
```

---

## Expected Output

```
Warning: .env file not found, using environment variables
✅ Connected to database successfully
✅ Database tables created/verified
✅ Connected to Redis successfully
🤖 Bot authorized on account @YourBot
🚀 Redis Matcher started (Random + Nearby + Cleanup)
👂 Listening on channel: search:random
👂 Listening on channel: search:nearby
📦 Registered command: /start from plugin: start
📦 Registered command: /search from plugin: chat
📦 Registered command: /next from plugin: chat
📦 Registered command: /stop from plugin: chat
📦 Registered command: /share from plugin: chat
📦 Registered command: /admin from plugin: admin
✅ Plugin loaded: start
✅ Plugin loaded: chat
✅ Matcher instance set in plugin manager and ChatPlugin
✅ Plugin loaded: admin
✅ All default plugins loaded
🚀 Bot is running in POLLING mode... Press Ctrl+C to stop
```

---

## Test Commands

1. Open Telegram
2. Search for your bot: `@YourBot`
3. Send `/start`
4. Follow registration flow
5. Send `/search` → Choose gender → Choose mode
6. Wait for match! 🎉

---

## Default Configuration

Bot menggunakan default values jika tidak di-set:

| Variable                | Default                                            |
| ----------------------- | -------------------------------------------------- |
| `REDIS_URL`             | `redis://default:5NQH...@redis-12448....com:12448` |
| `BOT_OWNER_URL`         | `https://t.me/ursweetbae`                          |
| `BOT_CHANNEL_URL`       | `https://t.me/baecorner`                           |
| `BOT_SUPPORT_URL`       | `https://t.me/tgnolimitchat`                       |
| `LOG_GROUP_ID`          | `-1002339919418`                                   |
| `OWNER_IDS`             | `1259894923`                                       |
| `MAX_WARNINGS`          | `3`                                                |
| `ADS_INTERVAL_MESSAGES` | `30`                                               |
| `USE_POLLING`           | `true`                                             |

---

## Troubleshooting

### Redis Connection Error

Bot akan otomatis gunakan default Redis. Jika error:

```bash
# Check Redis connection
redis-cli -u redis://default:5NQHBzWOhwHrczAy8SfFtqCCoPcHVTzn@redis-12448.crce194.ap-seast-1-1.ec2.cloud.redislabs.com:12448 ping
```

Expected output: `PONG`

### Database Connection Error

Check format DATABASE_URL:

```
postgresql://user:password@host/database?sslmode=require
```

NeonDB harus include `?sslmode=require` di akhir!

### Bot Not Responding

1. Check BOT_TOKEN valid
2. Check `USE_POLLING=true`
3. Check logs untuk error messages

---

## Production Deployment

### Heroku (Recommended)

1. Click deploy button di README
2. Set `BOT_TOKEN` dan `DATABASE_URL`
3. `REDIS_URL` optional (default sudah ada)
4. Deploy!

### Manual Deployment

```bash
heroku create your-app-name
heroku config:set BOT_TOKEN=your_token
heroku config:set DATABASE_URL=your_neondb_url
heroku config:set USE_POLLING=true
git push heroku main
```

---

## Advanced: Custom Redis

Jika mau gunakan Redis sendiri:

### Option 1: Upstash (Free)

1. Sign up di [Upstash](https://upstash.com)
2. Create Redis database
3. Copy Redis URL
4. Set environment variable:

```bash
export REDIS_URL=redis://default:password@host:port
```

### Option 2: Local Redis

```bash
# Install & start Redis
redis-server

# Set local Redis URL
export REDIS_URL=redis://localhost:6379
```

### Option 3: Render (Free)

1. Sign up di [Render](https://render.com)
2. Create Redis instance
3. Copy Internal Redis URL
4. Set environment variable

---

## Next Steps

✅ Bot ready! Sekarang bisa:

- Test matching dengan 2+ user
- Try gender preference filtering
- Test nearby matching dengan location
- Setup admin commands
- Add custom ads
- Deploy to Heroku

**Happy coding! 🚀**

---

**Need help?** Join [@tgnolimitchat](https://t.me/tgnolimitchat)
