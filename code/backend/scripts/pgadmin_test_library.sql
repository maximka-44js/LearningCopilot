-- pgAdmin test data for library structure with folders and files
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

BEGIN;

INSERT INTO users (id)
VALUES ('11e10f73-1eb6-428e-b412-28e95941b393')
ON CONFLICT DO NOTHING;

INSERT INTO libraries (id, user_id, title, description, public, created_at, updated_at)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    '11e10f73-1eb6-428e-b412-28e95941b393',
    'Demo Library',
    'Test library for pgAdmin',
    FALSE,
    now(),
    now()
)
ON CONFLICT DO NOTHING;

INSERT INTO folders (id, user_id, library_id, name, parent_id, created_at, updated_at)
VALUES
    (
        '33333333-3333-3333-3333-333333333333',
        '11e10f73-1eb6-428e-b412-28e95941b393',
        '22222222-2222-2222-2222-222222222222',
        'Projects',
        NULL,
        now(),
        now()
    ),
    (
        '44444444-4444-4444-4444-444444444444',
        '11e10f73-1eb6-428e-b412-28e95941b393',
        '22222222-2222-2222-2222-222222222222',
        '2026',
        '33333333-3333-3333-3333-333333333333',
        now(),
        now()
    ),
    (
        '55555555-5555-5555-5555-555555555555',
        '11e10f73-1eb6-428e-b412-28e95941b393',
        '22222222-2222-2222-2222-222222222222',
        'Photos',
        NULL,
        now(),
        now()
    )
ON CONFLICT (id) DO NOTHING;

INSERT INTO files (
    id,
    user_id,
    library_id,
    folder_id,
    name,
    size,
    mime_type,
    checksum_declared,
    checksum,
    status,
    scan_attempts,
    created_at,
    updated_at,
    deleted_at
)
VALUES
    (
        '66666666-6666-6666-6666-666666666666',
        '11e10f73-1eb6-428e-b412-28e95941b393',
        '22222222-2222-2222-2222-222222222222',
        NULL,
        'root-notes.txt',
        1234,
        'text/plain',
        NULL,
        NULL,
        'uploaded',
        0,
        now(),
        now(),
        NULL
    ),
    (
        '77777777-7777-7777-7777-777777777777',
        '11e10f73-1eb6-428e-b412-28e95941b393',
        '22222222-2222-2222-2222-222222222222',
        '33333333-3333-3333-3333-333333333333',
        'project-plan.md',
        2048,
        'text/markdown',
        NULL,
        NULL,
        'scan_ok',
        1,
        now(),
        now(),
        NULL
    ),
    (
        '88888888-8888-8888-8888-888888888888',
        '11e10f73-1eb6-428e-b412-28e95941b393',
        '22222222-2222-2222-2222-222222222222',
        '44444444-4444-4444-4444-444444444444',
        'budget.csv',
        4096,
        'text/csv',
        NULL,
        NULL,
        'scan_ok',
        1,
        now(),
        now(),
        NULL
    ),
    (
        '99999999-9999-9999-9999-999999999999',
        '11e10f73-1eb6-428e-b412-28e95941b393',
        '22222222-2222-2222-2222-222222222222',
        '55555555-5555-5555-5555-555555555555',
        'photo-01.jpg',
        524288,
        'image/jpeg',
        NULL,
        NULL,
        'uploaded',
        0,
        now(),
        now(),
        NULL
    )
ON CONFLICT (id) DO NOTHING;

COMMIT;
