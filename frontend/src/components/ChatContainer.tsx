// src/components/ChatContainer.tsx
import React from "react";
import ChatMessage from "./ChatMessage";
import { ResultItem } from "../types/query";

interface ChatContainerProps {
  chatContainerRef: React.RefObject<HTMLDivElement | null>;
  results: ResultItem[];
  activeCodeIndex: number | null;
  toggleCodeView: (index: number) => void;
  onExecuteSql: (sql: string) => void;
  loading: boolean;
}

export default function ChatContainer({
  chatContainerRef,
  results,
  activeCodeIndex,
  toggleCodeView,
  onExecuteSql,
  loading,
}: ChatContainerProps) {
  return (
    <main
      ref={chatContainerRef}
      className="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-100/50"
    >
      {results.length === 0 ? (
        <div className="flex items-center justify-center h-full text-center text-gray-600 text-lg">
          👋 Welcome! Ask something about your database to get started.
        </div>
      ) : (
        <>
          {results.map((message, i) => (
            <ChatMessage
              key={i}
              message={message}
              index={i}
              activeCodeIndex={activeCodeIndex}
              toggleCodeView={toggleCodeView}
              onExecuteSql={onExecuteSql}
            />
          ))}

          {loading && (
            <ChatMessage
              key="loading"
              loading
              activeCodeIndex={activeCodeIndex}
              toggleCodeView={toggleCodeView}
              onExecuteSql={onExecuteSql}
            />
          )}
        </>
      )}
    </main>
  );
}
