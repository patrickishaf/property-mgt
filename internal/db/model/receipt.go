package model

import "gorm.io/gorm"

type Receipt struct {
	gorm.Model
	PaymentID                     uint `gorm:"index"`
	AmountPaid                    string
	OutstandingInstallmentPayment string
	InstallmentAmount             string
	InstallmentAmountPaid         string
	InstallmentDueDate            string
	InstallmentStatus             string
	PaymentDate                   string
	PaymentMethod                 string
	PreviousPaymentAmount         string
	SenderFullName                string
	SenderEmail                   string
}
