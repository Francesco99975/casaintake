package helpers

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Francesco99975/casaintake/cmd/boot"
	"github.com/Francesco99975/casaintake/internal/models"
	"github.com/Francesco99975/casaintake/internal/tools"
	"github.com/resend/resend-go/v3"
)

func Notify(topic string, message string) {
	resp, err := http.Post(fmt.Sprintf("%s/%s", boot.Environment.NTFY, topic), "text/plain",
		strings.NewReader(message))

	if err != nil {
		slog.Warn("Failed to send notification", slog.Any("error", err))
	}

	if resp != nil {
		slog.Debug("Notification sent with status code", slog.Int("status", resp.StatusCode))
		defer func() { _ = resp.Body.Close() }()
	}
}

func ResendNewPatientTemplate(email string, patient models.PatientIntakeRequest) {
	client := resend.NewClient(boot.Environment.ResendAPIKey)

	filename, err := tools.GenerateCSV(patient)
	if err != nil {
		slog.Error("Failed to generate CSV", slog.Any("error", err))
		return
	}

	contentBytes, err := os.ReadFile(filename)
	if err != nil {
		slog.Error("Failed to read CSV", slog.Any("error", err))
		return
	}

	defer os.Remove(filename)

	params := &resend.SendEmailRequest{
		From: "Patients Intake <intake@auth.urx.ink>",
		To:   []string{email},
		Template: &resend.EmailTemplate{
			Id: "patient-form-submission",
			Variables: map[string]any{
				"FIRSTNAME":       patient.FirstName,
				"LASTNAME":        patient.LastName,
				"DOB":             patient.MakeDOBReadable(),
				"ADDRESS":         patient.StringifyAddress(),
				"PHONE":           patient.Phone,
				"OHIP":            patient.OHIP,
				"INSURANCE_NOTES": patient.OtherInsurance,
				"PATIENT_EMAIL":   patient.Email,
				"MEDICAL_HISTORY": strings.Join(patient.Conditions, ","),
				"EXTRA_NOTES":     patient.OtherConditions,
				"SUBMITTED_AT":    time.Now().Format("January 2, 2006 at 15:04"),
			},
		},
		Attachments: []*resend.Attachment{
			{
				Filename:    path.Base(filename),
				Content:     contentBytes,
				ContentType: "text/csv",
			},
		},
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		slog.Error("Failed to send email", slog.Any("error", err))
		return
	}

	slog.Info("Email sent to", slog.String("email", email), slog.String("id", sent.Id))
}
