package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MovementType string

const (
	MovementTypeTransfer    MovementType = "Transfer"
	MovementTypeCheckOut    MovementType = "Check Out"
	MovementTypeCheckIn     MovementType = "Check In"
	MovementTypeReturn      MovementType = "Return"
	MovementTypeMaintenance MovementType = "Maintenance"
)

type AssetMovementCreate struct {
	AssetID                int64        `json:"assetId" validate:"required"`
	ToLocationID           uuid.UUID    `json:"toLocationId"`
	MovementType           MovementType `json:"movementType" validate:"required"`
	MovedByUserID          *uuid.UUID   `json:"movedByUserId"`
	OtherPersonName        string       `json:"otherPersonName"`
	OtherPersonEmail       string       `json:"otherPersonEmail"`
	OtherPersonDepartment  string       `json:"otherPersonDepartment"`
	Notes                  string       `json:"notes"`
	PreviousStatus         string       `json:"previousStatus"`
	NewStatus              string       `json:"newStatus"`
	CheckedOutDueDate      *time.Time   `json:"checkedOutDueDate"`
	ReturnedDateTime       *time.Time   `json:"returnedDateTime"`
	ApprovedByUserID       *uuid.UUID   `json:"approvedByUserId"`
	ConditionBeforeMove    string       `json:"conditionBeforeMove"`
	ConditionAfterMove     string       `json:"conditionAfterMove"`
	OverrideCheckedOutLock bool         `json:"overrideCheckedOutLock"`
}

type AssetMovementHistory struct {
	MovementHistoryID     uuid.UUID  `json:"movementHistoryId"`
	AssetID               int64      `json:"assetId"`
	EntityID              uuid.UUID  `json:"entityId"`
	FromLocationID        *uuid.UUID `json:"fromLocationId"`
	ToLocationID          *uuid.UUID `json:"toLocationId"`
	MovementType          string     `json:"movementType"`
	MovedByUserID         *uuid.UUID `json:"movedByUserId"`
	OtherPersonName       string     `json:"otherPersonName"`
	OtherPersonEmail      string     `json:"otherPersonEmail"`
	OtherPersonDepartment string     `json:"otherPersonDepartment"`
	MovementDateTime      time.Time  `json:"movementDateTime"`
	Notes                 string     `json:"notes"`
	CreatedByUserID       uuid.UUID  `json:"createdByUserId"`
	CreatedDateTime       time.Time  `json:"createdDateTime"`
	PreviousStatus        string     `json:"previousStatus"`
	NewStatus             string     `json:"newStatus"`
	CheckedOutDueDate     *time.Time `json:"checkedOutDueDate"`
	ReturnedDateTime      *time.Time `json:"returnedDateTime"`
	ApprovedByUserID      *uuid.UUID `json:"approvedByUserId"`
	ConditionBeforeMove   string     `json:"conditionBeforeMove"`
	ConditionAfterMove    string     `json:"conditionAfterMove"`
	IsActive              bool       `json:"isActive"`
	IsDeleted             bool       `json:"isDeleted"`
}

func (r *EntityRepository) GetAssetMovementHistory(ctx context.Context, gid uuid.UUID, assetID int64, limit int) ([]AssetMovementHistory, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	query := `
		SELECT
			movement_history_id,
			asset_id,
			entity_id,
			from_location_id,
			to_location_id,
			movement_type,
			moved_by_user_id,
			COALESCE(other_person_name, ''),
			COALESCE(other_person_email, ''),
			COALESCE(other_person_department, ''),
			movement_date_time,
			COALESCE(notes, ''),
			created_by_user_id,
			created_date_time,
			COALESCE(previous_status, ''),
			COALESCE(new_status, ''),
			checked_out_due_date,
			returned_date_time,
			approved_by_user_id,
			COALESCE(condition_before_move, ''),
			COALESCE(condition_after_move, ''),
			is_active,
			is_deleted
		FROM asset_movement_history
		WHERE group_id = $1 AND asset_id = $2 AND is_deleted = false
		ORDER BY movement_date_time DESC
		LIMIT $3
	`

	rows, err := r.db.Sql().QueryContext(ctx, query, gid, assetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]AssetMovementHistory, 0)
	for rows.Next() {
		var row AssetMovementHistory
		if err := rows.Scan(
			&row.MovementHistoryID,
			&row.AssetID,
			&row.EntityID,
			&row.FromLocationID,
			&row.ToLocationID,
			&row.MovementType,
			&row.MovedByUserID,
			&row.OtherPersonName,
			&row.OtherPersonEmail,
			&row.OtherPersonDepartment,
			&row.MovementDateTime,
			&row.Notes,
			&row.CreatedByUserID,
			&row.CreatedDateTime,
			&row.PreviousStatus,
			&row.NewStatus,
			&row.CheckedOutDueDate,
			&row.ReturnedDateTime,
			&row.ApprovedByUserID,
			&row.ConditionBeforeMove,
			&row.ConditionAfterMove,
			&row.IsActive,
			&row.IsDeleted,
		); err != nil {
			return nil, err
		}
		results = append(results, row)
	}

	return results, rows.Err()
}

func (r *EntityRepository) CreateAssetMovement(ctx context.Context, gid, uid uuid.UUID, in AssetMovementCreate) (AssetMovementHistory, error) {
	if in.AssetID <= 0 {
		return AssetMovementHistory{}, fmt.Errorf("invalid asset id")
	}

	items, err := r.QueryByAssetID(ctx, gid, AssetID(in.AssetID), 1, 1)
	if err != nil {
		return AssetMovementHistory{}, err
	}
	if len(items.Items) == 0 {
		return AssetMovementHistory{}, fmt.Errorf("asset not found")
	}

	entity, err := r.GetOneByGroup(ctx, gid, items.Items[0].ID)
	if err != nil {
		return AssetMovementHistory{}, err
	}

	var fromLocationID *uuid.UUID
	if entity.Parent != nil {
		id := entity.Parent.ID
		fromLocationID = &id
	}

	movements, err := r.GetAssetMovementHistory(ctx, gid, in.AssetID, 1)
	if err != nil {
		return AssetMovementHistory{}, err
	}
	if len(movements) == 1 && movements[0].NewStatus == string(MovementTypeCheckOut) && !in.OverrideCheckedOutLock {
		return AssetMovementHistory{}, fmt.Errorf("asset is already checked out; confirmation override required")
	}

	newStatus := in.NewStatus
	if newStatus == "" {
		switch in.MovementType {
		case MovementTypeCheckOut:
			newStatus = "Checked Out"
		case MovementTypeCheckIn, MovementTypeReturn:
			newStatus = "Available"
		default:
			newStatus = in.PreviousStatus
		}
	}

	if in.ToLocationID != uuid.Nil {
		patch := EntityPatch{ID: entity.ID, ParentID: in.ToLocationID}
		err = r.Patch(ctx, gid, entity.ID, patch)
		if err != nil {
			return AssetMovementHistory{}, err
		}
	}

	id := uuid.New()
	now := time.Now().UTC()

	query := `
		INSERT INTO asset_movement_history (
			movement_history_id, group_id, asset_id, entity_id, from_location_id, to_location_id,
			movement_type, moved_by_user_id, other_person_name, other_person_email, other_person_department,
			movement_date_time, notes, created_by_user_id, created_date_time, previous_status, new_status,
			checked_out_due_date, returned_date_time, approved_by_user_id, condition_before_move, condition_after_move,
			is_active, is_deleted
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''),
			$12, NULLIF($13, ''), $14, $15, NULLIF($16, ''), NULLIF($17, ''),
			$18, $19, $20, NULLIF($21, ''), NULLIF($22, ''),
			true, false
		)
	`
	_, err = r.db.Sql().ExecContext(
		ctx,
		query,
		id,
		gid,
		in.AssetID,
		entity.ID,
		fromLocationID,
		nullableUUID(in.ToLocationID),
		in.MovementType,
		in.MovedByUserID,
		in.OtherPersonName,
		in.OtherPersonEmail,
		in.OtherPersonDepartment,
		now,
		in.Notes,
		uid,
		now,
		in.PreviousStatus,
		newStatus,
		in.CheckedOutDueDate,
		in.ReturnedDateTime,
		in.ApprovedByUserID,
		in.ConditionBeforeMove,
		in.ConditionAfterMove,
	)
	if err != nil {
		return AssetMovementHistory{}, err
	}

	created, err := r.GetAssetMovementHistory(ctx, gid, in.AssetID, 1)
	if err != nil {
		return AssetMovementHistory{}, err
	}
	if len(created) == 0 {
		return AssetMovementHistory{}, sql.ErrNoRows
	}

	return created[0], nil
}

func nullableUUID(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}
