import React, { useEffect, useRef, useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";
import SchemaSidebar from "@/components/SchemaSidebar";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useNavigate } from "react-router-dom";

export default function Query() {
  const navigate = useNavigate();
  const [showSidebar, setShowSidebar] = useState(true);
  const [schema, setSchema] = useState({});
  const [chatHistory, setChatHistory] = useState([
    {
      role: "system",
      content: "You are a helpful assistant. Only output SQL.",
    },
  ]);
  const [query, setQuery] = useState("");
  const [result, setResult] = useState(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [sqlPreview, setSqlPreview] = useState("");
  const [loading, setLoading] = useState(false);
  const chatContainerRef = useRef(null);
  const dbConfig = JSON.parse(localStorage.getItem("dbConfig") || "null");
  if (!dbConfig || !dbConfig.dbname) {
    navigate("/");
    return null;
  }

  useEffect(() => {
    fetch("/query")
      .then((res) => res.json())
      .then((data) => {
        setSchema(data.Schema || {});
      });
  }, []);

  const sendQuery = async (confirmed = false) => {
    if (!query) return;
    setLoading(true);
    const body = {
      nl_query: query,
      confirmed,
      history: chatHistory,
    };
    const res = await fetch("/query", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const data = await res.json();
    if (data.needs_confirmation) {
      setSqlPreview(data.sql_preview);
      setModalOpen(true);
    } else {
      setResult(data);
      setChatHistory(data.history || chatHistory);
    }
    setLoading(false);
    setQuery("");
    setTimeout(() => {
      chatContainerRef.current?.scrollTo({
        top: chatContainerRef.current.scrollHeight,
        behavior: "smooth",
      });
    }, 100);
  };

  const handleConfirm = () => {
    setModalOpen(false);
    sendQuery(true);
  };

  return (
    <div className="flex h-screen">
      <div
        className={`transition-all duration-300 ${
          showSidebar ? "w-[25rem]" : "w-0"
        } overflow-hidden`}
      >
        {showSidebar && <SchemaSidebar />}
      </div>
      <div className="flex-1 flex flex-col">
        <header className="border-b p-4 bg-white flex justify-between items-center">
          <h1 className="text-2xl font-bold text-gray-800">NL → SQL Chat</h1>
          <div className="flex gap-2 items-center">
            <Button
              className={`${
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
              onClick={() => (location.href = "/select")}
            >
              Change DB
            </Button>
          </div>
        </header>

        <main
          ref={chatContainerRef}
          className="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-100/50"
        >
          {chatHistory.filter((m) => m.role !== "system").length === 0 ? (
            <div className="flex items-center justify-center h-full text-center text-gray-600 text-lg">
              👋 Welcome! Ask something about your database to get started.
            </div>
          ) : (
            chatHistory
              .filter((m) => m.role !== "system")
              .map((m, i) => (
                <div
                  key={i}
                  className={`max-w-lg px-4 py-2 rounded-lg ${
                    m.role === "user"
                      ? "bg-blue-600 text-white self-end ml-auto"
                      : "bg-white text-gray-800"
                  }`}
                >
                  {m.content}
                </div>
              ))
          )}

          {result?.sql_preview && (
            <pre className="bg-gray-100 border p-3 rounded text-sm whitespace-pre-wrap font-mono">
              {result.sql_preview}
            </pre>
          )}

          {result?.message && (
            <div className="text-green-700 font-semibold">{result.message}</div>
          )}
          {result?.error && (
            <div className="text-red-600 font-semibold">{result.error}</div>
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

        <Dialog open={modalOpen} onOpenChange={setModalOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirm SQL Execution</DialogTitle>
            </DialogHeader>
            <pre className="bg-gray-100 p-3 rounded font-mono text-sm overflow-x-auto">
              {sqlPreview}
            </pre>
            <DialogFooter className="flex gap-2 justify-end">
              <Button variant="outline" onClick={() => setModalOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleConfirm}>Execute</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
}
