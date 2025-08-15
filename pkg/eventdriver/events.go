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

	msg := models.Event{
		ID:            uuid.New().String(),
		ParentID:      data.Id,
		ParentType:    models.EventParentSystem,
		TransactionID: txid,
		Tenant:        Config.Events.Data.Tenant,
		Attributes: models.EventAttributes{
			Source: "passage-server",
			Type:   fmt.Sprintf("%s.passage.accessRequest.created", Config.Events.Data.TypePrefix),
			Date:   time.Now(),
			Author: uid,
		},
		Message: fmt.Sprintf("[%s] [%s] Created AccessRequest: [%s] role [%s]", Config.Events.Data.Tenant, uid, data.Id, data.RoleRef.Name),
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

	msg := models.Event{
		ID:            uuid.New().String(),
		ParentID:      data.Id,
		ParentType:    models.EventParentSystem,
		TransactionID: txid,
		Tenant:        Config.Events.Data.Tenant,
		Attributes: models.EventAttributes{
			Source: "passage-server",
			Type:   fmt.Sprintf("%s.passage.accessRequest.approved", Config.Events.Data.TypePrefix),
			Date:   time.Now(),
			Author: uid,
		},
		Message: fmt.Sprintf("[%s] [%s] Approved AccessRequest [%s] Role [%s] added to user [%s]", Config.Events.Data.Tenant, uid, data.Id, data.RoleRef.Name, data.Status.RequestedBy),
		Data: map[string]interface{}{
			"resource": data,
		},
	}

	return e.handleEvent(ctx, msg)
}

func (e *Events) AccessRequestApprovalError(ctx context.Context, data models.AccessRequest, provider models.ProviderConfig, err error) error {

	ctx = shared.WithTransactionID(ctx)
	txid, _ := shared.GetTransactionID(ctx)
	uid, _ := shared.GetUserID(ctx)

	msg := models.Event{
		ID:            uuid.New().String(),
		ParentID:      data.Id,
		ParentType:    models.EventParentSystem,
		TransactionID: txid,
		Tenant:        Config.Events.Data.Tenant,
		Attributes: models.EventAttributes{
			Source: "passage-server",
			Type:   fmt.Sprintf("%s.passage.accessRequest.approvalError", Config.Events.Data.TypePrefix),
			Date:   time.Now(),
			Author: uid,
		},
		Message: fmt.Sprintf("[%s] [%s] There was an error approving AccessRequest [%s] Role [%s] provider [%s] error: [%s]", Config.Events.Data.Tenant, uid, data.Id, data.RoleRef.Name, provider.Name, err.Error()),
		Data: map[string]interface{}{
			"resource": data,
			"error":    err,
		},
	}

	return e.handleEvent(ctx, msg)
}

func (e *Events) AccessRequestExpireError(ctx context.Context, data models.AccessRequest, provider models.ProviderConfig, err error) error {

	ctx = shared.WithTransactionID(ctx)
	txid, _ := shared.GetTransactionID(ctx)
	uid, _ := shared.GetUserID(ctx)

	msg := models.Event{
		ID:            uuid.New().String(),
		ParentID:      data.Id,
		ParentType:    models.EventParentSystem,
		TransactionID: txid,
		Tenant:        Config.Events.Data.Tenant,
		Attributes: models.EventAttributes{
			Source: "passage-server",
			Type:   fmt.Sprintf("%s.passage.accessRequest.expireError", Config.Events.Data.TypePrefix),
			Date:   time.Now(),
			Author: uid,
		},
		Message: fmt.Sprintf("[%s] [%s] There was an error expiring AccessRequest [%s] Role [%s] provider [%s] error: [%s]", Config.Events.Data.Tenant, uid, data.Id, data.RoleRef.Name, provider.Name, err.Error()),
		Data: map[string]interface{}{
			"resource": data,
			"error":    err,
		},
	}

	return e.handleEvent(ctx, msg)
}

func (e *Events) AccessRequestExpired(ctx context.Context, data models.AccessRequest) error {

	ctx = shared.WithTransactionID(ctx)
	txid, _ := shared.GetTransactionID(ctx)
	uid, _ := shared.GetUserID(ctx)

	msg := models.Event{
		ID:            uuid.New().String(),
		ParentID:      data.Id,
		ParentType:    models.EventParentSystem,
		TransactionID: txid,
		Tenant:        Config.Events.Data.Tenant,
		Attributes: models.EventAttributes{
			Source: "passage-server",
			Type:   fmt.Sprintf("%s.passage.accessRequest.expired", Config.Events.Data.TypePrefix),
			Date:   time.Now(),
			Author: uid,
		},
		Message: fmt.Sprintf("[%s] AccessRequest [%s] expired. Role [%s] removed from user [%s]", Config.Events.Data.Tenant, data.Id, data.RoleRef.Name, data.Status.RequestedBy),
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

	msg := models.Event{
		ID:            uuid.New().String(),
		ParentID:      data.Id,
		ParentType:    models.EventParentSystem,
		TransactionID: txid,
		Tenant:        Config.Events.Data.Tenant,
		Attributes: models.EventAttributes{
			Source: "passage-server",
			Type:   fmt.Sprintf("%s.passage.accessRequest.deleted", Config.Events.Data.TypePrefix),
			Date:   time.Now(),
			Author: uid,
		},
		Message: fmt.Sprintf("[%s] [%s] Deleted AccessRequest [%s] Role [%s] User [%s]", Config.Events.Data.Tenant, uid, data.Id, data.RoleRef.Name, data.Status.RequestedBy),
		Data: map[string]interface{}{
			"resource": data,
		},
	}

	return e.handleEvent(ctx, msg)
}

func (e *Events) UserLoggedIn(ctx context.Context, claims models.ClaimsMap) error {

	ctx = shared.WithTransactionID(ctx)
	txid, _ := shared.GetTransactionID(ctx)

	msg := models.Event{
		ID:            uuid.New().String(),
		ParentID:      "",
		ParentType:    models.EventParentSecurity,
		TransactionID: txid,
		Tenant:        Config.Events.Data.Tenant,
		Attributes: models.EventAttributes{
			Source: "passage-server",
			Type:   fmt.Sprintf("%s.passage.user.loggedIn", Config.Events.Data.TypePrefix),
			Date:   time.Now(),
			Author: "system",
		},
		Message: fmt.Sprintf("[%s] User [%s] logged in", Config.Events.Data.Tenant, claims.GetString("username")),
		Data: map[string]interface{}{
			"resource": claims,
		},
	}

	return e.handleEvent(ctx, msg)
}

func (e *Events) PermissionDenied(ctx context.Context, sub string, grp []string, obj string, act string) error {

	ctx = shared.WithTransactionID(ctx)
	txid, _ := shared.GetTransactionID(ctx)

	msg := models.Event{
		ID:            uuid.New().String(),
		ParentID:      "",
		ParentType:    models.EventParentSecurity,
		TransactionID: txid,
		Tenant:        Config.Events.Data.Tenant,
		Attributes: models.EventAttributes{
			Source: "passage-server",
			Type:   fmt.Sprintf("%s.passage.user.permissionDenied", Config.Events.Data.TypePrefix),
			Date:   time.Now(),
			Author: "system",
		},
		Message: fmt.Sprintf("[%s] Permission denied: subj:%s obj:%s act:%s", Config.Events.Data.Tenant, sub, obj, act),
		Data: map[string]interface{}{
			"sub": sub,
			"grp": grp,
			"obj": obj,
			"act": act,
		},
	}

	return e.handleEvent(ctx, msg)
}
