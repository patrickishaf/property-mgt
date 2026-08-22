package common

type ProcessingStatus string

const (
	ProcessingStatusCancelled  ProcessingStatus = "cancelled"
	ProcessingStatusFailed     ProcessingStatus = "failed"
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusProcessing ProcessingStatus = "processing"
	ProcessingStatusSuccessful ProcessingStatus = "successful"
)

func (this ProcessingStatus) String() string {
	return string(this)
}
