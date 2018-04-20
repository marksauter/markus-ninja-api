SELECT
  a.access_level || ' ' || a.type operation,
  p.audience,
  array_agg(p.field) fields
FROM
  permission p
INNER JOIN permission_activity pa ON pa.permission_id = p.id
INNER JOIN activity a ON a.id = pa.activity_id
WHERE a.access_level = 'Update'
  AND a.type = 'User'
  AND pa.permission_id IN (
    SELECT
      permission_id
    FROM
      role_permission rp
    INNER JOIN role r ON rp.role_id = r.id
    WHERE r.name = 'ADMIN'
  )
GROUP BY
  operation, audience;
