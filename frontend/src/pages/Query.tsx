import React, { useRef, useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Loader2,
  Check,
  Code,
  AlertTriangle,
  Database,
  PlayCircle,
} from "lucide-react";
import SchemaSidebar from "@/components/SchemaSidebar";
import axios from "axios";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useNavigate } from "react-router-dom";

// Define types for clarity and TypeScript compatibility
interface DBConfig {
  host: string;
  port: string;
  user: string;
  pass: string;
  dbname: string;
}

interface ConfirmationDialog {
  open: boolean;
  sql: string;
  pendingQuery: string;
}

interface ResultItem {
  type: "user" | "assistant";
  content?: any[];
  responseType?: "success" | "error";
  message?: string;
  sql?: string;
  sqlType?: string;
  affectedRows?: number;
  isQAResponse?: boolean;
  extractedSql?: string | null;
}

interface TableCellProps {
  content: any;
  isQAColumn: boolean;
}

interface ResultTableProps {
  data: any[];
  isQAResponse?: boolean;
  extractedSql?: string | null;
  messageIndex: number;
}

export default function Query() {
  const navigate = useNavigate();
  const [showSidebar, setShowSidebar] = useState<boolean>(true);
  const [query, setQuery] = useState<string>("");
  const [results, setResults] = useState<ResultItem[]>([]);
  const [sqlCode, setSqlCode] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(false);
  const [activeCodeIndex, setActiveCodeIndex] = useState<number | null>(null);
  const chatContainerRef = useRef<HTMLDivElement | null>(null);
  const [shouldReRender, setShouldReRender] = useState<boolean>(false);

  // Add state for session ID
  const [sessionId, setSessionId] = useState<string>("");

  // New state for confirmation dialog
  const [confirmationDialog, setConfirmationDialog] =
    useState<ConfirmationDialog>({
      open: false,
      sql: "",
      pendingQuery: "",
    });

  const dbConfig = JSON.parse(
    localStorage.getItem("dbConfig") || "null"
  ) as DBConfig | null;
  if (!dbConfig || !dbConfig.dbname) {
    navigate("/");
    return null;
  }

  // Initialize session ID on component mount
  useEffect(() => {
    // Generate a session ID based on DB connection details
    const generatedSessionId = `${dbConfig.dbname}-${Date.now()}`;

    // Check if there's a stored session ID for this database
    const storedSessionId = localStorage.getItem(
      `sessionId-${dbConfig.dbname}`
    );

    // Use stored session ID if available, otherwise use the new one
    const activeSessionId = storedSessionId || generatedSessionId;

    // Store the session ID
    if (!storedSessionId) {
      localStorage.setItem(`sessionId-${dbConfig.dbname}`, activeSessionId);
    }

    setSessionId(activeSessionId);
  }, [dbConfig.dbname]);

  // Function to extract SQL from QA output's SELECT statement
  const extractSqlFromQaOutput = (output: string): string | null => {
    if (!output) return null;

    // Check for pattern: SELECT '...' AS output
    const selectMatch = output.match(/SELECT\s+'(.+?)'\s+AS\s+output/i);
    if (selectMatch && selectMatch[1]) {
      // Unescape single quotes in the SQL
      return selectMatch[1].replace(/''/g, "'");
    }
    return null;
  };

  // Function to execute extracted SQL
  const executeExtractedSql = async (sql: string): Promise<void> => {
    try {
      setLoading(true);

      const response = await axios.post("http://localhost:8080/query", {
        config: dbConfig,
        prompt: "Execute this SQL directly",
        confirmed: true,
        sqlToConfirm: sql,
        sessionId: sessionId,
      });

      const data = response.data;

      // Update session ID if provided by the backend
      if (data.session_id) {
        setSessionId(data.session_id);
        localStorage.setItem(`sessionId-${dbConfig.dbname}`, data.session_id);
      }

      // Create results entry
      let resultContent: any[] = [];
      let resultMessage = "";

      if (data.affected !== undefined) {
        // For modification queries with affected rows info
        setShouldReRender(!shouldReRender);
        resultMessage = `Operation completed. ${data.affected} rows affected.`;
      } else if (data.result_table && data.result_table.length > 0) {
        // For SELECT queries with data
        resultContent = data.result_table;
        resultMessage = `Query returned ${data.result_table.length} results`;
      } else {
        // Fallback for other cases
        resultContent = data.result_table || [];
        resultMessage = "Query executed successfully";
      }

      setResults((prevResults) => [
        ...prevResults,
        {
          type: "assistant",
          responseType: "success",
          content: resultContent,
          sql: data.sql,
          message: resultMessage,
          sqlType: data.sql_type,
          affectedRows: data.affected,
          isQAResponse: false,
        },
      ]);
    } catch (err: any) {
      const errorMessage =
        err.response?.data?.error ||
        "An error occurred while executing the extracted SQL";

      setResults((prevResults) => [
        ...prevResults,
        {
          type: "assistant",
          responseType: "error",
          message: errorMessage,
          sql: err.response?.data?.sql || "",
        },
      ]);
    } finally {
      setLoading(false);

      setTimeout(() => {
        chatContainerRef.current?.scrollTo({
          top: chatContainerRef.current.scrollHeight,
          behavior: "smooth",
        });
      }, 100);
    }
  };

  const sendQuery = async (confirmed = false): Promise<void> => {
    if (!query.trim() && !confirmed) return;

    setLoading(true);

    // Use the pending query from confirmation dialog if confirmed is true
    const queryToSend = confirmed ? confirmationDialog.pendingQuery : query;

    try {
      // Add the user query to results only when not confirmed
      // This prevents duplicate user messages
      if (!confirmed) {
        setResults((prevResults) => [
          ...prevResults,
          {
            type: "user",
            content: [],
            message: queryToSend,
          },
        ]);
      }

      const response = await axios.post("http://localhost:8080/query", {
        config: dbConfig,
        prompt: queryToSend,
        confirmed: confirmed,
        sqlToConfirm: confirmed ? confirmationDialog.sql : "",
        sessionId: sessionId, // Include the session ID with each request
      });

      const data = response.data;

      // Update session ID if provided by the backend
      if (data.session_id) {
        setSessionId(data.session_id);
        localStorage.setItem(`sessionId-${dbConfig.dbname}`, data.session_id);
      }

      // Handle confirmation request from backend
      if (data.needs_confirmation) {
        setConfirmationDialog({
          open: true,
          sql: data.sql_preview,
          pendingQuery: queryToSend,
        });
        setLoading(false);
        return;
      }

      // Create results entry
      let resultContent: any[] = [];
      let resultMessage = "";

      // Check if this is a QA response - if there's only one row with one column named "output"
      const isQAResponse =
        data.result_table &&
        data.result_table.length === 1 &&
        Object.keys(data.result_table[0]).length === 1 &&
        Object.keys(data.result_table[0])[0] === "output";

      // Special handling for QA responses that contain SQL statements
      let extractedSql = null;
      if (isQAResponse && data.result_table[0].output) {
        extractedSql = extractSqlFromQaOutput(data.result_table[0].output);
      }

      if (data.result_table && data.result_table.length > 0) {
        // For SELECT queries with data
        resultContent = data.result_table;
        resultMessage =
          data.message || `Query returned ${data.result_table.length} results`;
      } else if (data.affected !== undefined) {
        // For modification queries with affected rows info
        setShouldReRender(!shouldReRender);
        resultContent = [];
        resultMessage =
          data.message ||
          `Operation completed. ${data.affected} rows affected.`;
      } else {
        // Fallback for other cases
        resultContent = data.result_table || [];
        resultMessage = data.message || "Query executed successfully";
      }

      setResults((prevResults) => [
        ...prevResults,
        {
          type: "assistant",
          responseType: "success",
          content: resultContent,
          sql: data.sql,
          message: resultMessage,
          sqlType: data.sql_type,
          affectedRows: data.affected,
          isQAResponse: isQAResponse,
          extractedSql: extractedSql,
        },
      ]);

      setSqlCode(data.sql);
      setQuery("");
    } catch (err: any) {
      const errorMessage =
        err.response?.data?.error ||
        "An error occurred while processing your query";

      setResults((prevResults) => [
        ...prevResults,
        {
          type: "assistant",
          responseType: "error",
          message: errorMessage,
          sql: err.response?.data?.sql || "",
        },
      ]);
    } finally {
      setLoading(false);

      setTimeout(() => {
        chatContainerRef.current?.scrollTo({
          top: chatContainerRef.current.scrollHeight,
          behavior: "smooth",
        });
      }, 100);
    }
  };

  const confirmAndSendQuery = (): void => {
    setConfirmationDialog((prev) => ({ ...prev, open: false }));
    sendQuery(true);
  };

  const cancelQuery = (): void => {
    setConfirmationDialog({ open: false, sql: "", pendingQuery: "" });
    setLoading(false);
  };

  // Add function to reset conversation history
  const resetConversation = (): void => {
    // Clear the current session ID in localStorage
    localStorage.removeItem(`sessionId-${dbConfig.dbname}`);

    // Generate a new session ID
    const newSessionId = `${dbConfig.dbname}-${Date.now()}`;
    localStorage.setItem(`sessionId-${dbConfig.dbname}`, newSessionId);
    setSessionId(newSessionId);

    // Clear the conversation UI
    setResults([]);

    // Show confirmation
    alert("Conversation history has been reset.");
  };

  // Truncate long text and show in tooltip for database content
  // But display full text for QA responses
  const TruncatedCell: React.FC<TableCellProps> = ({ content, isQAColumn }) => {
    const contentStr = content === null ? "NULL" : String(content);

    // Always show full text for QA responses in the "output" column
    if (isQAColumn) {
      return content === null ? (
        <span className="text-gray-400">NULL</span>
      ) : (
        <div className="whitespace-pre-wrap">{contentStr}</div>
      );
    }

    // For non-QA data, truncate as before
    const isLong = contentStr.length > 20;

    if (!isLong) {
      return content === null ? (
        <span className="text-gray-400">NULL</span>
      ) : (
        <>{contentStr}</>
      );
    }

    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="cursor-help">
              {contentStr.substring(0, 20)}...
            </span>
          </TooltipTrigger>
          <TooltipContent side="top" className="max-w-md">
            <p className="break-words">{contentStr}</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    );
  };

  // Render a data table from the results using shadcn/ui Table component
  const ResultTable: React.FC<ResultTableProps> = ({
    data,
    isQAResponse = false,
    extractedSql = null,
    messageIndex,
  }) => {
    if (!data || data.length === 0)
      return <div className="text-gray-500">No results found</div>;

    const columns = Object.keys(data[0]);

    // For QA responses with embedded SQL, show a special UI
    if (isQAResponse && columns.includes("output") && extractedSql) {
      return (
        <div className="w-full border rounded bg-white">
          <div className="p-4 border-b">
            <div className="prose max-w-none mb-2">
              <pre className="bg-gray-100 p-3 rounded font-mono text-sm overflow-x-auto">
                {extractedSql}
              </pre>
            </div>
            <div className="flex justify-end mt-2">
              <Button
                onClick={() => executeExtractedSql(extractedSql)}
                className="bg-green-600 hover:bg-green-700"
              >
                <PlayCircle className="h-4 w-4 mr-2" />
                Execute SQL
              </Button>
            </div>
          </div>
        </div>
      );
    }

    // For QA responses, render a special version with full text
    else if (isQAResponse && columns.includes("output")) {
      return (
        <div className="w-full p-4 border rounded bg-white">
          <div className="prose max-w-none">{data[0].output}</div>
        </div>
      );
    }

    // For regular database results, render the table
    return (
      <div className="w-full overflow-auto max-h-96 border rounded">
        <Table>
          <TableHeader>
            <TableRow>
              {columns.map((column, i) => (
                <TableHead key={i} className="bg-gray-50">
                  {column}
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.map((row, rowIndex) => (
              <TableRow key={rowIndex}>
                {columns.map((column, colIndex) => (
                  <TableCell key={colIndex} className="max-w-xs">
                    <TruncatedCell
                      content={row[column]}
                      isQAColumn={isQAResponse && column === "output"}
                    />
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    );
  };

  const toggleCodeView = (index: number): void => {
    if (activeCodeIndex === index) {
      setActiveCodeIndex(null); // Hide code if already showing
    } else {
      setActiveCodeIndex(index); // Show code for this message
    }
  };

  return (
    <div className="flex h-screen">
      {showSidebar && (
        <div className="w-[25rem]">
          <SchemaSidebar shouldReRender={shouldReRender} />
        </div>
      )}

      <div
        className={`flex-1 flex flex-col ${
          showSidebar ? "w-[calc(100%-25rem)]" : "w-full"
        }`}
      >
        <header className="border-b p-4 bg-white flex justify-between items-center">
          <h1 className="text-2xl font-bold text-gray-800">NL → SQL Chat</h1>
          <div className="flex gap-2 items-center">
            <Button
              className={`xl:hidden ${
                showSidebar
                  ? "bg-yellow-500 hover:bg-yellow-600"
                  : "bg-blue-500 hover:bg-blue-600"
              } text-white`}
              onClick={() => setShowSidebar(!showSidebar)}
            >
              {showSidebar ? "Hide Schema" : "Show Schema"}
            </Button>
            <Button
              variant="outline"
              onClick={resetConversation}
              className="text-gray-700"
            >
              New Chat
            </Button>
            <Button
              className="bg-gray-700 hover:bg-gray-800 text-white"
              onClick={() => navigate("/select")}
            >
              Change DB
            </Button>
          </div>
        </header>

        <main
          ref={chatContainerRef}
          className="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-100/50"
        >
          {results.length === 0 ? (
            <div className="flex items-center justify-center h-full text-center text-gray-600 text-lg">
              👋 Welcome! Ask something about your database to get started.
            </div>
          ) : (
            results.map((message, i) => (
              <div
                key={i}
                className={`${
                  message.type === "user"
                    ? "w-fit px-4 py-2 rounded-lg bg-blue-600 text-white self-end ml-auto"
                    : "max-w-[95%] px-4 py-3 rounded-lg bg-white text-gray-800 shadow"
                }`}
              >
                {message.type === "user" ? (
                  message.message || ""
                ) : message.responseType === "error" ? (
                  <div>
                    <Alert className="mb-3 bg-red-50 border-red-200">
                      <AlertTriangle className="h-4 w-4 text-red-500" />
                      <AlertDescription className="text-red-700">
                        {message.message}
                      </AlertDescription>
                    </Alert>

                    {message.sql && (
                      <div className="mb-3 relative">
                        <h4 className="font-medium text-gray-700 mb-2">
                          Failed SQL:
                        </h4>
                        <div className="relative">
                          <pre className="bg-gray-100 p-3 rounded font-mono text-sm overflow-x-auto pr-10">
                            {message.sql}
                          </pre>
                        </div>
                      </div>
                    )}
                  </div>
                ) : (
                  <div>
                    <div className="flex justify-between items-center mb-2">
                      <h4 className="font-medium text-gray-700">Results:</h4>
                      {message.sql && (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 px-2 text-gray-600"
                          onClick={() => toggleCodeView(i)}
                        >
                          <Code size={16} className="mr-1" />
                          {activeCodeIndex === i ? "Hide SQL" : "Show SQL"}
                        </Button>
                      )}
                    </div>

                    {activeCodeIndex === i && message.sql && (
                      <div className="mb-3 relative">
                        <div className="relative">
                          <pre className="bg-gray-100 p-3 rounded font-mono text-sm overflow-x-auto pr-10">
                            {message.sql}
                          </pre>
                        </div>
                      </div>
                    )}

                    {/* Display operation result message for modification queries */}
                    {message.sqlType && message.affectedRows !== undefined && (
                      <Alert className="mb-3 bg-blue-50 border-blue-200">
                        <Database className="h-4 w-4 text-blue-500" />
                        <AlertDescription>{message.message}</AlertDescription>
                      </Alert>
                    )}

                    {/* Display general success message if no specific type is given */}
                    {!message.sqlType &&
                      message.message &&
                      !message.isQAResponse && (
                        <Alert className="mb-3 bg-green-50 border-green-200">
                          <Check className="h-4 w-4 text-green-500" />
                          <AlertDescription className="text-green-700">
                            {message.message}
                          </AlertDescription>
                        </Alert>
                      )}

                    {message.content && message.content.length > 0 && (
                      <ResultTable
                        data={message.content}
                        isQAResponse={message.isQAResponse}
                        extractedSql={message.extractedSql}
                        messageIndex={i}
                      />
                    )}
                  </div>
                )}
              </div>
            ))
          )}
        </main>

        <footer className="p-4 border-t bg-white">
          <form
            onSubmit={(e) => {
              e.preventDefault();
              sendQuery();
            }}
            className="flex gap-2"
          >
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Type your question about the data..."
              className="flex-1"
            />
            <Button type="submit" disabled={loading}>
              {loading ? (
                <Loader2 className="animate-spin mr-2" size={16} />
              ) : null}{" "}
              Send
            </Button>
          </form>
        </footer>
      </div>

      {/* SQL Confirmation Dialog */}
      <Dialog
        open={confirmationDialog.open}
        onOpenChange={(open) => {
          if (!open) cancelQuery();
          setConfirmationDialog((prev) => ({ ...prev, open }));
        }}
      >
        <DialogContent className="sm:max-w-xl">
          <DialogHeader>
            <DialogTitle className="flex items-center text-amber-600">
              <AlertTriangle className="h-5 w-5 mr-2" /> Confirmation Required
            </DialogTitle>
            <DialogDescription>
              You are about to execute a query that will modify your database.
              Please review the SQL before proceeding.
            </DialogDescription>
          </DialogHeader>

          <div className="bg-amber-50 p-4 rounded border border-amber-200 my-4 overflow-auto">
            <h4 className="font-semibold text-amber-800 mb-2">SQL Query:</h4>
            <pre className="bg-white p-3 rounded font-mono text-sm overflow-x-auto border border-amber-100">
              {confirmationDialog.sql}
            </pre>
          </div>

          <DialogFooter className="flex justify-between sm:justify-between">
            <Button
              type="button"
              variant="outline"
              onClick={cancelQuery}
              className="mt-2 sm:mt-0"
            >
              Cancel
            </Button>
            <Button
              type="button"
              className="bg-amber-600 hover:bg-amber-700 mt-2 sm:mt-0"
              onClick={confirmAndSendQuery}
            >
              Confirm and Execute
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
