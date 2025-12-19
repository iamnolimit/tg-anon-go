# 🎭 Telegram Anonymous Chat Bot (Golang + Redis + NeonDB)

Bot Telegram untuk anonymous chat dengan fitur realtime matching menggunakan Redis Pub/Sub, database PostgreSQL (NeonDB), dan gender preference filtering.

## ✨ Fitur Lengkap

### 🔐 User Management

- ✅ Registrasi user dengan gender selection (Pria/Wanita/Lainnya)
- ✅ Edit profil (nama, umur, gender, lokasi)
- ✅ View profil dengan statistik chat
- ✅ Ban/Unban user (admin only)

### 🔍 Smart Matching System

- ✅ **Gender Preference**: Pilih gender partner yang dicari (Pria/Wanita/Lainnya/Semua)
- ✅ **Random Mode**: Cari partner secara acak dengan gender filtering
- ✅ **Nearby Mode**: Cari partner terdekat berdasarkan lokasi (dalam radius 50km) + gender filtering
- ✅ **Realtime Matching**: Menggunakan Redis Pub/Sub untuk instant matching
- ✅ **Auto Fallback**: Jika tidak ada partner nearby, otomatis fallback ke random
- ✅ **Gender Compatibility Check**: Otomatis match user dengan preferensi yang compatible

### 💬 Chat Features

- ✅ Anonymous chat dengan berbagai tipe media:
  - Text messages
  - Photos
  - Videos
  - Stickers
  - Voice messages
  - Documents
  - GIFs/Animations
- ✅ `/next` - Skip partner dan cari yang baru
- ✅ `/stop` - Akhiri chat
- ✅ `/share` - Bagikan kontak ke partner

### 🛡️ Moderation System

- ✅ Log group untuk monitoring media
- ✅ Warn system dengan tombol "⚠️ Warn User"
- ✅ Auto-ban setelah 3 warnings
- ✅ Media auto-deleted dari partner saat warning
- ✅ Notification count (1/3, 2/3, etc)

### 📊 Admin Panel

- ✅ `/admin` - Panel admin dengan statistik
- ✅ `/stats` - Total users, active chats, messages
- ✅ `/broadcast` - Kirim pesan ke semua user
- ✅ `/resetdb` - Reset database (dengan konfirmasi)
- ✅ `/ban` & `/unban` - Manage banned users
- ✅ `/env` - Show environment variables dengan Heroku commands
- ✅ Ads system:
  - `/addads` - Tambah iklan
  - `/delads` - Hapus iklan
  - `/listads` - List semua iklan
  - `/toggleads` - Enable/disable ads

### 🚀 Deployment Ready

- ✅ Heroku support dengan `app.json`
- ✅ Polling mode (default, stable)
- ✅ Webhook mode (optional)
- ✅ IPv4 DNS resolution untuk NeonDB
- ✅ Self-ping untuk keep dyno awake
- ✅ Health check endpoint

## 🏗️ Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Telegram  │────▶│  Bot Server  │────▶│  PostgreSQL │
│   Bot API   │     │   (Golang)   │     │   (NeonDB)  │
└─────────────┘     └──────┬───────┘     └─────────────┘
                           │
                           │ Pub/Sub
                           ▼
                    ┌─────────────┐
                    │    Redis    │
                    │  (Matcher)  │
                    └─────────────┘
```

### Matching Flow:

1. User klik `/search` → Pilih gender preference → Pilih mode (Random/Nearby)
2. Bot publish search request ke Redis channel dengan user data (gender, preference, location)
3. Redis Matcher listen channel dan cari partner yang compatible:
   - Check gender compatibility (mutual preference matching)
   - Check availability (status = searching)
   - Check distance (untuk nearby mode)
4. Match found → Connect users di database → Send notifications
5. Start chatting!

## 🛠️ Tech Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL (NeonDB)
- **Cache/Matching**: Redis (Upstash/Render)
- **Bot Framework**: go-telegram-bot-api/v5
- **Deployment**: Heroku (or any PaaS)

## 📋 Prerequisites

1. **Telegram Bot Token** - Dari [@BotFather](https://t.me/BotFather)
2. **PostgreSQL Database** - [NeonDB](https://neon.tech) (Free tier available)
3. **Redis Instance** - [Upstash](https://upstash.com) atau [Render](https://render.com) (Free tier available)
4. **Log Group ID** - Buat Telegram group untuk log media
5. **Heroku Account** (optional, untuk deployment)

## 🚀 Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/yourusername/tg-anon-go.git
cd tg-anon-go
```

### 2. Setup Environment Variables

Buat file `.env`:

```env
# Bot Configuration
BOT_TOKEN=your_telegram_bot_token
DATABASE_URL=postgresql://user:password@host/database?sslmode=require
REDIS_URL=redis://default:password@host:port

# Admin Configuration
OWNER_IDS=1259894923,123456789
LOG_GROUP_ID=-1002339919418

# Bot URLs (optional)
BOT_OWNER_URL=https://t.me/yourusername
BOT_CHANNEL_URL=https://t.me/yourchannel
BOT_SUPPORT_URL=https://t.me/yourgroup

# Settings
USE_POLLING=true
BOT_DEBUG=false
MAX_WARNINGS=3
ADS_INTERVAL_MESSAGES=30

# Heroku (optional)
PORT=8080
WEBHOOK_URL=https://yourapp.herokuapp.com
APP_URL=https://yourapp.herokuapp.com
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run Bot

```bash
go run main.go
```

## 🌐 Deploy to Heroku

### Option 1: One-Click Deploy

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

1. Klik tombol di atas
2. Isi nama app
3. Set environment variables:
   - `BOT_TOKEN` - Token dari BotFather
   - `DATABASE_URL` - URL NeonDB
   - `REDIS_URL` - URL Redis dari Upstash/Render
   - `OWNER_IDS` - Your Telegram user ID
   - `LOG_GROUP_ID` - Group ID untuk logging
4. Deploy!

### Option 2: Manual Deploy

```bash
# Login to Heroku
heroku login

# Create app
heroku create your-app-name

# Add buildpack
heroku buildpacks:set heroku/go

# Set environment variables
heroku config:set BOT_TOKEN=your_token
heroku config:set DATABASE_URL=your_neondb_url
heroku config:set REDIS_URL=your_redis_url
heroku config:set OWNER_IDS=your_telegram_id
heroku config:set LOG_GROUP_ID=your_group_id
heroku config:set USE_POLLING=true

# Deploy
git push heroku main
```

## 🔧 Setup Redis

### Upstash (Recommended - Free)

1. Buat akun di [Upstash](https://upstash.com)
2. Create new Redis database
3. Copy **Redis URL** (format: `redis://default:password@host:port`)
4. Set sebagai `REDIS_URL` environment variable

### Render (Alternative - Free)

1. Buat akun di [Render](https://render.com)
2. Create new Redis instance (Free tier: 25MB)
3. Copy **Internal Redis URL**
4. Set sebagai `REDIS_URL`

### Local Redis (Development)

```bash
# Install Redis
# Windows: Download from https://github.com/microsoftarchive/redis/releases
# Mac: brew install redis
# Linux: sudo apt-get install redis

# Start Redis
redis-server

# Set REDIS_URL
REDIS_URL=redis://localhost:6379
```

## 📱 Bot Commands

### User Commands

- `/start` - Registrasi dan mulai bot
- `/search` - Cari partner chat (dengan gender preference)
- `/next` - Skip partner dan cari yang baru
- `/stop` - Akhiri chat
- `/profile` - Lihat profil kamu
- `/editprofile` - Edit profil
- `/share` - Bagikan kontak ke partner
- `/help` - Bantuan

### Admin Commands

- `/admin` - Panel admin
- `/stats` - Statistik bot
- `/broadcast` - Broadcast message
- `/resetdb` - Reset database
- `/ban <user_id>` - Ban user
- `/unban <user_id>` - Unban user
- `/addads` - Tambah iklan
- `/delads <index>` - Hapus iklan
- `/listads` - List iklan
- `/toggleads` - Toggle iklan on/off
- `/env` - Show environment variables

## 🎯 Gender Matching Logic

Bot menggunakan **mutual preference matching** untuk gender compatibility:

### Contoh Matching:

| User A Gender | User A Looking For | User B Gender | User B Looking For | Match? |
| ------------- | ------------------ | ------------- | ------------------ | ------ |
| Pria          | Wanita             | Wanita        | Pria               | ✅ Yes |
| Pria          | Semua              | Wanita        | Pria               | ✅ Yes |
| Pria          | Wanita             | Wanita        | Semua              | ✅ Yes |
| Pria          | Semua              | Wanita        | Semua              | ✅ Yes |
| Pria          | Pria               | Wanita        | Wanita             | ❌ No  |
| Pria          | Wanita             | Pria          | Wanita             | ❌ No  |

### Rules:

1. Jika kedua user pilih "Semua" → **Match** ✅
2. Jika salah satu pilih "Semua" → Check apakah yang satunya match dengan gender user → **Match/No** ✅/❌
3. Jika keduanya pilih spesifik gender → Harus mutual match (A cari gender B DAN B cari gender A) → **Match/No** ✅/❌

## 📁 Project Structure

```
tg-anon-go/
├── main.go                 # Entry point
├── go.mod                  # Dependencies
├── Procfile               # Heroku process
├── app.json               # Heroku deployment config
├── constants/
│   ├── vars.go           # Variable keys & env helpers
│   ├── messages.go       # Bot messages & commands
│   └── database.go       # SQL schemas
├── databases/
│   ├── db.go             # Database connection
│   ├── vars.go           # SetVar/GetVar system
│   └── queries.go        # Database queries
├── matcher/
│   └── matcher.go        # Redis Pub/Sub matching engine
└── plugins/
    ├── plugin.go         # Plugin interface
    ├── manager.go        # Plugin manager
    ├── start.go          # Registration & profile
    ├── chat.go           # Search & chat logic
    └── admin.go          # Admin commands
```

## 🔍 How Gender Matching Works

### Code Flow:

1. **User clicks `/search`**:

   ```go
   // Show gender preference options
   showGenderPreferenceOptions(bot, chatID)
   // Options: Pria, Wanita, Lainnya, Semua
   ```

2. **User selects gender preference**:

   ```go
   // Save preference
   databases.SetVar(ctx, userID, VarSearchGender, genderValue)
   // Show search mode (Random/Nearby)
   ```

3. **User selects mode → Publish to Redis**:

   ```go
   // Get user's gender and preference
   userGender := databases.GetVar(ctx, userID, VarGender)
   searchGender := databases.GetVar(ctx, userID, VarSearchGender)

   // Publish to Redis
   matcher.PublishSearch(ctx, userID, searchMode, lat, lon)
   ```

4. **Redis Matcher receives search request**:

   ```go
   // Check gender compatibility
   if !isGenderCompatible(user1, user2) {
       continue // Skip this potential match
   }

   // Match found!
   ConnectUsers(user1ID, user2ID)
   ```

5. **Gender Compatibility Check**:
   ```go
   func isGenderCompatible(user1, user2 *SearchRequest) bool {
       // Both want "any" → compatible
       if user1.SearchGender == "any" && user2.SearchGender == "any" {
           return true
       }

       // One wants "any" → check other's preference
       if user1.SearchGender == "any" {
           return user2.SearchGender == "any" || user2.SearchGender == user1.Gender
       }

       // Both specific → must be mutual match
       return user1.SearchGender == user2.Gender &&
              user2.SearchGender == user1.Gender
   }
   ```

## 🐛 Troubleshooting

### Bot tidak menerima update

- Check `BOT_TOKEN` sudah benar
- Pastikan `USE_POLLING=true` untuk mode polling
- Check logs: `heroku logs --tail`

### Matching tidak bekerja

- Check `REDIS_URL` sudah benar dan Redis running
- Test Redis connection: `redis-cli -u $REDIS_URL ping`
- Check matcher logs: Look for "🚀 Redis Matcher started"

### Database error

- Check `DATABASE_URL` format benar
- NeonDB harus include `?sslmode=require`
- Test connection: `psql $DATABASE_URL`

### Gender matching tidak sesuai

- Check user sudah set gender preference saat `/search`
- Default preference adalah "Semua" jika tidak diset
- Check logs untuk "Gender not compatible" messages

## 🔐 Security Notes

1. **Environment Variables**: Jangan commit `.env` ke Git
2. **Bot Token**: Keep your token secret
3. **Database URL**: Use connection pooling
4. **Redis URL**: Use TLS for production
5. **Admin IDs**: Only set trusted user IDs

## 📊 Performance

- **Matching Speed**: < 1 second (Redis Pub/Sub)
- **Database Queries**: Optimized with indexes
- **Concurrent Users**: Supports 100+ simultaneous searches
- **Redis Memory**: ~1MB per 1000 active searches

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create feature branch
3. Commit your changes
4. Push to the branch
5. Create Pull Request

## 📝 License

MIT License - feel free to use for your own projects!

## 👤 Author

Created with ❤️ by [@ursweetbae](https://t.me/ursweetbae)

## 🙏 Credits

- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) - Telegram Bot API wrapper
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver
- [go-redis](https://github.com/redis/go-redis) - Redis client
- [NeonDB](https://neon.tech) - Serverless PostgreSQL
- [Upstash](https://upstash.com) - Serverless Redis

## 📧 Support

Need help? Join our support group: [@tgnolimitchat](https://t.me/tgnolimitchat)

---

⭐ Star this repo if you find it useful!
