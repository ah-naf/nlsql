// components/schema/EmptyState.tsx
import { Database } from "lucide-react";

export function EmptyState() {
  return (
    <div className="mt-20 px-6 text-center text-gray-500">
      <Database size={48} className="mx-auto mb-4 text-indigo-300" />
      <p className="font-medium">No tables found.</p>
    </div>
  );
}

// components/schema/ErrorState.tsx
import { AlertCircle } from "lucide-react";

type ErrorStateProps = {
  error: string;
};

export function ErrorState({ error }: ErrorStateProps) {
  return (
    <div className="mt-20 px-6 text-center text-red-500">
      <AlertCircle size={48} className="mx-auto mb-4 text-red-400" />
      <p className="font-semibold">{error}</p>
      <p className="text-xs text-gray-500 mt-1">
        Check your connection or credentials.
      </p>
    </div>
  );
}
