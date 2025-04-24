export type BriefTable = {
  name: string;
  row_count: number;
};

export type ColumnInfo = {
  name: string;
  data_type: string;
  is_nullable: boolean;
  default_value: { String: string; Valid: boolean };
  is_primary_key: boolean;
  is_unique: boolean;
  foreign_table: { String: string; Valid: boolean };
  foreign_column: { String: string; Valid: boolean };
  description: { String: string; Valid: boolean };
};

export type TableInfo = {
  name: string;
  columns: ColumnInfo[];
  description: { String: string; Valid: boolean };
  row_count: number;
};

export type DatabaseSchema = {
  [tableName: string]: TableInfo;
};
