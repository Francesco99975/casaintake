package tools

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Francesco99975/casaintake/internal/models"
)

func GenerateCSV(patient models.PatientIntakeRequest) (string, error) {
	// Generate filename with SID and timestamp
	var filePrefix string
	if patient.OHIP == "" {
		filePrefix = "no-ohip" + patient.FirstName + "-" + patient.LastName
	} else {
		filePrefix = patient.OHIP + "-" + patient.FirstName + "-" + patient.LastName
	}
	filePrefix = strings.ReplaceAll(filePrefix, " ", "_")
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%s-%d.csv", filePrefix, timestamp)

	// Open the file for writing
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Initialize CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write Overall Results Section
	var ohip string
	if patient.OHIP == "" {
		ohip = "N/A"
	} else {
		ohip = patient.OHIP
	}
	overallData := [][]string{
		{"OHIP", ohip},
		{"First Name", patient.FirstName},
		{"Last Name", patient.LastName},
		{"Date of Birth", patient.MakeDOBReadable()},
		{"Email", patient.Email},
		{"Phone", patient.Phone},
		{"Address", patient.StringifyAddress()},
		{"Other Insurance", patient.OtherInsurance},
		{"Conditions", strings.Join(patient.Conditions, " ")},
		{"Other Conditions", patient.OtherConditions},
	}
	for _, row := range overallData {
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write overall data: %v", err)
		}
	}

	// Add blank row after overall results
	if err := writer.Write([]string{""}); err != nil {
		return "", fmt.Errorf("failed to write blank row: %v", err)
	}

	return filename, nil
}
