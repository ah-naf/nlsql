// components/SchemaSidebar.tsx
import { useVirtualizer, VirtualItem } from "@tanstack/react-virtual";
import { useEffect, useState, useRef, useMemo, useCallback } from "react";
import axios from "axios";
import { Database, ChevronUp } from "lucide-react";
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
  const [openTables, setOpenTables] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState<string>("");
  const [schemaError, setSchemaError] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [loadingTables, setLoadingTables] = useState<Record<string, boolean>>(
    {}
  );
  const parentRef = useRef<HTMLDivElement>(null);
  const itemHeights = useRef<Record<string, number>>({});

  // Fetch brief list on mount / shouldReRender
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

  // Filtered by search
  const filtered = useMemo(
    () =>
      briefTables.filter((t) =>
        t.name.toLowerCase().includes(search.trim().toLowerCase())
      ),
    [briefTables, search]
  );

  // Virtualizer setup
  const virtualizer = useVirtualizer({
    count: filtered.length,
    getScrollElement: () => parentRef.current,
    estimateSize: (index) => {
      const table = filtered[index];
      return table ? itemHeights.current[table.name] || 60 : 60;
    },
    overscan: 5,
    getItemKey: (index) => filtered[index]?.name || `item-${index}`,
  });

  // Measure each row when its element mounts / updates
  const measureElement = useCallback(
    (el: HTMLElement | null, tableName: string) => {
      if (!el) return;
      const h = el.getBoundingClientRect().height;
      if (itemHeights.current[tableName] !== h) {
        itemHeights.current[tableName] = h;
        virtualizer.measure();
      }
    },
    [virtualizer]
  );

  // Are any tables loading? if so, disable all toggles
  const anyLoading = useMemo(
    () => Object.values(loadingTables).some(Boolean),
    [loadingTables]
  );

  // Toggle a single table open/closed
  const handleToggle = async (tableName: string, open: boolean) => {
    if (open) {
      setOpenTables((prev) => [...prev, tableName]);
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
          console.error(e);
          setSchemaError(e.response?.data?.error || e.message);
        } finally {
          setLoadingTables((prev) => ({ ...prev, [tableName]: false }));
        }
      }
    } else {
      setOpenTables((prev) => prev.filter((n) => n !== tableName));
      itemHeights.current[tableName] = 60;
      virtualizer.measure();
    }
  };

  // **NEW**: collapse everything in one go
  const collapseAll = () => {
    setOpenTables([]);
    // reset ALL cached heights so we re-estimate at 60px
    itemHeights.current = {};
    virtualizer.measure();
  };

  const TableLoadingSkeleton = () => (
    <div className="animate-pulse px-2 py-2">
      <div className="space-y-2 py-5 bg-gray-200 rounded w-full" />
    </div>
  );

  return (
    <aside className="w-[25rem] flex flex-col h-full border-r bg-white shadow">
      <div className="p-4 py-2 pt-5 border-b bg-gray-50">
        <h2 className="text-xl font-bold text-gray-800 flex items-center gap-2">
          <Database size={20} className="text-indigo-600" />
          {dbConfig?.dbname || "Database Schema"}
        </h2>
        <TableSearch search={search} setSearch={setSearch} />

        <button
          onClick={collapseAll}
          disabled={anyLoading}
          className="mt-2 flex items-center ml-auto gap-1 px-2 py-1 text-xs text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <ChevronUp size={16} />
          Collapse All
        </button>
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
            className="relative w-full mt-2"
            style={{ height: `${virtualizer.getTotalSize()}px` }}
          >
            {virtualizer.getVirtualItems().map((virt: VirtualItem) => {
              const table = filtered[virt.index];
              const isOpen = openTables.includes(table.name);
              const tableData = fullSchema[table.name];
              const loading = loadingTables[table.name];

              return (
                <div
                  key={virt.key}
                  data-index={virt.index}
                  ref={(el) => measureElement(el, table.name)}
                  className="absolute top-0 left-0 w-full"
                  style={{ transform: `translateY(${virt.start}px)` }}
                >
                  <div className="py-1">
                    <TableItem
                      table={table}
                      isOpen={isOpen}
                      tableData={tableData}
                      schemaError={schemaError}
                      onToggle={handleToggle}
                      isLoading={loading}
                      disabled={anyLoading}
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
