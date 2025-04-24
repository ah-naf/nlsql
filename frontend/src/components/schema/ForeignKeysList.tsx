import { ColumnInfo } from "@/types/schema";
import { Link } from "lucide-react";

type ForeignKeysListProps = {
  columns: ColumnInfo[];
};

export function ForeignKeysList({ columns }: ForeignKeysListProps) {
  const foreignKeyColumns = columns.filter((c) => c.foreign_table.Valid);

  return (
    <div className="border rounded-md mt-2 overflow-hidden">
      <div className="px-3 py-2 text-xs font-medium uppercase bg-gray-50 text-gray-500 border-b">
        Foreign Keys
      </div>
      <ul className="text-sm divide-y">
        {foreignKeyColumns.length > 0 ? (
          foreignKeyColumns.map((c, i) => (
            <li
              key={i}
              className="px-3 py-2 flex items-center justify-between hover:bg-gray-50"
            >
              <div className="flex items-center gap-1">
                <Link size={14} className="text-green-600" />
                <span>
                  {c.name} ⟶ {c.foreign_table.String}.{c.foreign_column.String}
                </span>
              </div>
            </li>
          ))
        ) : (
          <li className="px-3 py-2 text-xs text-gray-500">None</li>
        )}
      </ul>
    </div>
  );
}
