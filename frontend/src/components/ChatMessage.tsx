import { Code, AlertTriangle, Database, Check } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import ResultTable from "./ResultTable";
import { ResultItem } from "../types/query";

interface ChatMessageProps {
  message: ResultItem;
  index: number;
  activeCodeIndex: number | null;
  toggleCodeView: (index: number) => void;
  onExecuteSql: (sql: string) => void;
}

export default function ChatMessage({
  message,
  index,
  activeCodeIndex,
  toggleCodeView,
  onExecuteSql,
}: ChatMessageProps) {
  // User message
  if (message.type === "user") {
    return (
      <div className="w-fit px-4 py-2 rounded-lg bg-blue-600 text-white self-end ml-auto">
        {message.message || ""}
      </div>
    );
  }

  // Error message
  if (message.responseType === "error") {
    return (
      <div className="max-w-[95%] px-4 py-3 rounded-lg bg-white text-gray-800 shadow">
        <Alert className="mb-3 bg-red-50 border-red-200">
          <AlertTriangle className="h-4 w-4 text-red-500" />
          <AlertDescription className="text-red-700">
            {message.message}
          </AlertDescription>
        </Alert>

        {message.sql && (
          <div className="mb-3 relative">
            <h4 className="font-medium text-gray-700 mb-2">Failed SQL:</h4>
            <div className="relative">
              <pre className="bg-gray-100 p-3 rounded font-mono text-sm overflow-x-auto pr-10">
                {message.sql}
              </pre>
            </div>
          </div>
        )}
      </div>
    );
  }

  // Success message
  return (
    <div className="max-w-[95%] px-4 py-3 rounded-lg bg-white text-gray-800 shadow">
      <div className="flex justify-between items-center mb-2">
        <h4 className="font-medium text-gray-700">Results:</h4>
        {message.sql && (
          <Button
            variant="ghost"
            size="sm"
            className="h-8 px-2 text-gray-600"
            onClick={() => toggleCodeView(index)}
          >
            <Code size={16} className="mr-1" />
            {activeCodeIndex === index ? "Hide SQL" : "Show SQL"}
          </Button>
        )}
      </div>

      {activeCodeIndex === index && message.sql && (
        <div className="mb-3 relative">
          <div className="relative">
            <pre className="bg-gray-100 p-3 rounded font-mono text-sm overflow-x-auto pr-10">
              {message.sql}
            </pre>
          </div>
        </div>
      )}

      <Alert className="mb-3 bg-blue-50 border-blue-200">
        <Database className="h-4 w-4 text-blue-500" />
        <AlertDescription>{message.message}</AlertDescription>
      </Alert>

      {message.content && message.content.length > 0 && (
        <ResultTable
          data={message.content}
          isQAResponse={message.isQAResponse}
          extractedSql={message.extractedSql}
          messageIndex={index}
          onExecuteSql={onExecuteSql}
        />
      )}
    </div>
  );
}
