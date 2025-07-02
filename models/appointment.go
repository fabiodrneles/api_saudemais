package models

import (
	"time"

	"gorm.io/gorm"
)

type Appointment struct {
	gorm.Model
	PatientID uint      `json:"patient_id"`
	Patient   Patient   `json:"-" gorm:"foreignKey:PatientID"`
	DateTime  time.Time `json:"datetime"`
}
