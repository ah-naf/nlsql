// components/SchemaSidebar.tsx
import { useVirtualizer, VirtualItem } from "@tanstack/react-virtual";
import { useEffect, useState, useRef, useMemo, useCallback } from "react";
import axios from "axios";
import { Database, Loader } from "lucide-react";
import { TableSearch } from "./schema/TableSearch";
import { TableItem } from "./schema/TableItem";
import { BriefTable, DatabaseSchema, TableInfo } from "@/types/schema";
import { EmptyState, ErrorState } from "./schema/utils";

type SchemaSidebarProps = {
  shouldReRender: boolean;
};

export default function SchemaSidebar({ shouldReRender }: SchemaSidebarProps) {
  const dbConfig = JSON.parse(localStorage.getItem("dbConfig") || "null");
  const [briefTables, setBriefTables] = useState<BriefTable[]>([]);
  const [fullSchema, setFullSchema] = useState<DatabaseSchema>({});
  const [openTable, setOpenTable] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState<string>("");
  const [schemaError, setSchemaError] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [loadingTables, setLoadingTables] = useState<Record<string, boolean>>(
    {}
  );
  const parentRef = useRef<HTMLDivElement>(null);
  const itemHeights = useRef<Record<string, number>>({});

  useEffect(() => {
    const fetchBriefSchema = async () => {
      setIsLoading(true);
      try {
        const res = await axios.get<{ tables: BriefTable[] }>(
          "http://localhost:8080/schema",
          { params: { ...dbConfig, brief: true } }
        );
        setBriefTables(res.data.tables || []);
        setError(null);
      } catch (e: any) {
        setError(e.response?.data?.error || e.message);
      } finally {
        setIsLoading(false);
      }
    };

    fetchBriefSchema();
  }, [shouldReRender]);

  const filtered = useMemo(() => {
    return briefTables.filter((t) =>
      t.name.toLowerCase().includes(search.trim().toLowerCase())
    );
  }, [briefTables, search]);

  // Create virtualizer with dynamic size estimation
  const virtualizer = useVirtualizer({
    count: filtered.length,
    getScrollElement: () => parentRef.current,
    estimateSize: (index: number) => {
      const table = filtered[index];
      return table ? itemHeights.current[table.name] || 60 : 60;
    },
    overscan: 5,
    getItemKey: (index: number) => filtered[index]?.name || `item-${index}`,
  });

  // Handle measurement of expanded items
  const measureElement = useCallback(
    (element: HTMLElement | null, tableName: string) => {
      if (element) {
        const height = element.getBoundingClientRect().height;
        if (itemHeights.current[tableName] !== height) {
          itemHeights.current[tableName] = height;
          virtualizer.measure();
        }
      }
    },
    [virtualizer]
  );

  const handleToggle = async (tableName: string, isOpen: boolean) => {
    if (isOpen) {
      setOpenTable(tableName);
      if (!fullSchema[tableName]) {
        try {
          setSchemaError("");
          setLoadingTables((prev) => ({ ...prev, [tableName]: true }));

          const res = await axios.get<{ table: TableInfo }>(
            `http://localhost:8080/schema/${tableName}`,
            { params: dbConfig }
          );
          setFullSchema((fs) => ({ ...fs, [tableName]: res.data.table }));
          virtualizer.measure();
        } catch (e: any) {
          console.error("Failed to load table:", e);
          setSchemaError(e.response?.data?.error || e.message);
        } finally {
          setLoadingTables((prev) => ({ ...prev, [tableName]: false }));
        }
      }
    } else {
      setOpenTable(null);

      itemHeights.current[tableName] = 60;
      virtualizer.measure();
    }
  };

  const TableLoadingSkeleton = () => (
    <div className="animate-pulse px-2 py-2">
      <div className="space-y-2 py-5 bg-gray-200 rounded w-full"></div>
    </div>
  );

  return (
    <aside className="w-96 flex flex-col h-full border-r bg-white shadow">
      <div className="p-4 py-5 border-b bg-gray-50">
        <h2 className="text-xl font-bold text-gray-800 flex items-center gap-2">
          <Database size={20} className="text-indigo-600" />
          {dbConfig?.dbname || "Database Schema"}
        </h2>
        <TableSearch search={search} setSearch={setSearch} />
      </div>

      <div ref={parentRef} className="flex-1 overflow-auto px-2">
        {isLoading ? (
          <div className="py-2">
            {[1, 2, 3, 4, 5].map((i) => (
              <TableLoadingSkeleton key={i} />
            ))}
          </div>
        ) : error ? (
          <ErrorState error={error} />
        ) : filtered.length === 0 ? (
          <EmptyState />
        ) : (
          <div
            className="relative w-full"
            style={{ height: `${virtualizer.getTotalSize()}px` }}
          >
            {virtualizer.getVirtualItems().map((virtualItem: VirtualItem) => {
              const table = filtered[virtualItem.index];
              const isOpen = openTable === table.name;
              const tableData = fullSchema[table.name];
              const isTableLoading = loadingTables[table.name];

              return (
                <div
                  key={virtualItem.key}
                  data-index={virtualItem.index}
                  ref={(el) => measureElement(el, table.name)}
                  className="absolute top-0 left-0 w-full"
                  style={{
                    transform: `translateY(${virtualItem.start}px)`,
                  }}
                >
                  <div className="py-1">
                    <TableItem
                      table={table}
                      isOpen={isOpen}
                      tableData={tableData}
                      schemaError={schemaError}
                      onToggle={handleToggle}
                      isLoading={isTableLoading}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </aside>
  );
}
