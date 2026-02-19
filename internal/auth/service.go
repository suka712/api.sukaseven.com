package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/resend/resend-go/v3"
)

func generateOTP() (string, error) {
	max := big.NewInt(1000000)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (h *Handler) sendEmail(to string, otp string) error {
	params := &resend.SendEmailRequest{
		From: "Khiem's little mailman <khiem@sukaseven.com>",
		To:   []string{to},
		Html: fmt.Sprintf(`
			<p>Hello!</p>
			<br>
			<p>Your OTP is <b>%s</b></p>
			<p>It will expire in 5 minutes.</p>
			<br>
			<p>Thankyou! Feel free to reply to this email to chat.</p>`, otp),
		Subject: "OTP for Sukaseven",
	}

	_, err := h.EmailClient.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}
