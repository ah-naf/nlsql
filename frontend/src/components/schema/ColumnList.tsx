import { Key, Hash, Link } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { TooltipProvider } from "@radix-ui/react-tooltip";
import { ColumnInfo } from "@/types/schema";

type ColumnsListProps = {
  columns: ColumnInfo[];
};

export function ColumnsList({ columns }: ColumnsListProps) {
  const getColumnIcon = (col: ColumnInfo) =>
    col.is_primary_key ? (
      <Key size={14} className="text-yellow-500" />
    ) : col.is_unique ? (
      <Hash size={14} className="text-blue-500" />
    ) : col.foreign_table.Valid ? (
      <Link size={14} className="text-green-600" />
    ) : null;

  const getColumnTooltip = (col: ColumnInfo) =>
    col.is_primary_key
      ? "Primary Key"
      : col.is_unique
      ? "Unique Constraint"
      : col.foreign_table.Valid
      ? `Foreign Key → ${col.foreign_table.String}.${col.foreign_column.String}`
      : col.description.Valid
      ? col.description.String
      : "";

  return (
    <div className="border rounded-md overflow-hidden mt-2">
      <div className="px-3 py-2 text-xs font-medium uppercase bg-gray-50 text-gray-500 border-b">
        Columns
      </div>
      <ul className="text-sm divide-y">
        {columns.map((col, i) => (
          <li key={i} className="px-3 py-2 flex items-center hover:bg-gray-50">
            <div className="flex-1 flex items-center gap-1.5">
              {getColumnIcon(col)}
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <span className={col.is_primary_key ? "font-medium" : ""}>
                      {col.name}
                    </span>
                  </TooltipTrigger>
                  {getColumnTooltip(col) && (
                    <TooltipContent>{getColumnTooltip(col)}</TooltipContent>
                  )}
                </Tooltip>
              </TooltipProvider>
            </div>
            <div className="flex items-center gap-1 text-xs text-gray-500">
              <span className="px-1.5 py-0.5 bg-gray-100 rounded-md">
                {col.data_type}
              </span>
              {!col.is_nullable && (
                <span className="text-red-500">required</span>
              )}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
