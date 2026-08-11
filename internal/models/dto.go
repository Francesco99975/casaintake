package models

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

// PatientIntakeRequest represents the data submitted by the patient intake form.
type PatientIntakeRequest struct {
	// Personal & Contact
	FirstName string `form:"firstname" validate:"required,min=1,max=100"`
	LastName  string `form:"lastname" validate:"required,min=1,max=100"`
	DOB       string `form:"dob" validate:"required"` //YYYY-MM-DD
	Phone     string `form:"phone" validate:"required"`
	Email     string `form:"email" validate:"required,email"`

	// Address (adjust nested form tags to match what ui.AddressInput actually posts)
	//
	AddressStreetNumber string `form:"address_street_number" validate:"required,max=100"`
	AddressStreetName   string `form:"address_street_name" validate:"required,max=200"`
	AddressUnit         string `form:"address_unit" validate:"required,max=200"`
	AddressCity         string `form:"address_city" validate:"required,max=100"`
	AddressRegion       string `form:"address_region" validate:"required,max=100"`
	AddressPostal       string `form:"address_postal" validate:"required,max=10"`
	AddressCountry      string `form:"address_country" validate:"required,max=10"`

	// Insurance
	OHIP           string `form:"ohip" validate:"omitempty,max=20"`
	OtherInsurance string `form:"other_insurance" validate:"omitempty,max=1000"`

	// Medical History
	Conditions      []string `form:"conditions" validate:"omitempty,dive,oneof=diabetes cholesterol blood_pressure heart_problems arthritis respiratory hearing_issues skin_conditions"`
	OtherConditions string   `form:"other_conditions" validate:"omitempty,max=2000"`
}

func (p *PatientIntakeRequest) MakeDOBReadable() string {
	t, err := time.Parse("2006-01-02", p.DOB)
	if err != nil {
		return ""
	}
	return t.Format("January 2, 2006")
}

func (p *PatientIntakeRequest) StringifyAddress() string {
	if p.AddressUnit != "" {
		return p.AddressStreetNumber + " " + p.AddressStreetName + ", " + p.AddressUnit + ", " + p.AddressCity + ", " + p.AddressRegion + " " + p.AddressPostal
	}
	return p.AddressStreetNumber + " " + p.AddressStreetName + ", " + p.AddressCity + ", " + p.AddressRegion + " " + p.AddressPostal
}

func (p *PatientIntakeRequest) NormalizePhone() (string, error) {
	// Keep only digits.
	digits := regexp.MustCompile(`\D`).ReplaceAllString(p.Phone, "")

	// Remove North American country code.
	if len(digits) == 11 && digits[0] == '1' {
		digits = digits[1:]
	}

	if len(digits) != 10 {
		return "", fmt.Errorf("invalid phone number: %q", p.Phone)
	}

	return fmt.Sprintf(
		"+1 (%s) %s-%s",
		digits[:3],
		digits[3:6],
		digits[6:],
	), nil
}

func (p *PatientIntakeRequest) NormalizeOHIP() (string, error) {
	// Remove spaces and dashes, normalize case.
	input := strings.ToUpper(strings.TrimSpace(p.OHIP))
	input = strings.ReplaceAll(input, " ", "")
	input = strings.ReplaceAll(input, "-", "")

	// OHIP: 10 digits total + 2-letter version code.
	re := regexp.MustCompile(`^(\d{9})(\d)([A-Z]{2})$`)
	matches := re.FindStringSubmatch(input)

	if matches == nil {
		return "", fmt.Errorf("invalid OHIP format: %q", input)
	}

	// Format as: 1234-567-889-VC
	return fmt.Sprintf(
		"%s-%s-%s-%s",
		matches[1][:4],
		matches[1][4:7],
		matches[1][7:],
		matches[3],
	), nil
}

func (p *PatientIntakeRequest) Validate() error {
	if p.FirstName == "" {
		return errors.New("first name is required")
	}
	if p.LastName == "" {
		return errors.New("last name is required")
	}
	if p.DOB == "" {
		return errors.New("dob is required")
	}
	if p.MakeDOBReadable() == "" {
		return errors.New("invalid dob")
	}
	if p.Phone == "" {
		return errors.New("phone is required")
	}

	normalized, err := p.NormalizePhone()
	if err != nil {
		return err
	}
	p.Phone = normalized

	if p.Email == "" {
		return errors.New("email is required")
	}

	addr, err := mail.ParseAddress(p.Email)
	if err != nil {
		return errors.New("invalid email address")
	}
	if addr.Name != "" {
		return errors.New("invalid email address")
	}
	p.Email = addr.Address

	if p.AddressStreetNumber == "" {
		return errors.New("address street number is required")
	}
	if p.AddressStreetName == "" {
		return errors.New("address street name is required")
	}
	if p.AddressCity == "" {
		return errors.New("address city is required")
	}
	if p.AddressRegion == "" {
		return errors.New("address region is required")
	}
	if p.AddressPostal == "" {
		return errors.New("address postal is required")
	}
	if p.AddressCountry == "" {
		return errors.New("address country is required")
	}

	if p.OHIP == "" && p.OtherInsurance == "" {
		return errors.New("no insurance information provided, either OHIP or other insurance is required")
	}

	if p.OHIP != "" {
		normalized, err := p.NormalizeOHIP()
		if err != nil {
			return err
		}
		p.OHIP = normalized
	}

	return nil
}
