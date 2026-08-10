package models

import (
	"errors"
	"net/mail"
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

	return nil
}
