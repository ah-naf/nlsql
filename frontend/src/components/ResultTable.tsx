// src/components/ResultTable.tsx
import React from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { PlayCircle } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { ResultTableProps, TableCellProps } from "../types/query";

// Truncated cell component
const TruncatedCell: React.FC<TableCellProps> = ({ content, isQAColumn }) => {
  const contentStr = content === null ? "NULL" : String(content);

  // Always show full text for QA responses in the "output" column
  if (isQAColumn) {
    return content === null ? (
      <span className="text-gray-400">NULL</span>
    ) : (
      <div className="whitespace-pre-wrap">{contentStr}</div>
    );
  }

  // For non-QA data, truncate as before
  const isLong = contentStr.length > 20;

  if (!isLong) {
    return content === null ? (
      <span className="text-gray-400">NULL</span>
    ) : (
      <>{contentStr}</>
    );
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <span className="cursor-help">{contentStr.substring(0, 20)}...</span>
        </TooltipTrigger>
        <TooltipContent side="top" className="max-w-md">
          <p className="break-words">{contentStr}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
};

// Result table component
export default function ResultTable({
  data,
  isQAResponse = false,
  extractedSql = null,
  onExecuteSql,
}: ResultTableProps) {
  if (!data || data.length === 0)
    return <div className="text-gray-500">No results found</div>;

  const columns = Object.keys(data[0]);

  // For QA responses with embedded SQL, show a special UI
  if (isQAResponse && columns.includes("output") && extractedSql) {
    return (
      <div className="w-full border rounded bg-white">
        <div className="p-4 border-b">
          <div className="prose max-w-none mb-2">
            <pre className="bg-gray-100 p-3 rounded font-mono text-sm overflow-x-auto">
              {extractedSql}
            </pre>
          </div>
          <div className="flex justify-end mt-2">
            <Button
              onClick={() => onExecuteSql && onExecuteSql(extractedSql)}
              className="bg-green-600 hover:bg-green-700"
            >
              <PlayCircle className="h-4 w-4 mr-2" />
              Execute SQL
            </Button>
          </div>
        </div>
      </div>
    );
  }

  // For QA responses, render a special version with full text
  else if (isQAResponse && columns.includes("output")) {
    return (
      <div className="w-full p-4 border rounded bg-white">
        <div className="prose max-w-none">{data[0].output}</div>
      </div>
    );
  }

  // For regular database results, render the table
  return (
    <div className="w-full overflow-auto max-h-96 border rounded">
      <Table>
        <TableHeader>
          <TableRow>
            {columns.map((column, i) => (
              <TableHead key={i} className="bg-gray-50">
                {column}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((row, rowIndex) => (
            <TableRow key={rowIndex}>
              {columns.map((column, colIndex) => (
                <TableCell key={colIndex} className="max-w-xs">
                  <TruncatedCell
                    content={row[column]}
                    isQAColumn={isQAResponse && column === "output"}
                  />
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
