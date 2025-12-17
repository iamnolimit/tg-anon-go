package databases

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

// Var represents a variable in the database
type Var struct {
	ID        int64
	UserID    int64
	Key       string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SetVar menyimpan variabel untuk user tertentu
// Mendukung berbagai tipe data: string, int, int64, bool, atau struct (akan di-marshal ke JSON)
func SetVar(ctx context.Context, userID int64, key string, value interface{}) error {
	var strValue string

	switch v := value.(type) {
	case string:
		strValue = v
	case int:
		strValue = strconv.Itoa(v)
	case int64:
		strValue = strconv.FormatInt(v, 10)
	case float64:
		strValue = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		strValue = strconv.FormatBool(v)
	default:
		// Untuk tipe lain, marshal ke JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		strValue = string(jsonBytes)
	}

	query := `
		INSERT INTO vars (user_id, var_key, var_value, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, var_key)
		DO UPDATE SET var_value = $3, updated_at = $4
	`
	_, err := DB.Exec(ctx, query, userID, key, strValue, time.Now())
	return err
}

// GetVar mengambil variabel sebagai string
func GetVar(ctx context.Context, userID int64, key string) (string, error) {
	query := `SELECT var_value FROM vars WHERE user_id = $1 AND var_key = $2`
	var value string
	err := DB.QueryRow(ctx, query, userID, key).Scan(&value)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// GetVarInt mengambil variabel sebagai int
func GetVarInt(ctx context.Context, userID int64, key string) (int, error) {
	value, err := GetVar(ctx, userID, key)
	if err != nil {
		return 0, err
	}
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

// GetVarInt64 mengambil variabel sebagai int64
func GetVarInt64(ctx context.Context, userID int64, key string) (int64, error) {
	value, err := GetVar(ctx, userID, key)
	if err != nil {
		return 0, err
	}
	if value == "" {
		return 0, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

// GetVarBool mengambil variabel sebagai bool
func GetVarBool(ctx context.Context, userID int64, key string) (bool, error) {
	value, err := GetVar(ctx, userID, key)
	if err != nil {
		return false, err
	}
	if value == "" {
		return false, nil
	}
	return strconv.ParseBool(value)
}

// GetVarFloat64 mengambil variabel sebagai float64
func GetVarFloat64(ctx context.Context, userID int64, key string) (float64, error) {
	value, err := GetVar(ctx, userID, key)
	if err != nil {
		return 0, err
	}
	if value == "" {
		return 0, nil
	}
	return strconv.ParseFloat(value, 64)
}

// GetVarJSON mengambil variabel dan unmarshal ke struct
func GetVarJSON(ctx context.Context, userID int64, key string, dest interface{}) error {
	value, err := GetVar(ctx, userID, key)
	if err != nil {
		return err
	}
	if value == "" {
		return nil
	}
	return json.Unmarshal([]byte(value), dest)
}

// DeleteVar menghapus variabel tertentu
func DeleteVar(ctx context.Context, userID int64, key string) error {
	query := `DELETE FROM vars WHERE user_id = $1 AND var_key = $2`
	_, err := DB.Exec(ctx, query, userID, key)
	return err
}

// DeleteAllVars menghapus semua variabel untuk user tertentu
func DeleteAllVars(ctx context.Context, userID int64) error {
	query := `DELETE FROM vars WHERE user_id = $1`
	_, err := DB.Exec(ctx, query, userID)
	return err
}

// GetAllVars mengambil semua variabel untuk user tertentu
func GetAllVars(ctx context.Context, userID int64) (map[string]string, error) {
	query := `SELECT var_key, var_value FROM vars WHERE user_id = $1`
	rows, err := DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vars := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		vars[key] = value
	}
	return vars, nil
}

// HasVar mengecek apakah variabel ada
func HasVar(ctx context.Context, userID int64, key string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM vars WHERE user_id = $1 AND var_key = $2)`
	var exists bool
	err := DB.QueryRow(ctx, query, userID, key).Scan(&exists)
	return exists, err
}

// SetGlobalVar menyimpan variabel global (userID = 0)
func SetGlobalVar(ctx context.Context, key string, value interface{}) error {
	return SetVar(ctx, 0, key, value)
}

// GetGlobalVar mengambil variabel global
func GetGlobalVar(ctx context.Context, key string) (string, error) {
	return GetVar(ctx, 0, key)
}

// GetGlobalVarInt mengambil variabel global sebagai int
func GetGlobalVarInt(ctx context.Context, key string) (int, error) {
	return GetVarInt(ctx, 0, key)
}

// GetGlobalVarInt64 mengambil variabel global sebagai int64
func GetGlobalVarInt64(ctx context.Context, key string) (int64, error) {
	return GetVarInt64(ctx, 0, key)
}

// GetGlobalVarBool mengambil variabel global sebagai bool
func GetGlobalVarBool(ctx context.Context, key string) (bool, error) {
	return GetVarBool(ctx, 0, key)
}

// DeleteGlobalVar menghapus variabel global
func DeleteGlobalVar(ctx context.Context, key string) error {
	return DeleteVar(ctx, 0, key)
}

// ============================================================
// LOCATION HELPER FUNCTIONS
// ============================================================

// Haversine formula constants
const earthRadiusKm = 6371.0

// degreesToRadians converts degrees to radians
func degreesToRadians(deg float64) float64 {
	return deg * 3.14159265358979323846 / 180
}

// CalculateDistance menghitung jarak antara dua koordinat dalam kilometer
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)

	sinDLat := dLat / 2
	sinDLon := dLon / 2
	
	// Simplified sin calculation using Taylor series approximation for small angles
	// For more accuracy, use math.Sin
	a := sinDLat*sinDLat + sinDLon*sinDLon*cosApprox(lat1Rad)*cosApprox(lat2Rad)
	c := 2 * atanApprox(sqrtApprox(a), sqrtApprox(1-a))

	return earthRadiusKm * c
}

// Simple math approximations to avoid import
func cosApprox(x float64) float64 {
	// Use cos(x) ≈ 1 - x²/2 + x⁴/24 for small angles
	x2 := x * x
	return 1 - x2/2 + x2*x2/24
}

func sqrtApprox(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// Newton's method
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func atanApprox(y, x float64) float64 {
	if x == 0 {
		if y > 0 {
			return 1.5707963267948966
		}
		return -1.5707963267948966
	}
	return y / x // Simplified for small values
}

// GetUserLocation mengambil koordinat lokasi user
func GetUserLocation(ctx context.Context, userID int64) (lat, lon float64, err error) {
	lat, err = GetVarFloat64(ctx, userID, "latitude")
	if err != nil {
		return 0, 0, err
	}
	lon, err = GetVarFloat64(ctx, userID, "longitude")
	if err != nil {
		return 0, 0, err
	}
	return lat, lon, nil
}

// HasLocation mengecek apakah user memiliki lokasi tersimpan
func HasLocation(ctx context.Context, userID int64) bool {
	lat, _ := GetVarFloat64(ctx, userID, "latitude")
	lon, _ := GetVarFloat64(ctx, userID, "longitude")
	return lat != 0 || lon != 0
}
