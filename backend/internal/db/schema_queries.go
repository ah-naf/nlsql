package db

// PostgreSQL queries
var PostgresQueries = struct {
	Columns     string
	PrimaryKeys string
	UniqueKeys  string
	ForeignKeys string
	RowCount    string
}{
	Columns: `
SELECT
  c.table_name,
  c.column_name,
  c.data_type,
  (c.is_nullable = 'YES') AS is_nullable,
  c.column_default,
  pgd.description   AS column_description,
  tbl_pgd.description AS table_description
FROM information_schema.columns AS c
LEFT JOIN pg_catalog.pg_stat_all_tables AS st
  ON c.table_name = st.relname
LEFT JOIN pg_catalog.pg_description AS pgd
  ON pgd.objoid = st.relid AND pgd.objsubid = c.ordinal_position
LEFT JOIN pg_catalog.pg_description AS tbl_pgd
  ON tbl_pgd.objoid = st.relid AND tbl_pgd.objsubid = 0
WHERE c.table_schema = 'public'
ORDER BY c.table_name, c.ordinal_position;
`,
	PrimaryKeys: `
SELECT tc.table_name, kcu.column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_schema = kcu.constraint_schema
 AND tc.constraint_name = kcu.constraint_name
WHERE tc.constraint_type='PRIMARY KEY' AND tc.table_schema='public';
`,
	UniqueKeys: `
SELECT tc.table_name, kcu.column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_schema = kcu.constraint_schema
 AND tc.constraint_name = kcu.constraint_name
WHERE tc.constraint_type='UNIQUE' AND tc.table_schema='public';
`,
	ForeignKeys: `
SELECT
  kcu.table_name,
  kcu.column_name,
  ccu.table_name AS foreign_table_name,
  ccu.column_name AS foreign_column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_schema = kcu.constraint_schema
 AND tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage ccu
  ON tc.constraint_schema = ccu.constraint_schema
 AND tc.constraint_name = ccu.constraint_name
WHERE tc.constraint_type='FOREIGN KEY' AND tc.table_schema='public';
`,
	RowCount: "SELECT COUNT(*) FROM %s",
}

// MySQL queries
var MySQLQueries = struct {
	Columns     string
	PrimaryKeys string
	UniqueKeys  string
	ForeignKeys string
	RowCount    string
}{
	Columns: `
SELECT
  c.table_name,
  c.column_name,
  c.data_type,
  (c.is_nullable = 'YES') AS is_nullable,
  c.column_default,
  NULL AS column_description,
  NULL AS table_description
FROM information_schema.columns AS c
WHERE c.table_schema = DATABASE()
ORDER BY c.table_name, c.ordinal_position;
`,
	PrimaryKeys: `
SELECT tc.table_name, kcu.column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_schema = kcu.constraint_schema
 AND tc.constraint_name = kcu.constraint_name
WHERE tc.constraint_type='PRIMARY KEY' AND tc.table_schema=DATABASE();
`,
	UniqueKeys: `
SELECT tc.table_name, kcu.column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_schema = kcu.constraint_schema
 AND tc.constraint_name = kcu.constraint_name
WHERE tc.constraint_type='UNIQUE' AND tc.table_schema=DATABASE();
`,
	ForeignKeys: `
SELECT
  kcu.table_name,
  kcu.column_name,
  kcu.referenced_table_name AS foreign_table_name,
  kcu.referenced_column_name AS foreign_column_name
FROM information_schema.key_column_usage kcu
WHERE kcu.constraint_schema = DATABASE()
  AND kcu.referenced_table_name IS NOT NULL;
`,
	RowCount: "SELECT COUNT(*) FROM `%s`",
}
