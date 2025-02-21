package eventdriver

import (
	"context"
	"fmt"
	"time"

	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/shared"
	"github.com/google/uuid"
)

func (e *Events) AccessRequestCreated(ctx context.Context, data models.AccessRequest) error {

	ctx = shared.WithTransactionID(ctx)
	txid, _ := shared.GetTransactionID(ctx)
	uid, _ := shared.GetUserID(ctx)

	msg := Event{
		ID:            uuid.New().String(),
		ParentID:      data.Id,
		ParentType:    EventParentSystem,
		TransactionID: txid,
		Tenant:        Config.Data.Tenant,
		Attributes: EventAttributes{
			Source: "passage-server/controllers",
			Type:   fmt.Sprintf("%s.passage.accessRequest.created", Config.Data.TypePrefix),
			Date:   time.Now(),
			Author: uid,
		},
		Message: fmt.Sprintf("[%s] [%s] Created AccessRequest: [%s] role [%s]", Config.Data.Tenant, uid, data.Id, data.RoleRef.Name),
		Data: map[string]interface{}{
			"resource": data,
		},
	}

	return e.handleEvent(ctx, msg)
}

func (e *Events) AccessRequestApproved(ctx context.Context, data models.AccessRequest) error {

	ctx = shared.WithTransactionID(ctx)
	txid, _ := shared.GetTransactionID(ctx)
	uid, _ := shared.GetUserID(ctx)

	msg := Event{
		ID:            uuid.New().String(),
		ParentID:      data.Id,
		ParentType:    EventParentSystem,
		TransactionID: txid,
		Tenant:        Config.Data.Tenant,
		Attributes: EventAttributes{
			Source: "passage-server/pkg/controllers",
			Type:   fmt.Sprintf("%s.passage.accessRequest.approved", Config.Data.TypePrefix),
			Date:   time.Now(),
			Author: uid,
		},
		Message: fmt.Sprintf("[%s] [%s] Approved AccessRequest [%s] Role [%s] added to user [%s]", Config.Data.Tenant, uid, data.Id, data.RoleRef.Name, data.Status.RequestedBy),
		Data: map[string]interface{}{
			"resource": data,
		},
	}

	return e.handleEvent(ctx, msg)
}

func (e *Events) AccessRequestExpired(ctx context.Context, data models.AccessRequest) error {

	ctx = shared.WithTransactionID(ctx)
	txid, _ := shared.GetTransactionID(ctx)
	uid, _ := shared.GetUserID(ctx)

	msg := Event{
		ID:            uuid.New().String(),
		ParentID:      data.Id,
		ParentType:    EventParentSystem,
		TransactionID: txid,
		Tenant:        Config.Data.Tenant,
		Attributes: EventAttributes{
			Source: "passage-server/pkg/controllers",
			Type:   fmt.Sprintf("%s.passage.accessRequest.expired", Config.Data.TypePrefix),
			Date:   time.Now(),
			Author: uid,
		},
		Message: fmt.Sprintf("[%s] AccessRequest [%s] expired. Role [%s] removed from user [%s]", Config.Data.Tenant, data.Id, data.RoleRef.Name, data.Status.RequestedBy),
		Data: map[string]interface{}{
			"resource": data,
		},
	}

	return e.handleEvent(ctx, msg)
}

func (e *Events) AccessRequestDeleted(ctx context.Context, data models.AccessRequest) error {

	ctx = shared.WithTransactionID(ctx)
	txid, _ := shared.GetTransactionID(ctx)
	uid, _ := shared.GetUserID(ctx)

	msg := Event{
		ID:            uuid.New().String(),
		ParentID:      data.Id,
		ParentType:    EventParentSystem,
		TransactionID: txid,
		Tenant:        Config.Data.Tenant,
		Attributes: EventAttributes{
			Source: "passage-server/pkg/controllers",
			Type:   fmt.Sprintf("%s.passage.accessRequest.deleted", Config.Data.TypePrefix),
			Date:   time.Now(),
			Author: uid,
		},
		Message: fmt.Sprintf("[%s] [%s] Deleted AccessRequest [%s] Role [%s] User [%s]", Config.Data.Tenant, uid, data.Id, data.RoleRef.Name, data.Status.RequestedBy),
		Data: map[string]interface{}{
			"resource": data,
		},
	}

	return e.handleEvent(ctx, msg)
}

func (e *Events) UserLoggedIn(ctx context.Context, claims models.ClaimsMap) error {

	ctx = shared.WithTransactionID(ctx)
	txid, _ := shared.GetTransactionID(ctx)

	msg := Event{
		ID:            uuid.New().String(),
		ParentID:      "",
		ParentType:    EventParentSecurity,
		TransactionID: txid,
		Tenant:        Config.Data.Tenant,
		Attributes: EventAttributes{
			Source: "passage-server/pkg/middlewares/auth",
			Type:   fmt.Sprintf("%s.passage.user.loggedIn", Config.Data.TypePrefix),
			Date:   time.Now(),
			Author: "system",
		},
		Message: fmt.Sprintf("[%s] User [%s] logged in", Config.Data.Tenant, claims.GetString("username")),
		Data: map[string]interface{}{
			"resource": claims,
		},
	}

	return e.handleEvent(ctx, msg)
}

func (e *Events) PermissionDenied(ctx context.Context, sub string, grp []string, obj string, act string) error {

	ctx = shared.WithTransactionID(ctx)
	txid, _ := shared.GetTransactionID(ctx)

	msg := Event{
		ID:            uuid.New().String(),
		ParentID:      "",
		ParentType:    EventParentSecurity,
		TransactionID: txid,
		Tenant:        Config.Data.Tenant,
		Attributes: EventAttributes{
			Source: "passage-server/pkg/middlewares/auth",
			Type:   fmt.Sprintf("%s.passage.user.permissionDenied", Config.Data.TypePrefix),
			Date:   time.Now(),
			Author: "system",
		},
		Message: fmt.Sprintf("[%s] Permission denied: subj:%s obj:%s act:%s", Config.Data.Tenant, sub, obj, act),
		Data: map[string]interface{}{
			"sub": sub,
			"grp": grp,
			"obj": obj,
			"act": act,
		},
	}

	return e.handleEvent(ctx, msg)
}
