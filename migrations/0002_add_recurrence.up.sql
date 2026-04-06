CREATE TABLE IF NOT EXISTS task_statuses (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

INSERT INTO task_statuses (code, name) VALUES
    ('new', 'Новая'),
    ('in_progress', 'В работе'),
    ('done', 'Выполнена')
ON CONFLICT DO NOTHING;

ALTER TABLE tasks 
    ADD CONSTRAINT fk_tasks_status FOREIGN KEY (status) REFERENCES task_statuses(code);

CREATE TABLE IF NOT EXISTS recurrence_rule_types (
    code TEXT PRIMARY KEY,
    description TEXT NOT NULL
);

INSERT INTO recurrence_rule_types (code, description) VALUES
    ('daily', 'Ежедневно (с интервалом)'),
    ('monthly_day', 'Ежемесячно (в определенный день)'),
    ('parity', 'По четным/нечетным дням'),
    ('specific_dates', 'На конкретные даты')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS parity_types (
    code TEXT PRIMARY KEY,
    description TEXT NOT NULL
);

INSERT INTO parity_types (code, description) VALUES
    ('even', 'Четные дни'),
    ('odd', 'Нечетные дни')
ON CONFLICT DO NOTHING;


CREATE TABLE IF NOT EXISTS task_recurrence_rules (
    id BIGSERIAL PRIMARY KEY,
    rule_type_code TEXT NOT NULL REFERENCES recurrence_rule_types(code),
    
    interval_days INT CHECK (interval_days > 0),
    monthly_day INT CHECK (monthly_day >= 1 AND monthly_day <= 31),
    parity_type_code TEXT REFERENCES parity_types(code),
    
    valid_until DATE, 
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS task_recurrence_dates (
    rule_id BIGINT NOT NULL REFERENCES task_recurrence_rules(id) ON DELETE CASCADE,
    exact_date DATE NOT NULL,
    PRIMARY KEY (rule_id, exact_date)
);

ALTER TABLE tasks 
    ADD COLUMN scheduled_date DATE NOT NULL DEFAULT CURRENT_DATE,
    ADD COLUMN recurrence_rule_id BIGINT REFERENCES task_recurrence_rules(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_date ON tasks (scheduled_date);
CREATE INDEX IF NOT EXISTS idx_tasks_recurrence_rule_id ON tasks (recurrence_rule_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_tasks_rule_date 
    ON tasks (recurrence_rule_id, scheduled_date)
    WHERE recurrence_rule_id IS NOT NULL;

COMMIT;