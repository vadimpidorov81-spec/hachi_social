package httptransport

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	domainusers "github.com/hachisocial/hachisocial/internal/domain/users"
	"github.com/hachisocial/hachisocial/internal/feature/users/application"
	"github.com/hachisocial/hachisocial/internal/platform/httpserver"
)

const maxRequestBody = 64 << 10

type Handler struct {
	useCase   UseCase
	principal PrincipalProvider
}

func NewHandler(useCase UseCase, principal PrincipalProvider) *Handler {
	return &Handler{useCase: useCase, principal: principal}
}

func (h *Handler) Routes() http.Handler {
	router := chi.NewRouter()
	router.Get("/me", h.getCurrent)
	router.Patch("/me", h.updateProfile)
	router.Get("/{username}", h.getPublicProfile)
	return router
}

func (h *Handler) AdminRoutes() http.Handler {
	router := chi.NewRouter()
	router.Put("/{id}/status", h.setStatus)
	return router
}

func (h *Handler) getCurrent(response http.ResponseWriter, request *http.Request) {
	principal, err := h.principal.Principal(request.Context())
	if err != nil {
		writeError(response, request, err)
		return
	}
	user, err := h.useCase.GetCurrent(request.Context(), principal.UserID)
	if err != nil {
		writeError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, envelope{Data: user})
}

func (h *Handler) getPublicProfile(response http.ResponseWriter, request *http.Request) {
	if _, err := h.principal.Principal(request.Context()); err != nil {
		writeError(response, request, err)
		return
	}
	profile, err := h.useCase.GetPublicProfile(request.Context(), chi.URLParam(request, "username"))
	if err != nil {
		writeError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, envelope{Data: profile})
}

func (h *Handler) updateProfile(response http.ResponseWriter, request *http.Request) {
	principal, err := h.principal.Principal(request.Context())
	if err != nil {
		writeError(response, request, err)
		return
	}

	var body updateProfileRequest
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, request, err)
		return
	}
	user, err := h.useCase.UpdateProfile(
		request.Context(),
		principal.UserID,
		application.UpdateProfileCommand{
			DisplayName: body.DisplayName,
			Bio:         body.Bio,
			Timezone:    body.Timezone,
		},
	)
	if err != nil {
		writeError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, envelope{Data: user})
}

func (h *Handler) setStatus(response http.ResponseWriter, request *http.Request) {
	principal, err := h.principal.Principal(request.Context())
	if err != nil {
		writeError(response, request, err)
		return
	}
	if principal.Role != domainusers.RoleAdmin {
		writeError(response, request, application.ErrForbidden)
		return
	}
	targetID, err := domainusers.ParseID(chi.URLParam(request, "id"))
	if err != nil {
		writeError(response, request, err)
		return
	}

	var body setStatusRequest
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, request, err)
		return
	}
	if err := h.useCase.SetStatus(
		request.Context(),
		principal.UserID,
		targetID,
		body.Status,
	); err != nil {
		writeError(response, request, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

type updateProfileRequest struct {
	DisplayName *string `json:"display_name"`
	Bio         *string `json:"bio"`
	Timezone    *string `json:"timezone"`
}

type setStatusRequest struct {
	Status string `json:"status"`
}

type envelope struct {
	Data any `json:"data"`
}

type errorEnvelope struct {
	Error errorResponse `json:"error"`
}

type errorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func decodeJSON(response http.ResponseWriter, request *http.Request, target any) error {
	request.Body = http.MaxBytesReader(response, request.Body, maxRequestBody)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return application.ErrEmptyUpdate
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return application.ErrEmptyUpdate
	}
	return nil
}

func writeJSON(response http.ResponseWriter, status int, payload any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(payload)
}

func writeError(response http.ResponseWriter, request *http.Request, err error) {
	status, code, message := mapError(err)
	writeJSON(response, status, errorEnvelope{Error: errorResponse{
		Code:      code,
		Message:   message,
		RequestID: httpserver.RequestID(request.Context()),
	}})
}

func mapError(err error) (int, string, string) {
	switch {
	case errors.Is(err, application.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized", "Authentication is required"
	case errors.Is(err, application.ErrForbidden):
		return http.StatusForbidden, "forbidden", "Action is forbidden"
	case errors.Is(err, application.ErrUserNotFound):
		return http.StatusNotFound, "user_not_found", "User was not found"
	case errors.Is(err, application.ErrUsernameAlreadyTaken):
		return http.StatusConflict, "username_already_taken", "Username is already taken"
	case errors.Is(err, application.ErrCannotBlockSelf):
		return http.StatusConflict, "cannot_block_self", "Administrator cannot block own account"
	case errors.Is(err, application.ErrEmptyUpdate):
		return http.StatusBadRequest, "invalid_request", "Request body is invalid or empty"
	case errors.Is(err, domainusers.ErrInvalidID),
		errors.Is(err, domainusers.ErrInvalidUsername),
		errors.Is(err, domainusers.ErrReservedUsername),
		errors.Is(err, domainusers.ErrInvalidDisplayName),
		errors.Is(err, domainusers.ErrInvalidBio),
		errors.Is(err, domainusers.ErrInvalidTimezone):
		return http.StatusBadRequest, "validation_failed", "Request validation failed"
	case errors.Is(err, domainusers.ErrInvalidStatus):
		return http.StatusUnprocessableEntity, "invalid_user_status", "User status is invalid"
	default:
		return http.StatusInternalServerError, "internal_error", "Internal error"
	}
}
