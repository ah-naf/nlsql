import { Dispatch, SetStateAction } from "react";

type TableSearchProps = {
  search: string;
  setSearch: Dispatch<SetStateAction<string>>;
};

export function TableSearch({ search, setSearch }: TableSearchProps) {
  return (
    <input
      type="search"
      placeholder="🔍 Search tables..."
      className="mt-3 w-full px-2 py-1 border rounded text-sm"
      value={search}
      onChange={(e) => setSearch(e.target.value)}
    />
  );
}
