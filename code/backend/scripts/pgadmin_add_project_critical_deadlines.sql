-- pgAdmin seed for project dashboard critical-to-deadline items.
-- This adds one demo stage and two tasks for the target project so the
-- project dashboard has visible items in criticalToDeadline.

BEGIN;

SET CONSTRAINTS ALL DEFERRED;

INSERT INTO stages (
    id,
    project_id,
    name,
    goal,
    stage_order,
    status,
    deadline_at,
    coverage,
    readiness,
    risk,
    created_at,
    updated_at,
    archived_at
)
VALUES (
    'b8d8c2e8-4b67-4b8a-a0d6-30b7b0e8e1b1',
    '3d7c7a60-6f52-4e3a-9f7f-68d4c2c3d1a1',
    'Critical deadlines demo',
    'Demo stage for dashboard criticalToDeadline examples',
    (
        SELECT COALESCE(MAX(stage_order), 0) + 1
        FROM stages
        WHERE project_id = '3d7c7a60-6f52-4e3a-9f7f-68d4c2c3d1a1'
    ),
    'active',
    NULL,
    0,
    0,
    'medium',
    now(),
    now(),
    NULL
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO tasks (
    id,
    stage_id,
    outcome,
    difficulty,
    priority,
    exam_ready_min_mastery,
    require_non_critical_block_risk,
    deadline_at,
    status,
    coverage,
    readiness,
    risk,
    created_at,
    updated_at,
    archived_at
)
VALUES
    (
        'c1f0f1b5-8c2d-4b8f-a7f0-6c1f7c0a0001',
        'b8d8c2e8-4b67-4b8a-a0d6-30b7b0e8e1b1',
        'Finish the first high-priority review pass',
        5,
        'high',
        70,
        TRUE,
        now() + interval '24 hours',
        'at_risk',
        0,
        0,
        'medium',
        now(),
        now(),
        NULL
    ),
    (
        'c1f0f1b5-8c2d-4b8f-a7f0-6c1f7c0a0002',
        'b8d8c2e8-4b67-4b8a-a0d6-30b7b0e8e1b1',
        'Prepare the second follow-up checkpoint',
        6,
        'high',
        75,
        TRUE,
        now() + interval '36 hours',
        'learning',
        0,
        0,
        'critical',
        now(),
        now(),
        NULL
    )
ON CONFLICT (id) DO NOTHING;

INSERT INTO task_competency_blocks (
    task_id,
    block_type,
    weight,
    mastery,
    last_assessed_at,
    decay_profile,
    risk_state,
    created_at,
    updated_at
)
VALUES
    (
        'c1f0f1b5-8c2d-4b8f-a7f0-6c1f7c0a0001',
        'comprehension',
        40,
        68,
        NULL,
        NULL,
        'high',
        now(),
        now()
    ),
    (
        'c1f0f1b5-8c2d-4b8f-a7f0-6c1f7c0a0001',
        'memorization',
        30,
        55,
        NULL,
        NULL,
        'medium',
        now(),
        now()
    ),
    (
        'c1f0f1b5-8c2d-4b8f-a7f0-6c1f7c0a0001',
        'practice',
        30,
        60,
        NULL,
        NULL,
        'medium',
        now(),
        now()
    ),
    (
        'c1f0f1b5-8c2d-4b8f-a7f0-6c1f7c0a0002',
        'comprehension',
        50,
        72,
        NULL,
        NULL,
        'critical',
        now(),
        now()
    ),
    (
        'c1f0f1b5-8c2d-4b8f-a7f0-6c1f7c0a0002',
        'memorization',
        30,
        58,
        NULL,
        NULL,
        'high',
        now(),
        now()
    ),
    (
        'c1f0f1b5-8c2d-4b8f-a7f0-6c1f7c0a0002',
        'practice',
        20,
        62,
        NULL,
        NULL,
        'medium',
        now(),
        now()
    )
ON CONFLICT DO NOTHING;

COMMIT;