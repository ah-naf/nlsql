// src/utils/dbUtils.ts
import axios from "axios";
import { DBConfig } from "../types/query";

export const extractSqlFromQaOutput = (output: string): string | null => {
  if (!output) return null;

  // Check for pattern: SELECT '...' AS output
  const selectMatch = output.match(/SELECT\s+'(.+?)'\s+AS\s+output/i);
  if (selectMatch && selectMatch[1]) {
    // Unescape single quotes in the SQL
    return selectMatch[1].replace(/''/g, "'");
  }
  return null;
};

export const sendQueryToBackend = async (
  dbConfig: DBConfig,
  query: string,
  confirmed: boolean,
  sqlToConfirm: string,
  sessionId: string
) => {
  return await axios.post(
    "https://nl-sql-gme3hme5fpfdg2ez.canadacentral-01.azurewebsites.net/query",
    {
      config: dbConfig,
      prompt: query,
      confirmed,
      sqlToConfirm,
      sessionId,
    }
  );
};

export const getSessionId = (dbName: string): string => {
  // Check if there's a stored session ID for this database
  const storedSessionId = localStorage.getItem(`sessionId-${dbName}`);

  // Generate a new session ID if none exists
  if (!storedSessionId) {
    const newSessionId = `${dbName}-${Date.now()}`;
    localStorage.setItem(`sessionId-${dbName}`, newSessionId);
    return newSessionId;
  }

  return storedSessionId;
};

export const resetSessionId = (dbName: string): string => {
  // Clear the current session ID in localStorage
  localStorage.removeItem(`sessionId-${dbName}`);

  // Generate a new session ID
  const newSessionId = `${dbName}-${Date.now()}`;
  localStorage.setItem(`sessionId-${dbName}`, newSessionId);

  return newSessionId;
};
