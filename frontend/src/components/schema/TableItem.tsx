// components/schema/TableItem.tsx
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  Table,
  ChevronDown,
  ChevronRight,
  AlertCircle,
  Loader,
} from "lucide-react";
import { BriefTable, TableInfo } from "@/types/schema";
import { ColumnsList } from "./ColumnList";
import { ForeignKeysList } from "./ForeignKeysList";

type TableItemProps = {
  table: BriefTable;
  isOpen: boolean;
  tableData?: TableInfo;
  schemaError: string;
  onToggle: (tableName: string, isOpen: boolean) => void;
  isLoading?: boolean;
  disabled?: boolean;
};

export function TableItem({
  table,
  isOpen,
  tableData,
  schemaError,
  onToggle,
  isLoading = false,
  disabled = false,
}: TableItemProps) {
  return (
    <Collapsible
      open={isOpen}
      onOpenChange={(open) => {
        if (!disabled && !isLoading) {
          onToggle(table.name, open);
        }
      }}
      className="w-full"
    >
      <CollapsibleTrigger
        className={`w-full px-3 py-3 flex items-center gap-2 border rounded-md bg-white
          ${
            disabled || isLoading
              ? "opacity-50 cursor-not-allowed"
              : "hover:bg-gray-50 cursor-pointer"
          }
        `}
      >
        {isLoading ? (
          <Loader size={16} className="text-indigo-600 animate-spin" />
        ) : (
          <Table size={16} className="text-indigo-600" />
        )}
        <span className="font-semibold text-gray-800 truncate max-w-48">
          {table.name}
        </span>
        <span className="text-xs px-1.5 py-0.5 bg-gray-100 rounded-full text-gray-600">
          {table.row_count} rows
        </span>
        {isOpen ? (
          <ChevronDown size={16} className="text-gray-400 ml-auto" />
        ) : (
          <ChevronRight size={16} className="text-gray-400 ml-auto" />
        )}
      </CollapsibleTrigger>

      <CollapsibleContent className="px-2 pb-3">
        {isLoading && (
          <div className="animate-pulse mt-3">
            <div className="h-4 bg-gray-200 rounded-full w-3/4 mb-2.5"></div>
            <div className="space-y-2 mt-3">
              {[1, 2, 3, 4].map((i) => (
                <div key={i} className="flex items-center gap-2">
                  <div className="h-3 bg-gray-200 rounded-full w-2"></div>
                  <div className="h-3 bg-gray-200 rounded-full w-1/4"></div>
                  <div className="ml-auto h-3 bg-gray-200 rounded-full w-1/5"></div>
                </div>
              ))}
            </div>
          </div>
        )}

        {!isLoading && schemaError && (
          <div className="mt-2 border rounded-md border-red-400 p-3 text-center text-red-500">
            <AlertCircle size={20} className="mx-auto mb-2 text-red-400" />
            <p className="text-sm">{schemaError}</p>
          </div>
        )}

        {!isLoading && tableData && (
          <>
            {tableData.description.Valid && (
              <div className="px-3 py-2 text-sm text-gray-600 italic">
                {tableData.description.String}
              </div>
            )}
            <ColumnsList columns={tableData.columns} />
            <ForeignKeysList columns={tableData.columns} />
          </>
        )}
      </CollapsibleContent>
    </Collapsible>
  );
}
