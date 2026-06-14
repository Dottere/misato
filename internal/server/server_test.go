package server

import (
	"database/sql"
	"errors"
	"misato/internal/config"
	"net/http"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// A szerver egységtesztjeinek futtatásához írd be a `go test -v .\internal\server\` parancsot a terminálba

func setupTestServer(t *testing.T) *AppServer {
	cfg := config.Config{
		BindAddress: "127.0.0.1",
		ServerPort:  8080,
		FilesDir:    "test_mangas",
		DBPath:      ":memory:", // :memory: garantálja, hogy a tesztek nem szemetelik tele a vinyót
		DebugMode:   false,
	}

	srv := NewAppServer(cfg)

	t.Cleanup(func() {
		srv.Shutdown()
	})

	return srv
}

// ================== //
// # SERVER TESTING # //
// ================== //

func TestRegisterRoute(t *testing.T) {
	srv := setupTestServer(t)

	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	// Sikeres regisztráció
	err := srv.RegisterRoute("/api/test", mockHandler)
	if err != nil {
		t.Fatalf("expected no error during initial registration, got: %v", err)
	}

	// Duplikált regisztráció (hibás)
	err = srv.RegisterRoute("/api/test", mockHandler)
	if !errors.Is(err, ErrRouteExists) {
		t.Errorf("expected ErrRouteExists, got: %v", err)
	}
}

func TestGetRoutes(t *testing.T) {

	srv := setupTestServer(t)

	// Regisztrálunk két teszt utat
	srv.RegisterRoute("/route1", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r) // do nothing
	}))
	srv.RegisterRoute("/route2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r) // do nothing
	}))

	routes := srv.GetRoutes()

	// A GetRoutes()-nak pontosan 2 utat kell visszaadnia
	if len(routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(routes))
	}

	// Ellenőrizzük, hogy a megfelelő minták vannak-e benne
	found1, found2 := false, false
	for _, r := range routes {
		if r.Pattern == "/route1" {
			found1 = true
		}
		if r.Pattern == "/route2" {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Errorf("GetRoutes did not return the expected route patterns")
	}
}

// ==================== //
// # DATABASE TESTING # //
// ==================== //

func TestUserCreation(t *testing.T) {
	srv := setupTestServer(t)

	username := "John 'sir User Creation' Test"
	password := "supersecretpassword"

	if err := srv.DB.RegisterUser(username, password); err != nil {
		t.Fatalf("User registration failed: %v", err)
	}

	query := `SELECT id, password_hash FROM users WHERE username = ?`

	var id int
	var storedHash string

	err := srv.DB.Conn.QueryRow(query, username).Scan(&id, &storedHash)
	if err != nil {
		t.Fatalf("User wasn't registered properly: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		t.Fatalf("Stored password doesn't match initial password: %v", err)
	}
}

func TestUserLogin(t *testing.T) {
	srv := setupTestServer(t)

	username := "John 'sir Login' Test"
	password := "supersecretpassword"

	if err := srv.DB.RegisterUser(username, password); err != nil {
		t.Fatalf("User registration failed: %v", err)
	}

	id, err := srv.DB.AuthenticateUser(username, password)
	if err != nil {
		t.Fatalf("User authentication failed: %v", err)
	}

	query := `SELECT username FROM users WHERE id = ?`
	var queried_username string
	queryError := srv.DB.Conn.QueryRow(query, id).Scan(&queried_username)
	if queryError != nil {
		t.Fatalf("Query failed: %v", queryError)
	}

	if queried_username != username {
		t.Fatalf("Queried username doesn't match initial username: %s != %s", queried_username, username)
	}
}

func TestUserDeletion(t *testing.T) {
	srv := setupTestServer(t)

	username := "John 'sir Deletion' Test"
	password := "supersecretpassword"

	if err := srv.DB.RegisterUser(username, password); err != nil {
		t.Fatalf("User registration failed: %v", err)
	}

	if err := srv.DB.DeleteUser(username, password); err != nil {
		t.Fatalf("User deletion failed: %v", err)
	}

	query := `SELECT id FROM users WHERE username = ?`

	var id int
	err := srv.DB.Conn.QueryRow(query, username).Scan(&id)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("User deletion didn't work as expected")
	}
}
