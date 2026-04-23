INSERT INTO catalog (name, type, usage_count, last_used, success_count, error_count, created_at)
VALUES ($1, $2, 1, NOW(), $3::BIGINT, $4::BIGINT, NOW())
ON CONFLICT (name, type)
DO UPDATE SET
    usage_count = catalog.usage_count + 1,
    last_used = NOW(),
    success_count = catalog.success_count + $3::BIGINT,
    error_count = catalog.error_count + $4::BIGINT;
