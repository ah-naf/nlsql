// components/schema/TableItem.tsx
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Table, ChevronDown, ChevronRight, AlertCircle } from "lucide-react";

import { BriefTable, TableInfo } from "@/types/schema";
import { ColumnsList } from "./ColumnList";
import { ForeignKeysList } from "./ForeignKeysList";

type TableItemProps = {
  table: BriefTable;
  isOpen: boolean;
  tableData: TableInfo | undefined;
  schemaError: string;
  onToggle: (tableName: string, isOpen: boolean) => void;
};

export function TableItem({
  table,
  isOpen,
  tableData,
  schemaError,
  onToggle,
}: TableItemProps) {
  return (
    <Collapsible
      open={isOpen}
      onOpenChange={(open) => onToggle(table.name, open)}
      className="w-full"
    >
      <CollapsibleTrigger className="w-full px-3 py-3 flex items-center justify-between border rounded-md bg-white hover:bg-gray-50 cursor-pointer">
        <div className="flex items-center gap-2">
          <Table size={16} className="text-indigo-600" />
          <span className="font-semibold text-gray-800 truncate max-w-48">
            {table.name}
          </span>
          <span className="text-xs px-1.5 py-0.5 bg-gray-100 rounded-full text-gray-600">
            {table.row_count} rows
          </span>
        </div>
        {isOpen ? (
          <ChevronDown size={16} className="text-gray-400 flex-shrink-0" />
        ) : (
          <ChevronRight size={16} className="text-gray-400 flex-shrink-0" />
        )}
      </CollapsibleTrigger>

      {tableData && (
        <CollapsibleContent className="px-2 pb-3">
          {schemaError && (
            <div className="mt-2 border rounded-md border-red-400 p-3 text-center text-red-500">
              <AlertCircle size={20} className="mx-auto mb-2 text-red-400" />
              <p className="text-sm">{schemaError}</p>
            </div>
          )}

          {/* Description */}
          {tableData.description.Valid && (
            <div className="px-3 py-2 text-sm text-gray-600 italic">
              {tableData.description.String}
            </div>
          )}

          {/* Columns */}
          <ColumnsList columns={tableData.columns} />

          {/* Foreign Keys */}
          <ForeignKeysList columns={tableData.columns} />
        </CollapsibleContent>
      )}
    </Collapsible>
  );
}
