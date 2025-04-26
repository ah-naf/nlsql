import React from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";

interface QueryInputProps {
  query: string;
  setQuery: (query: string) => void;
  loading: boolean;
  onSubmit: () => void;
}

export default function QueryInput({
  query,
  setQuery,
  loading,
  onSubmit,
}: QueryInputProps) {
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit();
  };

  return (
    <footer className="p-4 border-t bg-white">
      <form onSubmit={handleSubmit} className="flex gap-2">
        <Input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Type your question about the data..."
          className="flex-1"
        />
        <Button type="submit" disabled={loading}>
          {loading ? <Loader2 className="animate-spin mr-2" size={16} /> : null}{" "}
          Send
        </Button>
      </form>
      <p className="text-center text-xs mt-2 text-gray-700 font-medium">
        NLSQL is wouldn't give you 100% accurate query.
      </p>
    </footer>
  );
}
