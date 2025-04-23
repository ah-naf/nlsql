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
  const dbConfig = JSON.parse(localStorage.getItem("dbConfig") || "null");

  return (
    <Router>
      <Routes>
        <Route
          path="/"
          element={
            dbConfig ? dbConfig.dbname ? <Query /> : <Select /> : <Connect />
          }
        />
        <Route
          path="/select"
          element={dbConfig ? <Select /> : <Navigate to="/" />}
        />
        <Route
          path="/query"
          element={
            dbConfig && dbConfig.dbname ? <Query /> : <Navigate to="/" />
          }
        />
      </Routes>
    </Router>
  );
}
