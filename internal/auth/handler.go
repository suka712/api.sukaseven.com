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

type EmailResponse struct {
	Message string `json:"message"`
}

type Handler struct {
	Queries *gendb.Queries
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

	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(5 * time.Minute),
		Valid: true,
	}
	ctx := r.Context()

	err = h.Queries.SaveOTP(ctx, gendb.SaveOTPParams{
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
	
	util.WriteJSON(w, http.StatusBadRequest, EmailResponse{Message: "OTP sent"})
}
