package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
	"rest-refs/internal/app/api"
)

// Handler struct wraps service interface, which interacts with business logic
type Handler struct {
	service api.Service
	logger  *logrus.Logger
}

// New creates new Handler instance and takes api.Service and logger as parameters
func New(service api.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers HTTP routes
func (h *Handler) RegisterRoutes(r *mux.Router) {
	authRouter := r.PathPrefix("/auth").Subrouter()

	authRouter.HandleFunc("/register", h.RegisterUserHandler).Methods("POST")
	authRouter.HandleFunc("/login", h.LoginUserHandler).Methods("POST")
	//authRouter.HandleFunc("/register/referral", h.RegisterWithReferralHandler).Methods("POST")

	referralCodeRouter := r.PathPrefix("/referral_code").Subrouter()

	createReferralCodeRouter := http.HandlerFunc(h.CreateReferralCodeHandler)
	referralCodeRouter.Handle("", h.RequireValidTokenMiddleware(createReferralCodeRouter)).Methods("POST")

	deleteReferralCodeRouter := http.HandlerFunc(h.DeleteReferralCodeHandler)
	referralCodeRouter.Handle("", h.RequireValidTokenMiddleware(deleteReferralCodeRouter)).Methods("DELETE")

	referralCodeRouter.HandleFunc("/email/{email}", h.GetReferralCodeByEmailHandler).Methods("GET")

	referralRouter := r.PathPrefix("/referral").Subrouter()
	referralRouter.HandleFunc("/id/{referrer_id}", h.GetReferralsByReferrerIDHandler).Methods("GET")

	// Swagger documentation endpoint
	r.PathPrefix("/docs/swagger/").Handler(httpSwagger.WrapHandler)
	r.HandleFunc("/docs/swagger/index.html", httpSwagger.WrapHandler)
}

// StartServer initializes and starts HTTP server on given port
func (h *Handler) StartServer(port string) {
	router := mux.NewRouter()

	// Middleware for processing request ID
	router.Use(h.RequestIDMiddleware)
	h.RegisterRoutes(router)

	if err := http.ListenAndServe(port, router); err != nil {
		h.logger.Fatalf("Не удалось запустить сервер: %s", err)
	}
}
