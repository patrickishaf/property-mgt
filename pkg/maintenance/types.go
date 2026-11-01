package maintenance

import (
	"time"

	"github.com/patrickishaf/property-mgt/internal/db/model"
)

const defaultPageSize = 10

type pageReq struct {
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
	Page  int `form:"page" binding:"omitempty,min=1"`
}

func (this pageReq) values() (int, int) {
	page, limit := this.Page, this.Limit
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = defaultPageSize
	}
	return page, limit
}

type createComplaintReq struct {
	Category string `json:"category" binding:"required,oneof=electrical plumbing structural lighting security other"`
	Message  string `json:"message" binding:"required"`
	Urgency  string `json:"urgency" binding:"required,oneof=low normal urgent"`
}

type getComplaintsReq struct {
	pageReq
	Category   string `form:"category" binding:"omitempty,oneof=electrical plumbing structural lighting security other"`
	Linked     *bool  `form:"linked"`
	PropertyID uint   `form:"propertyId"`
	Status     string `form:"status" binding:"omitempty,oneof=open in_progress fixed"`
	Urgency    string `form:"urgency" binding:"omitempty,oneof=low normal urgent"`
}

type getLinkedComplaintsReq struct {
	IssueID uint `form:"issueId" binding:"required"`
}

type createIssueReq struct {
	Category     string `json:"category" binding:"required,oneof=electrical plumbing structural lighting security other"`
	ComplaintIDs []uint `json:"complaintIds" binding:"omitempty,dive,gt=0"`
	Message      string `json:"message" binding:"required"`
	PropertyID   uint   `json:"propertyId" binding:"required"`
	Title        string `json:"title" binding:"required"`
	Urgency      string `json:"urgency" binding:"required,oneof=low normal urgent"`
}

type linkComplaintsReq struct {
	ComplaintIDs []uint `json:"complaintIds" binding:"required,min=1,dive,gt=0"`
}

type postIssueUpdateReq struct {
	Message        string   `json:"message" binding:"required"`
	NotifyChannels []string `json:"notifyChannels" binding:"omitempty,dive,oneof=in_app sms email"`
	Status         string   `json:"status" binding:"omitempty,oneof=open in_progress fixed"`
}

type getIssuesReq struct {
	pageReq
	Category   string `form:"category" binding:"omitempty,oneof=electrical plumbing structural lighting security other"`
	PropertyID uint   `form:"propertyId"`
	Status     string `form:"status" binding:"omitempty,oneof=open in_progress fixed"`
	Urgency    string `form:"urgency" binding:"omitempty,oneof=low normal urgent"`
}

type listRes[T any] struct {
	Items []T   `json:"items"`
	Limit int   `json:"limit"`
	Page  int   `json:"page"`
	Total int64 `json:"total"`
}

type complaintRes struct {
	ID             uint      `json:"id"`
	Category       string    `json:"category"`
	CreatedAt      time.Time `json:"createdAt"`
	IssueID        *uint     `json:"issueId"`
	Message        string    `json:"message"`
	PropertyID     int       `json:"propertyId"`
	Status         string    `json:"status"`
	TenantFullName string    `json:"tenantFullName"`
	TenantID       int       `json:"tenantId"`
	UnitNumber     string    `json:"unitNumber"`
	Urgency        string    `json:"urgency"`
}

type complaintDetailRes struct {
	complaintRes
	IssueTitle string `json:"issueTitle"`
}

type issueRes struct {
	ID                   uint      `json:"id"`
	Category             string    `json:"category"`
	CreatedAt            time.Time `json:"createdAt"`
	CreatorFullName      string    `json:"creatorFullName"`
	LinkedComplaintCount int64     `json:"linkedComplaintCount"`
	Message              string    `json:"message"`
	PropertyID           int       `json:"propertyId"`
	Status               string    `json:"status"`
	Title                string    `json:"title"`
	Urgency              string    `json:"urgency"`
}

type issueDetailRes struct {
	issueRes
	PropertyName string `json:"propertyName"`
}

// updateRes is one entry in an issue's or a complaint's timeline.
type updateRes struct {
	ID              uint      `json:"id"`
	CreatedAt       time.Time `json:"createdAt"`
	CreatorFullName string    `json:"creatorFullName"`
	Message         string    `json:"message"`
	Status          string    `json:"status"`
}

type notificationRes struct {
	Channel    string `json:"channel"`
	Error      string `json:"error,omitempty"`
	MessageID  *uint  `json:"messageId"`
	Recipients int    `json:"recipients"`
	Skipped    int    `json:"skipped"`
}

type postIssueUpdateRes struct {
	Notifications []notificationRes `json:"notifications"`
	Update        updateRes         `json:"update"`
}

func toComplaintRes(c model.Complaint) complaintRes {
	return complaintRes{
		ID:             c.ID,
		Category:       c.Category,
		CreatedAt:      c.CreatedAt,
		IssueID:        c.IssueID,
		Message:        c.Message,
		PropertyID:     c.PropertyID,
		Status:         c.Status,
		TenantFullName: c.CreatorFullName,
		TenantID:       c.CreatorID,
		UnitNumber:     c.UnitNumber,
		Urgency:        c.Urgency,
	}
}

func toIssueRes(i model.Issue, linkedCount int64) issueRes {
	return issueRes{
		ID:                   i.ID,
		Category:             i.Category,
		CreatedAt:            i.CreatedAt,
		CreatorFullName:      i.CreatorFullName,
		LinkedComplaintCount: linkedCount,
		Message:              i.Message,
		PropertyID:           i.PropertyID,
		Status:               i.Status,
		Title:                i.Title,
		Urgency:              i.Urgency,
	}
}

func toIssueUpdateRes(u model.IssueUpdate) updateRes {
	return updateRes{ID: u.ID, CreatedAt: u.CreatedAt, CreatorFullName: u.CreatorFullName, Message: u.Message, Status: u.Status}
}

func toComplaintUpdateRes(u model.ComplaintUpdate) updateRes {
	return updateRes{ID: u.ID, CreatedAt: u.CreatedAt, CreatorFullName: u.CreatorFullName, Message: u.Message, Status: u.Status}
}
