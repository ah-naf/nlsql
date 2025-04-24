// pages/Select.tsx
import React, { useEffect, useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import axios from "axios";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Select as UISelect,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import { useNavigate } from "react-router-dom";

export default function Select() {
  const [databases, setDatabases] = useState<string[]>([]);
  const [newdb, setNewdb] = useState("");
  const [selected, setSelected] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [dbToDelete, setDbToDelete] = useState<string | null>(null);
  const navigate = useNavigate();
  const dbConfig = JSON.parse(localStorage.getItem("dbConfig") || "null");
  if (!dbConfig) {
    navigate("/");
    return null;
  }

  useEffect(() => {
    const fetchDatabases = async () => {
      try {
        const res = await axios.get("http://localhost:8080/databases", {
          params: dbConfig,
        });
        setDatabases(res.data.databases || []);
        localStorage.setItem(
          "databases",
          JSON.stringify(res.data.databases || [])
        );
      } catch (err: any) {
        setError(err.response?.data?.error || "Failed to load databases");
      }
    };

    fetchDatabases();
  }, []);

  const handleCreate = async () => {
    try {
      await axios.post("http://localhost:8080/create", {
        ...dbConfig,
        dbname: newdb,
      });
      const updated = [...databases, newdb];
      setDatabases(updated);
      localStorage.setItem("databases", JSON.stringify(updated));
      setNewdb("");
      setSuccess(`Database '${newdb}' created successfully.`);
      setError("");
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to create database");
      setSuccess("");
    }
  };

  const confirmDelete = async () => {
    if (!dbToDelete) return;
    try {
      await axios.post("http://localhost:8080/delete", {
        ...dbConfig,
        dbname: dbToDelete,
      });
      const updated = databases.filter((db) => db !== dbToDelete);
      setDatabases(updated);
      localStorage.setItem("databases", JSON.stringify(updated));
      if (selected === dbToDelete) setSelected("");
      setDbToDelete(null);
      setSuccess(`Database '${dbToDelete}' deleted successfully.`);
      setError("");
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to delete database");
      setSuccess("");
    }
  };

  const handleSelect = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!dbConfig) return navigate("/");
    dbConfig.dbname = selected;
    localStorage.setItem("dbConfig", JSON.stringify(dbConfig));
    window.location.href = "/query";
  };

  return (
    <div className="min-h-screen flex items-center justify-center px-4 py-12 bg-gradient-to-br from-sky-50 to-indigo-100">
      <div className="bg-white w-full max-w-xl rounded-2xl shadow-xl p-8 sm:p-10 border border-gray-200 space-y-6">
        <h1 className="text-3xl font-bold text-center text-gray-800">
          Select a Database
        </h1>
        {error && (
          <div className="text-red-600 text-sm bg-red-100 border border-red-300 rounded p-3">
            {error}
          </div>
        )}
        {success && (
          <div className="text-green-700 text-sm bg-green-100 border border-green-300 rounded p-3">
            {success}
          </div>
        )}

        <div className="flex items-center gap-2">
          <Input
            value={newdb}
            onChange={(e) => setNewdb(e.target.value)}
            placeholder="New database name"
          />
          <Button
            className="bg-green-600 hover:bg-green-700"
            onClick={handleCreate}
          >
            Create
          </Button>
        </div>

        <form onSubmit={handleSelect} className="space-y-4">
          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Select a Database
            </label>
            <UISelect onValueChange={setSelected} value={selected}>
              <SelectTrigger className="w-full">
                <SelectValue placeholder="-- Choose Database --" />
              </SelectTrigger>
              <SelectContent>
                {databases.map((db) => (
                  <SelectItem key={db} value={db}>
                    {db}
                  </SelectItem>
                ))}
              </SelectContent>
            </UISelect>
          </div>

          <div className="flex justify-between gap-2">
            <Button
              variant="secondary"
              onClick={(e) => {
                e.preventDefault();
                localStorage.removeItem("dbConfig");
                localStorage.removeItem("databases");
                window.location.href = "/";
              }}
            >
              Reset
            </Button>
            <div className="flex gap-2">
              <Button
                type="button"
                variant="destructive"
                onClick={() => selected && setDbToDelete(selected)}
              >
                Delete
              </Button>
              <Button
                className="bg-indigo-600 hover:bg-indigo-700"
                type="submit"
              >
                Query
              </Button>
            </div>
          </div>
        </form>

        <Dialog open={!!dbToDelete} onOpenChange={() => setDbToDelete(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirm Delete</DialogTitle>
            </DialogHeader>
            <p className="text-sm text-gray-700">
              Are you sure you want to delete <strong>{dbToDelete}</strong>?
            </p>
            <DialogFooter className="pt-4">
              <Button variant="secondary" onClick={() => setDbToDelete(null)}>
                Cancel
              </Button>
              <Button
                className="bg-red-600 hover:bg-red-700"
                onClick={confirmDelete}
              >
                Delete
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
}
