import { Button } from "@/components/ui/button";
import { useNavigate } from "react-router-dom";

interface QueryHeaderProps {
  showSidebar: boolean;
  setShowSidebar: (show: boolean) => void;
  resetConversation: () => void;
}

export default function QueryHeader({
  showSidebar,
  setShowSidebar,
  resetConversation,
}: QueryHeaderProps) {
  const navigate = useNavigate();
  const dbConfig = JSON.parse(localStorage.getItem("dbConfig") || "{}");

  return (
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
          onClick={() => {
            if (dbConfig && dbConfig.provider == "demo") {
              localStorage.removeItem("dbConfig");
              localStorage.removeItem("databases");
              window.location.href = "/";
              return;
            }
            navigate("/select");
          }}
        >
          Change DB
        </Button>
      </div>
    </header>
  );
}
