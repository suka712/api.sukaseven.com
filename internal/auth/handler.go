package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/resend/resend-go/v3"
	"github.com/suka712/api.sukaseven.com/internal/db/generated"
	"github.com/suka712/api.sukaseven.com/util"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type EmailRequest struct {
	Email string `json:"email"`
}

type OTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type EmailResponse struct {
	Message string `json:"message"`
}

type Handler struct {
	Queries     *gendb.Queries
	EmailClient *resend.Client
}

func (h *Handler) Email(w http.ResponseWriter, r *http.Request) {
	var req EmailRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Print("Error decoding request:", err)
		util.WriteJSON(w, http.StatusBadRequest, util.ErrorResponse{Error: "Invalid request"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	if !emailRegex.MatchString(email) {
		log.Print("Invalid email")
		util.WriteJSON(w, http.StatusBadRequest, util.ErrorResponse{Error: "Invalid email"})
		return
	}

	otpString, err := generateOTP()
	if err != nil {
		log.Print("Failed to generate OTP")
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}

	h.Queries.DeleteOTP(r.Context(), email)

	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(5 * time.Minute),
		Valid: true,
	}

	err = h.Queries.SaveOTP(r.Context(), gendb.SaveOTPParams{
		Email:     email,
		Otp:       otpString,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		log.Print("Error saving OTP:", err)
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}

	err = h.sendEmail(email, otpString)
	if err != nil {
		log.Print("Failed to send email:", err)
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}

	util.WriteJSON(w, http.StatusOK, EmailResponse{Message: "OTP sent"})
}

func (h *Handler) Session(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		util.WriteJSON(w, http.StatusUnauthorized, util.ErrorResponse{Error: "No session"})
		return
	}

	var token pgtype.UUID
	err = token.Scan(cookie.Value)
	if err != nil {
		util.WriteJSON(w, http.StatusUnauthorized, util.ErrorResponse{Error: "Invalid session"})
		return
	}

	session, err := h.Queries.GetSession(r.Context(), token)
	if err != nil {
		util.WriteJSON(w, http.StatusUnauthorized, util.ErrorResponse{Error: "Invalid session"})
		return
	}

	util.WriteJSON(w, http.StatusOK, map[string]string{"email": session.Email})
}

func (h *Handler) OTP(w http.ResponseWriter, r *http.Request) {
	var req OTPRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Print("Error decoding request:", err)
		util.WriteJSON(w, http.StatusBadRequest, util.ErrorResponse{Error: "Invalid request"})
		return
	}

	otp := strings.TrimSpace(req.OTP)
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if len(otp) != 6 || email == "" {
		log.Print("Invalid otp or email")
		util.WriteJSON(w, http.StatusBadRequest, util.ErrorResponse{Error: "Invalid request"})
		return
	}

	ctx := r.Context()
	stored, err := h.Queries.GetOTP(ctx, email)
	if err != nil {
		log.Print("Error retrieving OTP from db:", err)
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}

	if stored.Otp != otp {
		log.Print("Invalid OTP")
		util.WriteJSON(w, http.StatusBadRequest, util.ErrorResponse{Error: "Invalid OTP"})
		return
	}

	err = h.Queries.DeleteOTP(ctx, email)

	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	sessionId, err := h.Queries.CreateSession(ctx, gendb.CreateSessionParams{
		Email:     email,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		log.Print("Failed to create session")
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionId.Token.String(),
		Domain:   ".sukaseven.com",
		Expires:  expiresAt.Time,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	util.WriteJSON(w, http.StatusOK, EmailResponse{Message: "Token set"})
}
