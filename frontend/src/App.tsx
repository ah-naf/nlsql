import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import Query from "./pages/Query";
import Connect from "./pages/Connect";
import Select from "./pages/Select";

export default function App() {
  const isConnected = localStorage.getItem("dbConfig") !== null;

  return (
    <Router>
      <Routes>
        <Route path="/" element={isConnected ? <Select /> : <Connect />} />
        <Route
          path="/select"
          element={isConnected ? <Select /> : <Navigate to="/" />}
        />
        <Route
          path="/query"
          element={isConnected ? <Query /> : <Navigate to="/" />}
        />
      </Routes>
    </Router>
  );
}
