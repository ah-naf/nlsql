import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Card } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import axios from "axios";
import {
  ChevronDown,
  Table,
  Database,
  AlertCircle,
  ChevronRight,
} from "lucide-react";
import { useEffect, useState } from "react";

type ColumnInfo = {
  name: string;
  data_type: string;
  is_nullable: boolean;
  default_value: string | null;
  is_primary_key: boolean;
  is_unique: boolean;
  foreign_table: string | null;
  foreign_column: string | null;
  description: string | null;
};

type TableInfo = {
  name: string;
  columns: ColumnInfo[];
  description: string | null;
  row_count: number;
};

type DatabaseSchema = {
  [tableName: string]: TableInfo;
};

export default function SchemaSidebar() {
  const [schema, setSchema] = useState<DatabaseSchema>({});
  const [error, setError] = useState<string | null>(null);
  const [openTable, setOpenTable] = useState<string | null>(null);
  const dbConfig = JSON.parse(localStorage.getItem("dbConfig") || "null");

  useEffect(() => {
    const fetchSchema = async () => {
      try {
        const res = await axios.get("http://localhost:8080/schema", {
          params: dbConfig,
        });
        setSchema(res.data.schema || {});
        setError(null);
      } catch (err: any) {
        setError(err.response?.data?.error || "Failed to load schema");
      }
    };

    fetchSchema();
  }, []);

  const isEmpty = !schema || Object.keys(schema).length === 0;

  return (
    <aside className="w-[25rem] hidden h-full shadow lg:flex flex-col border-r bg-white">
      <div className="p-4 py-5 border-b">
        <h2 className="text-xl font-bold text-gray-800 flex items-center gap-2">
          {dbConfig?.dbname || "Database"}
        </h2>
      </div>

      <ScrollArea className="flex-1 px-3 pt-2 pb-20 overflow-y-auto">
        {error ? (
          <div className="flex flex-col items-center justify-center text-center text-red-500 mt-20 px-6">
            <AlertCircle size={48} className="mb-4 text-red-400" />
            <p className="text-sm font-semibold">{error}</p>
            <p className="text-xs text-gray-500 mt-1">
              Please check your connection or database credentials.
            </p>
          </div>
        ) : isEmpty ? (
          <div className="flex flex-col items-center justify-center text-center text-gray-500 mt-20 px-6">
            <Database size={48} className="mb-4 text-indigo-300" />
            <p className="text-sm font-medium">
              No schema available for this database.
            </p>
            <p className="text-xs mt-1">
              You can try running a query first or refresh the schema later.
            </p>
          </div>
        ) : (
          Object.entries(schema).map(([tableName, table]) => {
            const isOpen = openTable === tableName;

            return (
              <Collapsible
                key={tableName}
                open={isOpen}
                onOpenChange={(state) => setOpenTable(state ? tableName : null)}
              >
                <CollapsibleTrigger className="w-full px-3 py-2 flex items-center justify-between border-b bg-white hover:bg-gray-50 cursor-pointer">
                  <div className="flex items-center gap-2">
                    <Table size={16} className="text-indigo-600" />
                    <span className="font-semibold text-gray-800">
                      {tableName}
                    </span>
                    <span className="text-xs text-gray-600">
                      ({table.columns.length})
                    </span>
                  </div>
                  {isOpen ? (
                    <ChevronDown className="text-gray-400" size={16} />
                  ) : (
                    <ChevronRight className="text-gray-400" size={16} />
                  )}
                </CollapsibleTrigger>
                <CollapsibleContent className="px-4 pb-3">
                  <ul className="text-sm text-gray-700 space-y-1 mt-2">
                    {table.columns.map((col, idx) => (
                      <li key={idx} className="flex justify-between">
                        <span>{col.name}</span>
                        <span className="text-xs text-gray-500">
                          {col.data_type}
                        </span>
                      </li>
                    ))}
                  </ul>
                </CollapsibleContent>
              </Collapsible>
            );
          })
        )}
      </ScrollArea>
    </aside>
  );
}
