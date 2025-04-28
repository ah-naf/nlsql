export interface DBConfig {
  host: string;
  port: string;
  user: string;
  pass: string;
  dbname: string;
}

export interface ConfirmationDialog {
  open: boolean;
  sql: string;
  pendingQuery: string;
}

export interface ResultItem {
  type: "user" | "assistant";
  // eslint-disable-next-line
  content?: any[];
  responseType?: "success" | "error";
  message?: string;
  sql?: string;
  sqlType?: string;
  affectedRows?: number;
  isQAResponse?: boolean;
  extractedSql?: string | null;
}

export interface TableCellProps {
  // eslint-disable-next-line
  content: any;
  isQAColumn: boolean;
}

export interface ResultTableProps {
  // eslint-disable-next-line
  data: any[];
  isQAResponse?: boolean;
  extractedSql?: string | null;
  messageIndex: number;
  onExecuteSql?: (sql: string) => void;
}
