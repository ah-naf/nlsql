import React from "react";
import { Card } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { ChevronDown, Table } from "lucide-react";

export default function SchemaSidebar({ schema }: { schema: any }) {
  return (
    <aside className="w-[25rem] hidden lg:flex flex-col border-r bg-white">
      <div className="p-4 border-b">
        <h2 className="text-xl font-bold text-gray-800 flex items-center gap-2">
          Schema
        </h2>
      </div>
      <ScrollArea className="flex-1 px-3 pt-2 pb-20">
        {Object.entries(schema).map(([tableName, columns]) => (
          <Card key={tableName} className="mb-4">
            <div className="flex items-center justify-between p-3 cursor-pointer">
              <div className="flex items-center gap-2">
                <Table size={16} className="text-indigo-600" />
                <span className="font-semibold text-gray-800">{tableName}</span>
                <span className="text-xs text-gray-600">({columns.length})</span>
              </div>
              <ChevronDown className="text-gray-400" size={16} />
            </div>
            <div className="px-3 pb-3">
              <ul className="text-sm text-gray-700 space-y-1">
                {columns.map((col: any, idx: number) => (
                  <li key={idx} className="flex justify-between">
                    <span>{col.name}</span>
                    <span className="text-xs text-gray-500">{col.data_type}</span>
                  </li>
                ))}
              </ul>
            </div>
          </Card>
        ))}
      </ScrollArea>
    </aside>
  );
}