-- +goose Up
CREATE TABLE IF NOT EXISTS asset_movement_history (
  movement_history_id uuid PRIMARY KEY,
  group_id uuid NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  asset_id integer NOT NULL,
  entity_id uuid NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
  from_location_id uuid NULL REFERENCES entities(id) ON DELETE SET NULL,
  to_location_id uuid NULL REFERENCES entities(id) ON DELETE SET NULL,
  movement_type varchar(100) NOT NULL,
  moved_by_user_id uuid NULL REFERENCES users(id) ON DELETE SET NULL,
  other_person_name varchar(255) NULL,
  other_person_email varchar(255) NULL,
  other_person_department varchar(255) NULL,
  movement_date_time datetime NOT NULL,
  notes text NULL,
  created_by_user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_date_time datetime NOT NULL,
  previous_status varchar(100) NULL,
  new_status varchar(100) NULL,
  checked_out_due_date datetime NULL,
  returned_date_time datetime NULL,
  approved_by_user_id uuid NULL REFERENCES users(id) ON DELETE SET NULL,
  condition_before_move varchar(255) NULL,
  condition_after_move varchar(255) NULL,
  is_active boolean NOT NULL DEFAULT true,
  is_deleted boolean NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_asset_movement_history_asset_group_date
  ON asset_movement_history (group_id, asset_id, movement_date_time DESC);

-- +goose Down
DROP TABLE IF EXISTS asset_movement_history;
