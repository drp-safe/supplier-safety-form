package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/obeyler/supplier-safety-form/internal/auth"
	"github.com/obeyler/supplier-safety-form/internal/crypto"
	"github.com/obeyler/supplier-safety-form/internal/db"
	"github.com/obeyler/supplier-safety-form/internal/handlers"
	"github.com/obeyler/supplier-safety-form/internal/middleware"
	"github.com/obeyler/supplier-safety-form/internal/services"
	"github.com/obeyler/supplier-safety-form/migrations"
)

func main() {
	_ = godotenv.Load()

	if err := crypto.Init(os.Getenv("ENCRYPTION_KEY")); err != nil {
		log.Fatalf("encryption init failed: %v", err)
	}

	database, err := db.Connect()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database, migrations.FS); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	userService := services.NewUserService(database)
	orgService := services.NewOrganizationService(database)
	sectorService := services.NewSectorService(database)
	themeService := services.NewThemeService(database)
	adminService := services.NewAdminService(database)
	registrationService := services.NewRegistrationService(userService, orgService, sectorService)

	registrationHandler := handlers.NewRegistrationHandler(registrationService)
	meHandler := handlers.NewMeHandler(orgService)
	referenceHandler := handlers.NewReferenceHandler(sectorService, themeService)
	adminHandler := handlers.NewAdminHandler(adminService, sectorService)
	migrationsHandler := handlers.NewMigrationsHandler(database)

	verifier, err := auth.NewKeycloakVerifier(auth.KeycloakConfig{
		IssuerURL:   os.Getenv("KEYCLOAK_ISSUER_URL"),
		ClientID:    os.Getenv("KEYCLOAK_CLIENT_ID"),
		InternalURL: os.Getenv("KEYCLOAK_INTERNAL_URL"),
	})
	if err != nil {
		log.Fatalf("keycloak verifier init failed: %v", err)
	}

	r := mux.NewRouter()

	// Health check (public, unauthenticated).
	r.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`)) //nolint:errcheck
	}).Methods(http.MethodGet)

	// Public routes: self-registration + reference lists needed before login.
	pub := r.PathPrefix("/api").Subrouter()
	pub.HandleFunc("/register", registrationHandler.Register).Methods(http.MethodPost)
	pub.HandleFunc("/sectors", referenceHandler.ListSectors).Methods(http.MethodGet)
	pub.HandleFunc("/themes", referenceHandler.ListThemes).Methods(http.MethodGet)

	// Authenticated API.
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.KeycloakAuth(verifier))
	api.Use(middleware.LoadUser(userService))

	api.HandleFunc("/me", meHandler.GetMe).Methods(http.MethodGet)

	// Administration — superadmin only (platform-wide flag, independent of
	// any organization_members role).
	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.RequireSuperAdmin)

	admin.HandleFunc("/email-config", adminHandler.GetEmailConfig).Methods(http.MethodGet)
	admin.HandleFunc("/email-config", adminHandler.UpdateEmailConfig).Methods(http.MethodPut)
	admin.HandleFunc("/email-config/test", adminHandler.TestEmailConfig).Methods(http.MethodPost)

	admin.HandleFunc("/sectors", adminHandler.ListAllSectors).Methods(http.MethodGet)
	admin.HandleFunc("/sectors", adminHandler.CreateSector).Methods(http.MethodPost)
	admin.HandleFunc("/sectors/{code}", adminHandler.UpdateSector).Methods(http.MethodPut)
	admin.HandleFunc("/sectors/{code}", adminHandler.DeleteSector).Methods(http.MethodDelete)

	admin.HandleFunc("/oauth-config", adminHandler.GetOAuthConfig).Methods(http.MethodGet)
	admin.HandleFunc("/db-info", adminHandler.GetDBInfo).Methods(http.MethodGet)
	admin.HandleFunc("/migrations", migrationsHandler.ListMigrations).Methods(http.MethodGet)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Printf("supplier-safety-form backend listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
