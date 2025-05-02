// src/components/QueryInput.tsx
import React from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";

export type Mode = "nl" | "sql";

interface QueryInputProps {
  query: string;
  setQuery: (q: string) => void;
  loading: boolean;
  onSubmit: (mode: Mode) => void;
}

export default function QueryInput({
  query,
  setQuery,
  loading,
  onSubmit,
}: QueryInputProps) {
  const handleNl = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit("nl");
  };
  const handleSql = () => {
    if (!loading && query.trim()) onSubmit("sql");
  };

  return (
    <footer className="p-4 border-t bg-white">
      <form onSubmit={handleNl} className="flex gap-2">
        <Input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Type question or SQL..."
          className="flex-1"
        />

        <Button type="submit" disabled={loading}>
          {loading && <Loader2 className="animate-spin mr-2" size={16} />}
          Send
        </Button>
        <Button
          type="button"
          variant="outline"
          disabled={loading || !query.trim()}
          onClick={handleSql}
        >
          {loading && <Loader2 className="animate-spin mr-2" size={16} />}
          Run SQL
        </Button>
      </form>
      <p className="text-center text-xs mt-2 text-gray-700 font-medium">
        “Send” interprets your text; “Run SQL” executes it directly.
      </p>
    </footer>
  );
}
