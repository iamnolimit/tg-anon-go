package constants

// Bot Commands
const (
	CmdStart       = "start"
	CmdSearch      = "search"
	CmdNext        = "next"
	CmdStop        = "stop"
	CmdHelp        = "help"
	CmdShare       = "share"
	CmdProfile     = "profile"
	CmdEditProfile = "editprofile"
)

// Admin Commands
const (
	CmdAdmin     = "admin"
	CmdBroadcast = "broadcast"
	CmdResetDB   = "resetdb"
	CmdAddAds    = "addads"
	CmdDelAds    = "delads"
	CmdListAds   = "listads"
	CmdToggleAds = "toggleads"
	CmdStats     = "stats"
	CmdBan       = "ban"
	CmdUnban     = "unban"
	CmdEnv       = "env"
	CmdUpdate    = "update"
)

// User Status
const (
	StatusIdle      = "idle"      // User tidak sedang melakukan apapun
	StatusSearching = "searching" // User sedang mencari partner
	StatusChatting  = "chatting"  // User sedang chatting
)

// Messages
const (
	MsgWelcome = `🎭 *Selamat datang di Anonymous Chat Bot!*

Bot ini memungkinkan kamu untuk chat dengan orang asing secara anonim.

📋 *Perintah yang tersedia:*
/search - Mencari partner chat
/next - Mencari partner baru (skip current)
/stop - Mengakhiri percakapan
/help - Menampilkan bantuan

⚠️ *Peraturan:*
• Dilarang mengirim konten NSFW
• Hormati partner chat kamu
• Jangan spam

Ketik /search untuk mulai mencari partner!`
	MsgHelp = `📋 *Daftar Perintah:*

/search - Mencari partner chat baru
/next - Skip partner dan cari yang baru
/stop - Mengakhiri percakapan saat ini
/profile - Lihat profil kamu
/editprofile - Edit profil kamu
/help - Menampilkan pesan bantuan ini

💡 *Tips:*
• Jadilah ramah dan sopan
• Jika tidak nyaman, gunakan /next atau /stop`

	MsgSearching          = "🔍 Mencari partner chat... Mohon tunggu."
	MsgAlreadySearching   = "⏳ Kamu sudah dalam antrian pencarian. Mohon tunggu."
	MsgAlreadyChatting    = "💬 Kamu sudah terhubung dengan seseorang. Gunakan /stop untuk mengakhiri."
	MsgPartnerFound       = "🎉 Partner ditemukan! Silakan mulai percakapan.\n\nKetik /next untuk skip atau /stop untuk mengakhiri."
	MsgPartnerLeft        = "😔 Partner telah meninggalkan chat.\n\nKetik /search untuk mencari partner baru."
	MsgChatEnded          = "👋 Chat telah diakhiri.\n\nKetik /search untuk mencari partner baru."
	MsgNotChatting        = "❌ Kamu tidak sedang dalam percakapan."
	MsgNotSearching       = "❌ Kamu tidak sedang mencari partner."
	MsgSearchCancelled    = "❎ Pencarian dibatalkan."
	MsgCannotSendToSelf   = "❌ Tidak bisa mengirim pesan ke diri sendiri."
	MsgPartnerDisconnect  = "⚠️ Partner terputus dari chat."
	MsgError              = "❌ Terjadi kesalahan. Silakan coba lagi."
	MsgRegistered         = "✅ Kamu telah terdaftar!"
	MsgNotRegistered      = "❌ Kamu belum terdaftar. Silakan ketik /start untuk mendaftar."
	MsgAutoClosedInactive = "⏰ Chat kamu telah otomatis ditutup karena sudah lebih dari 2 hari.\n\nKetik /search untuk mencari partner baru!"

	// Share Messages
	MsgShareSent     = "✅ Kontak kamu telah dikirim ke partner!"
	MsgShareReceived = `📱 *Partner membagikan kontaknya:*

👤 Nama: *%s*
🆔 Username: @%s
🔗 Link: [Klik untuk chat](tg://user?id=%d)

⚠️ Hati-hati saat berbagi informasi pribadi!`
	MsgShareNoUsername = `📱 *Partner membagikan kontaknya:*

👤 Nama: *%s*
🔗 Link: [Klik untuk chat](tg://user?id=%d)

⚠️ Hati-hati saat berbagi informasi pribadi!`
	MsgShareNotChatting = "❌ Kamu harus sedang dalam chat untuk membagikan kontak."
)

// Registration Messages
const (
	MsgRegWelcome = `🎭 *Selamat datang di Anonymous Chat Bot!*

Sebelum mulai, yuk lengkapi profil kamu dulu! 📝

*Silakan masukkan nama kamu:*`
	MsgRegAskAge = `👤 Hai *%s*! Nama yang bagus!

*Sekarang masukkan umur kamu:*
(Contoh: 20)`

	MsgRegAskGender = `📅 Umur kamu *%s tahun* ya!

*Pilih jenis kelamin kamu:*`

	MsgRegAskLocation = `✅ Gender tersimpan!

*Terakhir, bagikan lokasi kamu:*
📍 Klik tombol di bawah untuk share lokasi.

💡 Lokasi digunakan untuk fitur "Cari Partner Terdekat"`

	MsgRegComplete = `🎉 *Registrasi Selesai!*

📋 *Profil kamu:*
👤 Nama: *%s*
📅 Umur: *%s tahun*
👥 Gender: *%s*
📍 Lokasi: *%s*

Sekarang kamu bisa mulai mencari partner chat!
Ketik /search untuk memulai.`
	MsgRegInvalidAge      = "⚠️ Umur tidak valid. Silakan masukkan angka antara 13-100."
	MsgRegInvalidGender   = "⚠️ Pilihan tidak valid. Silakan pilih gender menggunakan tombol."
	MsgRegInvalidLocation = "⚠️ Silakan kirim lokasi menggunakan tombol di bawah atau fitur 📎 Attachment > Location di Telegram."

	MsgProfileInfo = `📋 *Profil Kamu:*

👤 Nama: *%s*
📅 Umur: *%s tahun*
👥 Gender: *%s*
📍 Lokasi: *%s*
📊 Total Chat: *%d*
💬 Total Pesan: *%d*

Gunakan /search untuk mencari partner!
Gunakan /editprofile untuk edit profil.`

	MsgEditProfile = `✏️ *Edit Profil*

Pilih data yang ingin kamu ubah:`

	MsgEditName = `✏️ *Edit Nama*

Nama saat ini: *%s*

Silakan kirim nama baru kamu:`

	MsgEditAge = `✏️ *Edit Umur*

Umur saat ini: *%s tahun*

Silakan kirim umur baru kamu (13-100):`

	MsgEditGender = `✏️ *Edit Gender*

Gender saat ini: *%s*

Pilih gender baru:`

	MsgEditLocation = `✏️ *Edit Lokasi*

Lokasi saat ini: *%s*

Bagikan lokasi baru kamu:`

	MsgProfileUpdated = "✅ Profil berhasil diupdate!"
	MsgEditCancelled  = "❌ Edit profil dibatalkan."
)

// Search Messages
const (
	MsgSearchNearbyNoLocation = "⚠️ Kamu belum menyimpan lokasi. Silakan update lokasi dengan /updatelocation"
	MsgSearchNearbySearching  = "🔍 Mencari partner di sekitar lokasi kamu... Mohon tunggu."
	MsgSearchNearbyNotFound   = "😔 Tidak ada partner terdekat yang tersedia saat ini.\n\nMencari secara random..."
	MsgPartnerDistance        = "🎉 Partner ditemukan! (📍 Jarak: *%.1f km*)\n\nSilakan mulai percakapan.\nKetik /next untuk skip atau /stop untuk mengakhiri."
)

// Admin Messages
const (
	MsgAdminPanel = `🔐 *Admin Panel*

📊 *Statistik:*
👥 Total Users: *%d*
💬 Active Chats: *%d*
📨 Total Messages: *%d*

🛠 *Commands:*
/stats - Lihat statistik
/env - Lihat environment variables
/broadcast <pesan> - Broadcast ke semua user
/update - Update bot ke versi terbaru
/resetdb - Reset database (⚠️ BAHAYA!)
/addads <pesan> - Tambah ads baru
/delads <id> - Hapus ads
/listads - Lihat daftar ads
/toggleads - Enable/Disable ads
/ban <user_id> - Ban user
/unban <user_id> - Unban user`

	MsgAdminOnly      = "❌ Command ini hanya untuk admin."
	MsgBroadcastStart = "📢 Memulai broadcast ke %d users..."
	MsgBroadcastDone  = "✅ Broadcast selesai!\n\n📊 Sukses: %d\n❌ Gagal: %d"
	MsgResetDBConfirm = "⚠️ *PERINGATAN!*\n\nApakah kamu yakin ingin reset database?\nSemua data akan DIHAPUS PERMANEN!\n\nKetik /confirmreset untuk konfirmasi."
	MsgResetDBSuccess = "✅ Database berhasil direset!"
	MsgAdsAdded       = "✅ Ads berhasil ditambahkan dengan ID: %d"
	MsgAdsDeleted     = "✅ Ads dengan ID %d berhasil dihapus."
	MsgAdsNotFound    = "❌ Ads dengan ID %d tidak ditemukan."
	MsgAdsToggled     = "✅ Ads sekarang: *%s*"
	MsgAdsList        = "📋 *Daftar Ads:*\n\n%s"
	MsgAdsEmpty       = "📋 Tidak ada ads yang tersedia."
	MsgUserBanned     = "✅ User %d berhasil dibanned."
	MsgUserUnbanned   = "✅ User %d berhasil diunban."
	MsgInvalidUserID  = "❌ User ID tidak valid."
	MsgStatsInfo      = `📊 *Statistik Bot*

👥 Total Users: *%d*
💬 Active Chats: *%d*
🔍 Searching: *%d*
📨 Total Messages: *%d*
📢 Ads Enabled: *%s*
📝 Total Ads: *%d*`

	MsgEnvInfo = `⚙️ *Environment Variables*

🔗 *Bot URLs:*
• Owner: %s
• Channel: %s
• Support: %s

📋 *Settings:*
• Log Group ID: ` + "`%d`" + `
• Owner IDs: ` + "`%s`" + `
• Max Warnings: ` + "`%d`" + `
• Ads Interval: ` + "`%d`" + ` messages

💡 *Heroku Commands:*
` + "```" + `
# Set env variable
heroku config:set VAR_NAME=value -a app-name

# Get all env variables
heroku config -a app-name

# Examples:
heroku config:set MAX_WARNINGS=5 -a app-name
heroku config:set LOG_GROUP_ID=-100123456 -a app-name
heroku config:set OWNER_IDS=123,456,789 -a app-name
` + "```"

	MsgUpdateStart    = "🔄 *Memulai update bot...*"
	MsgUpdatePulling  = "📥 Pulling latest code dari git..."
	MsgUpdateBuilding = "🔨 Building binary baru..."
	MsgUpdateSuccess  = "✅ Update berhasil! Bot akan restart dalam 3 detik..."
	MsgUpdateFailed   = "❌ Update gagal: %s"
)

// Ads Format
const (
	MsgAdsPrefix = "📢 *Sponsor:*\n\n"
)

// Log Group Messages
const (
	MsgLogMedia = `📸 *Media Log*

👤 From: [User %d](tg://user?id=%d)
🆔 Partner: [User %d](tg://user?id=%d)
📝 Type: *%s*
⏰ Time: %s`

	MsgWarnSuccess  = "⚠️ User %d telah diberi peringatan (%d/%d)"
	MsgWarnAutoBan  = "🚫 User %d telah dibanned otomatis setelah %d peringatan!"
	MsgWarnedNotify = "⚠️ *PERINGATAN!*\n\nKamu mendapat peringatan dari admin karena mengirim konten tidak pantas.\n\n⚠️ Warning: *%d/%d*\n\nJika mencapai %d warning, kamu akan dibanned otomatis!"
	MsgWarnedBanned = "🚫 *KAMU TELAH DIBANNED!*\n\nKamu telah menerima %d peringatan karena mengirim konten tidak pantas dan sekarang dibanned dari bot ini."
	MsgMediaDeleted = "🗑️ Media dari user yang mendapat peringatan telah dihapus."
)

// Callback Prefixes
const (
	CallbackWarnUser = "warn_user_" // Format: warn_user_{userID}_{messageID}
)
