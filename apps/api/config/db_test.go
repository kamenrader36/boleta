package config

import (
	"context"
	"testing"
	"time"
)

func TestConnectDB(t *testing.T) {
	// Llamar a ConnectDB ejecuta la inicialización
	ConnectDB()

	if DB == nil {
		t.Fatal("DB no debería ser nil después de ConnectDB()")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := DB.PingContext(ctx); err == nil {
		// Si la base de datos está disponible (ej. en CI con service postgres)
		initSchema()
	}
}
