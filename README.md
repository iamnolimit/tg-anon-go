# Telegram Anonymous Chat Bot 🎭

Bot Telegram untuk chat anonim menggunakan Golang dan NeonDB (PostgreSQL).

## 📁 Struktur Project

```
tg-anon-go/
├── main.go              # Entry point aplikasi
├── constants/           # Static variables & constants
│   ├── messages.go      # Pesan-pesan bot
│   └── database.go      # Query database
├── databases/           # Database layer
│   ├── db.go           # Inisialisasi koneksi NeonDB
│   └── queries.go      # Fungsi-fungsi query database
├── plugins/            # Plugin system
│   ├── plugin.go       # Interface plugin
│   ├── manager.go      # Plugin manager
│   ├── start.go        # Handler /start & /help
│   └── chat.go         # Handler chat (search, next, stop)
├── .env.example        # Contoh file environment
├── go.mod              # Go modules
└── README.md           # Dokumentasi
```

## 🚀 Fitur

- ✅ Chat anonim dengan partner random
- ✅ Pencarian partner otomatis
- ✅ Skip partner dan cari baru
- ✅ Support berbagai jenis pesan (text, foto, sticker, voice, video, dokumen, GIF)
- ✅ Plugin system yang mudah di-extend
- ✅ Database NeonDB (PostgreSQL)

## 📋 Commands

### User Commands

| Command   | Deskripsi                  |
| --------- | -------------------------- |
| `/start`  | Memulai bot dan registrasi |
| `/search` | Mencari partner chat       |
| `/next`   | Skip partner dan cari baru |
| `/stop`   | Mengakhiri percakapan      |
| `/help`   | Menampilkan bantuan        |
| `/share`  | Bagikan kontak ke partner  |

### Admin Commands

| Command            | Deskripsi                     |
| ------------------ | ----------------------------- |
| `/admin`           | Buka admin panel              |
| `/stats`           | Lihat statistik bot           |
| `/env`             | Lihat environment variables   |
| `/broadcast <msg>` | Broadcast pesan ke semua user |
| `/resetdb`         | Reset database (⚠️ BAHAYA!)   |
| `/addads <msg>`    | Tambah iklan baru             |
| `/delads <id>`     | Hapus iklan                   |
| `/listads`         | Lihat daftar iklan            |
| `/toggleads`       | Enable/Disable iklan          |
| `/ban <user_id>`   | Ban user                      |
| `/unban <user_id>` | Unban user                    |

## 🛠️ Setup

### 1. Clone Repository

```bash
git clone <repository-url>
cd tg-anon-go
```

### 2. Setup NeonDB

1. Buat akun di [Neon](https://neon.tech)
2. Buat project baru
3. Copy connection string

### 3. Konfigurasi Environment

```bash
cp .env.example .env
```

Edit file `.env`:

```env
BOT_TOKEN=your_telegram_bot_token_here
DATABASE_URL=postgres://username:password@ep-xxx.region.aws.neon.tech/dbname?sslmode=require
BOT_DEBUG=false
```

### 4. Install Dependencies

```bash
go mod tidy
```

### 5. Jalankan Bot

```bash
go run main.go
```

## 🔧 Development

### Menambahkan Plugin Baru

1. Buat file baru di folder `plugins/`
2. Implement interface `Plugin`:

```go
package plugins

import (
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MyPlugin struct {
    BasePlugin
}

func NewMyPlugin() *MyPlugin {
    return &MyPlugin{}
}

func (p *MyPlugin) Name() string {
    return "myplugin"
}

func (p *MyPlugin) Commands() []string {
    return []string{"mycommand"}
}

func (p *MyPlugin) HandleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message, command string) error {
    // Handle command logic
    return nil
}
```

3. Register plugin di `main.go`:

```go
pluginManager.Register(plugins.NewMyPlugin())
```

### Menambahkan Query Database

Tambahkan fungsi baru di `databases/queries.go`:

```go
func MyNewQuery(ctx context.Context, param string) error {
    query := `SELECT * FROM table WHERE column = $1`
    _, err := DB.Exec(ctx, query, param)
    return err
}
```

## 📝 Database Schema

### Users Table

- `id` - Primary key
- `telegram_id` - Telegram user ID (unique)
- `username` - Username Telegram
- `first_name` - Nama depan
- `status` - Status user (idle/searching/chatting)
- `partner_id` - ID partner chat
- `created_at` - Tanggal dibuat
- `updated_at` - Tanggal diupdate

### Chat Sessions Table

- `id` - Primary key
- `user1_id` - User 1 Telegram ID
- `user2_id` - User 2 Telegram ID
- `started_at` - Waktu mulai chat
- `ended_at` - Waktu selesai chat
- `is_active` - Status aktif

### Messages Table

- `id` - Primary key
- `session_id` - Foreign key ke chat_sessions
- `sender_id` - Pengirim
- `receiver_id` - Penerima
- `message_type` - Tipe pesan
- `content` - Isi pesan
- `created_at` - Waktu kirim

## 📄 License

MIT License

## 🌐 Deploy ke Heroku

### Metode 1: Deploy via Heroku CLI

#### 1. Install Heroku CLI

Download dan install dari [Heroku CLI](https://devcenter.heroku.com/articles/heroku-cli)

#### 2. Login ke Heroku

```bash
heroku login
```

#### 3. Buat aplikasi Heroku baru

```bash
heroku create nama-app-kamu
```

#### 4. Set environment variables

```bash
heroku config:set BOT_TOKEN=your_telegram_bot_token
heroku config:set DATABASE_URL=your_neondb_connection_string
heroku config:set WEBHOOK_URL=https://nama-app-kamu.herokuapp.com
```

#### 5. Deploy

```bash
git push heroku main
```

#### 6. Pastikan dyno aktif

```bash
heroku ps:scale web=1
```

### Metode 2: Deploy via GitHub

1. Push code ke GitHub repository
2. Buka [Heroku Dashboard](https://dashboard.heroku.com)
3. Klik "New" > "Create new app"
4. Masukkan nama app dan pilih region
5. Di tab "Deploy", pilih "GitHub" sebagai deployment method
6. Connect ke repository GitHub kamu
7. Enable "Automatic Deploys" (opsional)
8. Klik "Deploy Branch"
9. Set environment variables di tab "Settings" > "Config Vars":
   - `BOT_TOKEN` - Token bot dari @BotFather
   - `DATABASE_URL` - Connection string NeonDB
   - `WEBHOOK_URL` - URL Heroku app (https://nama-app.herokuapp.com)

### Environment Variables untuk Heroku

| Variable                | Deskripsi                          | Required | Default                      |
| ----------------------- | ---------------------------------- | -------- | ---------------------------- |
| `BOT_TOKEN`             | Token bot dari BotFather           | ✅       | -                            |
| `DATABASE_URL`          | NeonDB connection string           | ✅       | -                            |
| `WEBHOOK_URL`           | URL app Heroku                     | ✅       | -                            |
| `OWNER_IDS`             | Comma-separated owner Telegram IDs | ✅       | -                            |
| `LOG_GROUP_ID`          | Telegram group ID for media logs   | ✅       | -                            |
| `BOT_DEBUG`             | Mode debug                         | ❌       | `false`                      |
| `BOT_OWNER_URL`         | URL Telegram owner                 | ❌       | `https://t.me/ursweetbae`    |
| `BOT_CHANNEL_URL`       | URL channel bot                    | ❌       | `https://t.me/baecorner`     |
| `BOT_SUPPORT_URL`       | URL support group                  | ❌       | `https://t.me/tgnolimitchat` |
| `MAX_WARNINGS`          | Max warnings before auto-ban       | ❌       | `3`                          |
| `ADS_INTERVAL_MESSAGES` | Send ads every N messages          | ❌       | `30`                         |

### Set Environment Variables via Heroku CLI

```bash
# Required variables
heroku config:set BOT_TOKEN=your_telegram_bot_token -a app-name
heroku config:set DATABASE_URL=your_neondb_connection_string -a app-name
heroku config:set WEBHOOK_URL=https://app-name.herokuapp.com -a app-name
heroku config:set OWNER_IDS=123456789,987654321 -a app-name
heroku config:set LOG_GROUP_ID=-1001234567890 -a app-name

# Optional variables
heroku config:set BOT_DEBUG=false -a app-name
heroku config:set BOT_OWNER_URL=https://t.me/yourusername -a app-name
heroku config:set BOT_CHANNEL_URL=https://t.me/yourchannel -a app-name
heroku config:set BOT_SUPPORT_URL=https://t.me/yoursupport -a app-name
heroku config:set MAX_WARNINGS=3 -a app-name
heroku config:set ADS_INTERVAL_MESSAGES=30 -a app-name
```

### Get Environment Variables

```bash
# Get all config vars
heroku config -a app-name

# Get specific var
heroku config:get BOT_TOKEN -a app-name
```

### Remove Environment Variable

```bash
heroku config:unset VAR_NAME -a app-name
```

### Cek Logs

```bash
heroku logs --tail
```

### Troubleshooting

**Bot tidak merespon:**

1. Cek logs dengan `heroku logs --tail`
2. Pastikan `WEBHOOK_URL` sudah benar
3. Pastikan dyno aktif: `heroku ps`

**Database error:**

1. Pastikan `DATABASE_URL` benar
2. Cek apakah NeonDB bisa diakses dari luar

**Webhook error:**

1. Pastikan HTTPS (bukan HTTP)
2. Pastikan URL tanpa trailing slash

## 🤝 Contributing

Pull requests are welcome!
