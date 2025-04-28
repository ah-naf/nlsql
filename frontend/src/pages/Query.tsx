// src/pages/Query.tsx
import { useRef, useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import SchemaSidebar from "@/components/SchemaSidebar";
import QueryHeader from "@/components/QueryHeader";
import QueryInput from "@/components/QueryInput";
import ChatContainer from "@/components/ChatContainer";
import SqlConfirmationDialog from "@/components/SqlConfirmationDialog";
import { DBConfig, ConfirmationDialog, ResultItem } from "../types/query";
import {
  extractSqlFromQaOutput,
  sendQueryToBackend,
  getSessionId,
  resetSessionId,
} from "../utils/dbUtils";

export default function Query() {
  const navigate = useNavigate();
  const [showSidebar, setShowSidebar] = useState<boolean>(true);
  const [query, setQuery] = useState<string>("");
  const [results, setResults] = useState<ResultItem[]>([]);
  // const [sqlCode, setSqlCode] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(false);
  const [activeCodeIndex, setActiveCodeIndex] = useState<number | null>(null);
  const chatContainerRef = useRef<HTMLDivElement | null>(null);
  const [shouldReRender, setShouldReRender] = useState<boolean>(false);
  const [sessionId, setSessionId] = useState<string>("");
  const [confirmationDialog, setConfirmationDialog] =
    useState<ConfirmationDialog>({
      open: false,
      sql: "",
      pendingQuery: "",
    });

  // Load database configuration
  const dbConfig = JSON.parse(
    localStorage.getItem("dbConfig") || "null"
  ) as DBConfig | null;

  // Redirect if no database config
  if (!dbConfig || !dbConfig.dbname) {
    navigate("/");
    return null;
  }

  // Initialize session ID on component mount
  useEffect(() => {
    const activeSessionId = getSessionId(dbConfig.dbname);
    setSessionId(activeSessionId);
  }, [dbConfig.dbname]);

  // Function to toggle code view
  const toggleCodeView = (index: number): void => {
    if (activeCodeIndex === index) {
      setActiveCodeIndex(null); // Hide code if already showing
    } else {
      setActiveCodeIndex(index); // Show code for this message
    }
  };

  // Execute extracted SQL from QA response
  const executeExtractedSql = async (sql: string): Promise<void> => {
    try {
      setLoading(true);

      const response = await sendQueryToBackend(
        dbConfig,
        "Execute this SQL directly",
        true,
        sql,
        sessionId
      );

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
      scrollToBottom();
    }
  };

  // Scroll chat to bottom
  const scrollToBottom = () => {
    setTimeout(() => {
      chatContainerRef.current?.scrollTo({
        top: chatContainerRef.current.scrollHeight,
        behavior: "smooth",
      });
    }, 100);
  };

  // Send query to backend
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

      const response = await sendQueryToBackend(
        dbConfig,
        queryToSend,
        confirmed,
        confirmed ? confirmationDialog.sql : "",
        sessionId
      );

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

      // Check if this is a QA response
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
        console.log("first");
        // For SELECT queries with data
        resultContent = data.result_table;
        resultMessage =
          data.message || `Query returned ${data.result_table.length} results`;
      } else if (data.affected !== undefined) {
        console.log("second");
        // For modification queries with affected rows info
        setShouldReRender(!shouldReRender);
        resultContent = [];
        resultMessage =
          data.message ||
          `Operation completed. ${data.affected} rows affected.`;
      } else {
        setShouldReRender(!shouldReRender);
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

      // setSqlCode(data.sql);
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
      scrollToBottom();
      setQuery("");
    }
  };

  // Confirm and execute SQL
  const confirmAndSendQuery = (): void => {
    setConfirmationDialog((prev) => ({ ...prev, open: false }));
    sendQuery(true);
  };

  // Cancel query
  const cancelQuery = (): void => {
    setConfirmationDialog({ open: false, sql: "", pendingQuery: "" });
    setLoading(false);
  };

  // Reset conversation
  const resetConversation = (): void => {
    // Reset the session ID
    const newSessionId = resetSessionId(dbConfig.dbname);
    setSessionId(newSessionId);

    // Clear the conversation UI
    setResults([]);

    // Show confirmation
    alert("Conversation history has been reset.");
  };

  return (
    <div className="flex h-screen">
      <div className={`w-[25rem] ${showSidebar ? "block" : "hidden"}`}>
        <SchemaSidebar shouldReRender={shouldReRender} />
      </div>

      <div
        className={`flex-1 flex flex-col ${
          showSidebar ? "w-[calc(100%-25rem)]" : "w-full"
        }`}
      >
        <QueryHeader
          showSidebar={showSidebar}
          setShowSidebar={setShowSidebar}
          resetConversation={resetConversation}
        />

        <ChatContainer
          chatContainerRef={chatContainerRef}
          results={results}
          activeCodeIndex={activeCodeIndex}
          toggleCodeView={toggleCodeView}
          onExecuteSql={executeExtractedSql}
          loading={loading}
        />

        <QueryInput
          query={query}
          setQuery={setQuery}
          loading={loading}
          onSubmit={() => sendQuery()}
        />
      </div>

      <SqlConfirmationDialog
        confirmationDialog={confirmationDialog}
        onConfirm={confirmAndSendQuery}
        onCancel={cancelQuery}
      />
    </div>
  );
}
