INSERT INTO partners (
    uuid,
    name,
    endpoint,
    is_enabled,
    countries,
    device_types,
    min_bid_floor,
    blocked_categories
)
VALUES
(
    '123e4567-e89b-12d3-a456-426655440001',
    'DSP Alpha',
    'http://localhost:9001/bid',
    TRUE,
    ARRAY['RU', 'KZ'],
    ARRAY['mobile', 'desktop'],
    0.5,
    ARRAY['gambling']
),
(
    '123e4567-e89b-12d3-a456-426655440002',
    'DSP Beta',
    'http://localhost:9002/bid',
    TRUE,
    ARRAY[]::TEXT[],
    ARRAY['mobile'],
    1.0,
    ARRAY['adult']
),
(
    '123e4567-e89b-12d3-a456-426655440003',
    'DSP Gamma',
    'http://localhost:9003/bid',
    TRUE,
    ARRAY['RU'],
    ARRAY[]::TEXT[],
    1.5,
    ARRAY[]::TEXT[]
),
(
    '123e4567-e89b-12d3-a456-426655440004',
    'DSP Delta',
    'http://localhost:9004/bid',
    FALSE,
    ARRAY[]::TEXT[],
    ARRAY[]::TEXT[],
    0,
    ARRAY[]::TEXT[]
);