import React, { useRef, useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Loader2, Check, Code, AlertTriangle, Database } from "lucide-react";
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

export default function Query() {
  const navigate = useNavigate();
  const [showSidebar, setShowSidebar] = useState(true);
  const [query, setQuery] = useState("");
  const [results, setResults] = useState([]);
  const [sqlCode, setSqlCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [activeCodeIndex, setActiveCodeIndex] = useState(null);
  const chatContainerRef = useRef(null);
  const [shouldReRender, setShouldReRender] = useState(false);

  // New state for confirmation dialog
  const [confirmationDialog, setConfirmationDialog] = useState({
    open: false,
    sql: "",
    pendingQuery: "",
  });

  const dbConfig = JSON.parse(localStorage.getItem("dbConfig") || "null");
  if (!dbConfig || !dbConfig.dbname) {
    navigate("/");
    return null;
  }

  const sendQuery = async (confirmed = false) => {
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
            content: queryToSend,
          },
        ]);
      }

      const response = await axios.post("http://localhost:8080/query", {
        config: dbConfig,
        prompt: queryToSend,
        confirmed: confirmed,
        sqlToConfirm: confirmed ? confirmationDialog.sql : "",
      });

      const data = response.data;

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
      let resultContent;
      let resultMessage = "";

      // Check if this is a QA response - if there's only one row with one column named "output"
      const isQAResponse =
        data.result_table &&
        data.result_table.length === 1 &&
        Object.keys(data.result_table[0]).length === 1 &&
        Object.keys(data.result_table[0])[0] === "output";

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
        },
      ]);

      setSqlCode(data.sql);
      setQuery("");
    } catch (err) {
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

  const confirmAndSendQuery = () => {
    setConfirmationDialog((prev) => ({ ...prev, open: false }));
    sendQuery(true);
  };

  const cancelQuery = () => {
    setConfirmationDialog({ open: false, sql: "", pendingQuery: "" });
    setLoading(false);
  };

  // Truncate long text and show in tooltip for database content
  // But display full text for QA responses
  const TruncatedCell = ({ content, isQAColumn }) => {
    const contentStr = String(content);

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
        contentStr
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
  const ResultTable = ({ data, isQAResponse }) => {
    if (!data || data.length === 0)
      return <div className="text-gray-500">No results found</div>;

    const columns = Object.keys(data[0]);

    // For QA responses, render a special version with full text
    if (isQAResponse && columns.includes("output")) {
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

  const toggleCodeView = (index) => {
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
              className="bg-gray-700 hover:bg-gray-800 text-white"
              onClick={() => navigate("/")}
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
                  message.content
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
