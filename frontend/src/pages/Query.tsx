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
  sendSqlToBackend,
  getSessionId,
  resetSessionId,
} from "../utils/dbUtils";
import { AxiosError } from "axios";

type Mode = "nl" | "sql";

export default function Query() {
  const navigate = useNavigate();
  const [showSidebar, setShowSidebar] = useState<boolean>(true);
  const [query, setQuery] = useState<string>("");
  const [results, setResults] = useState<ResultItem[]>([]);
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
      mode: "nl",
    });

  // Load database configuration
  const dbConfig = JSON.parse(
    localStorage.getItem("dbConfig") || "null"
  ) as DBConfig | null;

  // Initialize session ID on component mount
  useEffect(() => {
    if (dbConfig?.dbname) {
      const activeSessionId = getSessionId(dbConfig.dbname);
      setSessionId(activeSessionId);
    }
  }, [dbConfig]);

  if (!dbConfig || !dbConfig.dbname) {
    navigate("/");
    return null;
  }

  // Toggle code view for SQL snippets
  const toggleCodeView = (index: number): void => {
    setActiveCodeIndex((prev) => (prev === index ? null : index));
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

  // Execute a raw SQL string via the /execute-sql endpoint
  const executeExtractedSql = async (sql: string): Promise<void> => {
    try {
      setLoading(true);

      const response = await sendSqlToBackend(dbConfig, sql, true, sessionId);
      const data = response.data;

      if (data.session_id) {
        setSessionId(data.session_id);
        localStorage.setItem(`sessionId-${dbConfig.dbname}`, data.session_id);
      }

      let resultContent: Record<string, unknown>[] = [];
      let resultMessage = "";

      if (data.affected !== undefined) {
        setShouldReRender((prev) => !prev);
        resultMessage = `Operation completed. ${data.affected} rows affected.`;
      } else if (data.result_table?.length) {
        resultContent = data.result_table;
        resultMessage = `Query returned ${data.result_table.length} results`;
      } else {
        resultContent = data.result_table || [];
        resultMessage = "Query executed successfully";
      }

      setResults((prev) => [
        ...prev,
        {
          type: "assistant",
          responseType: "success",
          content: resultContent,
          sql: data.sql,
          message: resultMessage,
          sqlType: data.sql_type,
          affectedRows: data.affected,
          isQAResponse: false,
          isSQL: true,
        },
      ]);
    } catch (err: unknown) {
      const error = err as AxiosError<{ error?: string; sql?: string }>;

      setResults((prev) => [
        ...prev,
        {
          type: "assistant",
          responseType: "error",
          message:
            error.response?.data?.error ||
            "An error occurred while executing the extracted SQL",
          sql: error.response?.data?.sql,
          isSQL: true,
        },
      ]);
    } finally {
      setLoading(false);
      scrollToBottom();
    }
  };

  /**
   * Send either a natural-language query (/query) or raw SQL (/execute-sql).
   * @param mode "nl" for NL→SQL, "sql" for direct SQL
   * @param confirmed whether the user has confirmed a destructive operation
   */
  const sendQuery = async (
    mode: Mode = "nl",
    confirmed = false
  ): Promise<void> => {
    // if no input (and not confirmation retry), bail
    if (!query.trim() && !(confirmed && mode === "sql")) return;

    setLoading(true);
    const textToSend = confirmed
      ? confirmationDialog.pendingQuery
      : query.trim();

    // show the user's message
    if (!confirmed) {
      setResults((prev) => [
        ...prev,
        { type: "user", message: textToSend, isSQL: mode === "sql" },
      ]);
    }

    try {
      const response =
        mode === "nl"
          ? await sendQueryToBackend(
              dbConfig,
              textToSend,
              confirmed,
              confirmed ? confirmationDialog.sql : "",
              sessionId
            )
          : await sendSqlToBackend(dbConfig, textToSend, confirmed, sessionId);

      const data = response.data;

      if (data.session_id) {
        setSessionId(data.session_id);
        localStorage.setItem(`sessionId-${dbConfig.dbname}`, data.session_id);
      }

      if (mode === "nl" && data.needs_confirmation) {
        setConfirmationDialog({
          open: true,
          sql: data.sql_preview,
          pendingQuery: textToSend,
          mode: "nl",
        });
        setLoading(false);
        return;
      }

      let resultContent: Record<string, unknown>[] = [];
      let resultMessage = "";

      const isQAResponse =
        Array.isArray(data.result_table) &&
        data.result_table.length === 1 &&
        Object.keys(data.result_table[0]).length === 1 &&
        Object.keys(data.result_table[0])[0] === "output";

      let extractedSql: string | null = null;
      if (isQAResponse) {
        extractedSql = extractSqlFromQaOutput(
          data.result_table[0].output as string
        );
      }

      if (Array.isArray(data.result_table) && data.result_table.length > 0) {
        resultContent = data.result_table;
        resultMessage =
          data.message || `Query returned ${data.result_table.length} results`;
      } else if (data.affected !== undefined) {
        setShouldReRender((prev) => !prev);
        resultMessage =
          data.message ||
          `Operation completed. ${data.affected} rows affected.`;
      } else {
        resultContent = data.result_table || [];
        resultMessage = data.message || "Query executed successfully";
      }

      setResults((prev) => [
        ...prev,
        {
          type: "assistant",
          responseType: "success",
          content: resultContent,
          sql: data.sql,
          message: resultMessage,
          sqlType: data.sql_type,
          affectedRows: data.affected,
          isQAResponse,
          extractedSql,
          isSQL: mode === "sql",
        },
      ]);

      setQuery("");
    } catch (err: unknown) {
      const error = err as AxiosError<{ error?: string; sql?: string }>;

      setResults((prev) => [
        ...prev,
        {
          type: "assistant",
          responseType: "error",
          message:
            error.response?.data?.error ||
            "An error occurred while processing your query",
          sql: error.response?.data?.sql,
          isSQL: mode === "sql",
        },
      ]);
    } finally {
      setLoading(false);
      scrollToBottom();
    }
  };

  const confirmAndSendQuery = (): void => {
    setConfirmationDialog((prev) => ({ ...prev, open: false }));
    sendQuery(confirmationDialog.mode, true);
  };

  const cancelQuery = (): void => {
    setConfirmationDialog({
      open: false,
      sql: "",
      pendingQuery: "",
      mode: "nl",
    });
    setLoading(false);
  };

  const resetConversation = (): void => {
    const newSessionId = resetSessionId(dbConfig.dbname);
    setSessionId(newSessionId);
    setResults([]);
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
          onSubmit={(mode) => sendQuery(mode)}
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
