import React, { useRef, useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Loader2, Copy, Check, Code } from "lucide-react";
import SchemaSidebar from "@/components/SchemaSidebar";
import axios from "axios";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
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
  const [error, setError] = useState("");
  const [copied, setCopied] = useState(false);
  const [showSql, setShowSql] = useState(false);
  const chatContainerRef = useRef(null);

  const dbConfig = JSON.parse(localStorage.getItem("dbConfig") || "null");
  if (!dbConfig || !dbConfig.dbname) {
    navigate("/");
    return null;
  }

  const sendQuery = async (confirmed = false) => {
    if (!query.trim()) return;

    setLoading(true);
    setError("");

    try {
      const response = await axios.post("http://localhost:8080/query", {
        config: dbConfig,
        prompt: query,
      });

      const data = response.data;

      if (data.sql) {
        setSqlCode(data.sql);
        setResults([
          ...results,
          {
            type: "user",
            content: query,
          },
          {
            type: "assistant",
            content: data.result_table,
            sql: data.sql,
          },
        ]);
        setQuery("");
      } else if (data.error) {
        setError(data.error);
      }
    } catch (err) {
      setError(
        err.response?.data?.error ||
          "An error occurred while processing your query"
      );
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

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  // Truncate long text and show in tooltip
  const TruncatedCell = ({ content }) => {
    const contentStr = String(content);
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
  const ResultTable = ({ data }) => {
    if (!data || data.length === 0)
      return <div className="text-gray-500">No results found</div>;

    const columns = Object.keys(data[0]);

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
                    <TruncatedCell content={row[column]} />
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    );
  };

  return (
    <div className="flex h-screen">
      {showSidebar && (
        <div className="w-[25rem]">
          <SchemaSidebar />
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
                ) : (
                  <div>
                    <div className="flex justify-between items-center mb-2">
                      <h4 className="font-medium text-gray-700">Results:</h4>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-8 px-2 text-gray-600"
                        onClick={() =>
                          setShowSql((prev) =>
                            i === results.length - 1 ? !prev : prev
                          )
                        }
                      >
                        <Code size={16} className="mr-1" />
                        {showSql && i === results.length - 1
                          ? "Hide SQL"
                          : "Show SQL"}
                      </Button>
                    </div>

                    {showSql && i === results.length - 1 && (
                      <div className="mb-3 relative">
                        <pre className="bg-gray-100 p-3 rounded font-mono text-sm overflow-x-auto">
                          {message.sql}
                        </pre>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="absolute top-2 right-2 h-8 w-8 p-0"
                          onClick={() => copyToClipboard(message.sql)}
                        >
                          {copied ? <Check size={16} /> : <Copy size={16} />}
                        </Button>
                      </div>
                    )}

                    <ResultTable data={message.content} />
                  </div>
                )}
              </div>
            ))
          )}

          {error && (
            <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
              {error}
            </div>
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
    </div>
  );
}
