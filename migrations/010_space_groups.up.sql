CREATE TABLE IF NOT EXISTS space_groups (
    space_id UUID NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id),
    PRIMARY KEY (space_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_space_groups_group ON space_groups(group_id);
