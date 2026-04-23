SELECT name, type, usage_count, last_used, success_count, error_count, tags, created_at
FROM catalog
WHERE name = $1 AND type = $2;
