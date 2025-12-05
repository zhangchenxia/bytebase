-- Add sensitive_data_rule table to store sensitive data classification rules
CREATE TABLE sensitive_data_rule (
    id serial PRIMARY KEY,
    deleted boolean NOT NULL DEFAULT FALSE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    name text NOT NULL,
    project_id integer NOT NULL,
    -- Stored as SensitiveDataRule (proto/v1/sensitive_data_service.proto)
    rule jsonb NOT NULL DEFAULT '{}',
    creator_id integer NOT NULL,
    updater_id integer NOT NULL,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE,
    FOREIGN KEY (creator_id) REFERENCES principal(id) ON DELETE SET NULL,
    FOREIGN KEY (updater_id) REFERENCES principal(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX idx_sensitive_data_rule_unique_name_project ON sensitive_data_rule(name, project_id) WHERE deleted = FALSE;
CREATE INDEX idx_sensitive_data_rule_project ON sensitive_data_rule(project_id) WHERE deleted = FALSE;

ALTER SEQUENCE sensitive_data_rule_id_seq RESTART WITH 101;

-- Add approval_flow table to store approval flow configurations
CREATE TABLE approval_flow (
    id serial PRIMARY KEY,
    deleted boolean NOT NULL DEFAULT FALSE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    name text NOT NULL,
    project_id integer NOT NULL,
    -- Stored as ApprovalFlow (proto/v1/sensitive_data_service.proto)
    flow jsonb NOT NULL DEFAULT '{}',
    creator_id integer NOT NULL,
    updater_id integer NOT NULL,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE,
    FOREIGN KEY (creator_id) REFERENCES principal(id) ON DELETE SET NULL,
    FOREIGN KEY (updater_id) REFERENCES principal(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX idx_approval_flow_unique_name_project ON approval_flow(name, project_id) WHERE deleted = FALSE;
CREATE INDEX idx_approval_flow_project ON approval_flow(project_id) WHERE deleted = FALSE;

ALTER SEQUENCE approval_flow_id_seq RESTART WITH 101;

-- Add columns to issue table for sensitive data tracking
ALTER TABLE issue ADD COLUMN IF NOT EXISTS sensitive_data_level text;
ALTER TABLE issue ADD COLUMN IF NOT EXISTS sensitive_field_changes jsonb NOT NULL DEFAULT '[]';
ALTER TABLE issue ADD COLUMN IF NOT EXISTS sensitive_data_approval_flow jsonb;

-- Add index for sensitive data level
CREATE INDEX IF NOT EXISTS idx_issue_sensitive_data_level ON issue(sensitive_data_level);
