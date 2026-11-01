package lease

import (
	"sort"
	"time"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/db/model"
)

// buildDueItems turns open installments and open leases without installments
// into due items, keeps those matching status ("overdue", "upcoming" or ""),
// and sorts them by due date so overdue items come first. Names are filled in
// by the caller.
func buildDueItems(installments []model.Installment, leases []model.Lease, now time.Time, status string) []dueItemRes {
	items := make([]dueItemRes, 0, len(installments)+len(leases))
	for _, i := range installments {
		if i.DueDate == nil {
			continue
		}
		id := i.ID
		items = append(items, dueItemRes{
			Amount:        i.OutstandingAmount,
			DueDate:       i.DueDate.Format(dateLayout),
			InstallmentID: &id,
			LeaseID:       uint(i.LeaseID),
			Overdue:       i.IsOverdue(now),
			PropertyID:    uint(i.PropertyID),
			TenantID:      uint(i.TenantID),
			Type:          dueTypeInstallment,
			dueAt:         *i.DueDate,
		})
	}
	for _, l := range leases {
		items = append(items, dueItemRes{
			Amount:     l.OutstandingRent,
			DueDate:    l.StartDate.Format(dateLayout),
			LeaseID:    l.ID,
			Overdue:    l.StartDate.Before(now) && l.OutstandingRent > 0,
			PropertyID: uint(l.PropertyID),
			TenantID:   uint(l.TenantID),
			Type:       dueTypeLease,
			UnitNumber: l.UnitNumber,
			dueAt:      l.StartDate,
		})
	}

	filtered := items[:0]
	for _, item := range items {
		if (status == "overdue" && !item.Overdue) || (status == "upcoming" && item.Overdue) {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.SliceStable(filtered, func(a, b int) bool {
		if !filtered[a].dueAt.Equal(filtered[b].dueAt) {
			return filtered[a].dueAt.Before(filtered[b].dueAt)
		}
		return filtered[a].LeaseID < filtered[b].LeaseID
	})
	return filtered
}

// paginate returns one page of items.
func paginate[T any](items []T, page int, limit int) []T {
	start := (page - 1) * limit
	if start >= len(items) {
		return []T{}
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func (this *service) getDue(actor *common.Actor, dto getDueReq) common.ServiceResponse {
	logger := this.logger.With("caller", "lease.service.getDue", "actor", actor.Email)
	var ownerID, managerID, tenantID uint
	switch {
	case actor.Role == common.ActorRoleTenant:
		tenantID = actor.ID
	case isAdmin(actor):
		ownerID, managerID = adminScope(actor)
	default:
		return forbidden()
	}

	installments, err := this.installmentRepo.FindOpen(ownerID, managerID, dto.PropertyID, tenantID)
	if err != nil {
		logger.Error("failed to find open installments", "err", err.Error())
		return serverError()
	}
	leases, err := this.leaseRepo.FindOpenUnscheduled(ownerID, managerID, dto.PropertyID, tenantID)
	if err != nil {
		logger.Error("failed to find open leases", "err", err.Error())
		return serverError()
	}

	items := buildDueItems(installments, leases, time.Now(), dto.Status)
	page, limit := dto.values()
	pageItems := paginate(items, page, limit)
	if err := this.fillDueNames(pageItems); err != nil {
		logger.Error("failed to load due item names", "err", err.Error())
		return serverError()
	}
	return common.ServiceResponse{StatusCode: 200, Data: listRes[dueItemRes]{
		Items: pageItems,
		Limit: limit,
		Page:  page,
		Total: int64(len(items)),
	}}
}

// fillDueNames adds tenant names, property names and (for installments) unit
// numbers to a page of due items.
func (this *service) fillDueNames(items []dueItemRes) error {
	var tenantIDs, propertyIDs, leaseIDs []uint
	for _, item := range items {
		tenantIDs = append(tenantIDs, item.TenantID)
		propertyIDs = append(propertyIDs, item.PropertyID)
		if item.Type == dueTypeInstallment {
			leaseIDs = append(leaseIDs, item.LeaseID)
		}
	}
	tenants, err := this.tenantRepo.FindByIds(tenantIDs)
	if err != nil {
		return err
	}
	properties, err := this.propertyRepo.FindByIds(propertyIDs)
	if err != nil {
		return err
	}
	leases, err := this.leaseRepo.FindByIds(leaseIDs)
	if err != nil {
		return err
	}
	names := make(map[uint]string, len(tenants))
	for _, t := range tenants {
		names[t.ID] = t.FirstName + " " + t.LastName
	}
	propertyNames := make(map[uint]string, len(properties))
	for _, p := range properties {
		propertyNames[p.ID] = p.Name
	}
	units := make(map[uint]string, len(leases))
	for _, l := range leases {
		units[l.ID] = l.UnitNumber
	}
	for i := range items {
		items[i].TenantFullName = names[items[i].TenantID]
		items[i].PropertyName = propertyNames[items[i].PropertyID]
		if items[i].Type == dueTypeInstallment {
			items[i].UnitNumber = units[items[i].LeaseID]
		}
	}
	return nil
}
