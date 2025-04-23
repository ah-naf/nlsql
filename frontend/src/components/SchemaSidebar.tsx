// components/SchemaSidebar.tsx
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { TooltipProvider } from "@radix-ui/react-tooltip";
import axios from "axios";
import {
  ChevronDown,
  Table,
  Database,
  AlertCircle,
  ChevronRight,
  Key,
  Link,
  Info,
  Hash,
} from "lucide-react";
import { useEffect, useState, useCallback } from "react";

type BriefTable = {
  name: string;
  row_count: number;
};

type ColumnInfo = {
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

type TableInfo = {
  name: string;
  columns: ColumnInfo[];
  description: { String: string; Valid: boolean };
  row_count: number;
};

type DatabaseSchema = {
  [tableName: string]: TableInfo;
};

export default function SchemaSidebar() {
  const dbConfig = JSON.parse(localStorage.getItem("dbConfig") || "null");
  const [briefTables, setBriefTables] = useState<BriefTable[]>([]);
  const [fullSchema, setFullSchema] = useState<DatabaseSchema>({});
  const [openTable, setOpenTable] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState<string>("");
  const [visibleCount, setVisibleCount] = useState<number>(20);
  const [schemaError, setSchemaError] = useState("");

  useEffect(() => {
    (async () => {
      try {
        const res = await axios.get<{ tables: BriefTable[] }>(
          "http://localhost:8080/schema",
          { params: { ...dbConfig, brief: true } }
        );
        setBriefTables(res.data.tables);
        setError(null);
      } catch (e: any) {
        setError(e.response?.data?.error || e.message);
      }
    })();
  }, []);

  const filtered = briefTables.filter((t) =>
    t.name.toLowerCase().includes(search.trim().toLowerCase())
  );
  const visible = filtered.slice(0, visibleCount);

  const handleScroll = useCallback(
    (e: React.UIEvent<HTMLDivElement>) => {
      const { scrollTop, scrollHeight, clientHeight } = e.currentTarget;

      if (
        scrollHeight - scrollTop <= clientHeight + 20 &&
        visibleCount < filtered.length
      ) {
        setVisibleCount((v) => Math.min(v + 20, filtered.length));
      }
    },
    [filtered.length, visibleCount]
  );

  const handleToggle = async (tableName: string, isOpen: boolean) => {
    if (isOpen) {
      setOpenTable(tableName);
      if (!fullSchema[tableName]) {
        try {
          setError("");
          const res = await axios.get<{ table: TableInfo }>(
            `http://localhost:8080/schema/${tableName}`,
            { params: dbConfig }
          );
          setFullSchema((fs) => ({ ...fs, [tableName]: res.data.table }));
        } catch (e: any) {
          console.error("Failed to load table:", e);
          setSchemaError(e.response?.data?.error || e.message);
          setFullSchema({});
        }
      }
    } else {
      setOpenTable(null);
    }
  };

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

  // Render
  return (
    <aside className="w-[25rem] hidden lg:flex flex-col h-full border-r bg-white shadow">
      <div className="p-4 py-5 border-b bg-gray-50">
        <h2 className="text-xl font-bold text-gray-800 flex items-center gap-2">
          <Database size={20} className="text-indigo-600" />
          {dbConfig?.dbname || "Database Schema"}
        </h2>
        <input
          type="search"
          placeholder="🔍 Search tables..."
          className="mt-3 w-full px-2 py-1 border rounded text-sm"
          value={search}
          onChange={(e) => {
            setSearch(e.target.value);
            setVisibleCount(20);
          }}
        />
      </div>

      <ScrollArea
        className="flex-1 px-3 pt-2 pb-8 overflow-y-auto"
        onScrollCapture={handleScroll}
      >
        {error ? (
          <div className="mt-20 px-6 text-center text-red-500">
            <AlertCircle size={48} className="mx-auto mb-4 text-red-400" />
            <p className="font-semibold">{error}</p>
            <p className="text-xs text-gray-500 mt-1">
              Check your connection or credentials.
            </p>
          </div>
        ) : visible.length === 0 ? (
          <div className="mt-20 px-6 text-center text-gray-500">
            <Database size={48} className="mx-auto mb-4 text-indigo-300" />
            <p className="font-medium">No tables found.</p>
          </div>
        ) : (
          visible.map(({ name, row_count }) => {
            const isOpen = openTable === name;
            const table = fullSchema[name];

            return (
              <Collapsible
                key={name}
                open={isOpen}
                onOpenChange={(open) => handleToggle(name, open)}
              >
                <CollapsibleTrigger className="w-full px-3 py-3 mt-1 flex items-center justify-between border rounded-md bg-white hover:bg-gray-50 cursor-pointer">
                  <div className="flex items-center gap-2">
                    <Table size={16} className="text-indigo-600" />
                    <span className="font-semibold text-gray-800">{name}</span>
                    <span className="text-xs px-1.5 py-0.5 bg-gray-100 rounded-full text-gray-600">
                      {row_count} rows
                    </span>
                  </div>
                  {isOpen ? (
                    <ChevronDown size={16} className="text-gray-400" />
                  ) : (
                    <ChevronRight size={16} className="text-gray-400" />
                  )}
                </CollapsibleTrigger>

                {table && (
                  <CollapsibleContent className="px-2 pb-3">
                    {schemaError && (
                      <div className="mt-2 border rounded-md border-red-400 p-3 text-center text-red-500">
                        <AlertCircle
                          size={48}
                          className="mx-auto mb-2 text-red-400"
                        />
                        <p>{schemaError}</p>
                      </div>
                    )}
                    {/* Description */}
                    {table.description.Valid && (
                      <div className="px-3 py-2 text-sm text-gray-600 italic">
                        {table.description.String}
                      </div>
                    )}

                    {/* Columns */}
                    <div className="border rounded-md overflow-hidden">
                      <div className="px-3 py-2 text-xs font-medium uppercase bg-gray-50 text-gray-500 border-b">
                        Columns
                      </div>
                      <ul className="text-sm divide-y">
                        {table.columns.map((col, i) => (
                          <li
                            key={i}
                            className="px-3 py-2 flex items-center hover:bg-gray-50"
                          >
                            <div className="flex-1 flex items-center gap-1.5">
                              {getColumnIcon(col)}
                              <TooltipProvider>
                                <Tooltip>
                                  <TooltipTrigger>
                                    <span
                                      className={
                                        col.is_primary_key ? "font-medium" : ""
                                      }
                                    >
                                      {col.name}
                                    </span>
                                  </TooltipTrigger>
                                  {getColumnTooltip(col) && (
                                    <TooltipContent>
                                      {getColumnTooltip(col)}
                                    </TooltipContent>
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

                    {/* Foreign Keys */}
                    <div className="border rounded-md mt-2 overflow-hidden">
                      <div className="px-3 py-2 text-xs font-medium uppercase bg-gray-50 text-gray-500 border-b">
                        Foreign Keys
                      </div>
                      <ul className="text-sm divide-y">
                        {table.columns
                          .filter((c) => c.foreign_table.Valid)
                          .map((c, i) => (
                            <li
                              key={i}
                              className="px-3 py-2 flex items-center justify-between hover:bg-gray-50"
                            >
                              <div className="flex items-center gap-1">
                                <Link size={14} className="text-green-600" />
                                <span>
                                  {c.name} ⟶ {c.foreign_table.String}.
                                  {c.foreign_column.String}
                                </span>
                              </div>
                            </li>
                          ))}
                        {table.columns.filter((c) => c.foreign_table.Valid)
                          .length === 0 && (
                          <li className="px-3 py-2 text-xs text-gray-500">
                            None
                          </li>
                        )}
                      </ul>
                    </div>
                  </CollapsibleContent>
                )}
              </Collapsible>
            );
          })
        )}
      </ScrollArea>
    </aside>
  );
}
