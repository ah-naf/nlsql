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
import { Loader, Database, Trash, Plus } from "lucide-react";

export default function Select() {
  const [databases, setDatabases] = useState<string[]>([]);
  const [newdb, setNewdb] = useState("");
  const [selected, setSelected] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [dbToDelete, setDbToDelete] = useState<string | null>(null);

  // Loading states
  const [fetchingDatabases, setFetchingDatabases] = useState(true);
  const [creatingDatabase, setCreatingDatabase] = useState(false);
  const [deletingDatabase, setDeletingDatabase] = useState(false);
  const [navigatingToQuery, setNavigatingToQuery] = useState(false);

  const navigate = useNavigate();
  const dbConfig = JSON.parse(localStorage.getItem("dbConfig") || "null");

  useEffect(() => {
    const fetchDatabases = async () => {
      setFetchingDatabases(true);
      try {
        const res = await axios.get("http://localhost:8080/databases", {
          params: dbConfig,
        });
        setDatabases(res.data.databases || []);
        localStorage.setItem(
          "databases",
          JSON.stringify(res.data.databases || [])
        );

        // eslint-disable-next-line
      } catch (err: any) {
        setError(err.response?.data?.error || "Failed to load databases");
      } finally {
        setFetchingDatabases(false);
      }
    };

    fetchDatabases();
  }, []);

  const handleCreate = async () => {
    if (!newdb.trim()) return;

    setCreatingDatabase(true);
    setError("");
    setSuccess("");

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

      // eslint-disable-next-line
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to create database");
    } finally {
      setCreatingDatabase(false);
    }
  };

  const confirmDelete = async () => {
    if (!dbToDelete) return;

    setDeletingDatabase(true);
    setError("");
    setSuccess("");

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

      // eslint-disable-next-line
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to delete database");
    } finally {
      setDeletingDatabase(false);
    }
  };

  const handleSelect = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selected) return;
    if (!dbConfig) return navigate("/");

    setNavigatingToQuery(true);
    dbConfig.dbname = selected;
    localStorage.setItem("dbConfig", JSON.stringify(dbConfig));
    window.location.href = "/query";
  };

  if (!dbConfig) {
    navigate("/");
    return null;
  }

  return (
    <div className="min-h-screen flex items-center justify-center px-4 py-12 bg-gradient-to-br from-blue-50 to-indigo-50">
      <div className="bg-white w-full max-w-xl rounded-2xl shadow-lg border-0 p-8 sm:p-10 space-y-6">
        <div className="flex items-center justify-center space-x-2">
          <Database className="h-6 w-6 text-blue-600" />
          <h1 className="text-2xl font-bold text-center text-gray-800">
            Select a Database
          </h1>
        </div>

        {error && (
          <div className="text-red-600 text-sm bg-red-50 border border-red-200 rounded-md p-3">
            {error}
          </div>
        )}

        {success && (
          <div className="text-green-700 text-sm bg-green-50 border border-green-200 rounded-md p-3">
            {success}
          </div>
        )}

        {fetchingDatabases ? (
          <div className="py-8 flex flex-col items-center justify-center text-gray-500">
            <Loader className="h-8 w-8 animate-spin mb-3" />
            <p>Fetching databases...</p>
          </div>
        ) : (
          <>
            <div className="flex items-center gap-2">
              <div className="relative flex-1">
                <Input
                  value={newdb}
                  onChange={(e) => setNewdb(e.target.value)}
                  placeholder="New database name"
                  disabled={creatingDatabase}
                  className="pr-3 bg-white border border-gray-200"
                />
              </div>
              <Button
                className="bg-green-600 hover:bg-green-700 min-w-24"
                onClick={handleCreate}
                disabled={creatingDatabase || !newdb.trim()}
              >
                {creatingDatabase ? (
                  <>
                    <Loader className="h-4 w-4 animate-spin mr-2" />
                    Creating
                  </>
                ) : (
                  <>
                    <Plus className="h-4 w-4 mr-2" />
                    Create
                  </>
                )}
              </Button>
            </div>

            <form onSubmit={handleSelect} className="space-y-5">
              <div className="space-y-1.5">
                <label className="block text-sm font-medium text-gray-700">
                  Select a Database
                </label>
                <UISelect onValueChange={setSelected} value={selected}>
                  <SelectTrigger className="w-full bg-white border border-gray-200">
                    <SelectValue placeholder="-- Choose Database --" />
                  </SelectTrigger>
                  <SelectContent>
                    {databases.length === 0 ? (
                      <SelectItem value="no-db" disabled>
                        No databases available
                      </SelectItem>
                    ) : (
                      databases.map((db) => (
                        <SelectItem key={db} value={db}>
                          {db}
                        </SelectItem>
                      ))
                    )}
                  </SelectContent>
                </UISelect>
              </div>

              <div className="flex justify-between gap-2 pt-2">
                <Button
                  variant="outline"
                  type="button"
                  className="border-gray-300 text-gray-700"
                  onClick={() => {
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
                    variant="outline"
                    className="border-red-200 text-red-600 hover:bg-red-50 hover:text-red-700"
                    onClick={() => selected && setDbToDelete(selected)}
                    disabled={!selected || deletingDatabase}
                  >
                    <Trash className="h-4 w-4 mr-2" />
                    Delete
                  </Button>
                  <Button
                    className="bg-blue-600 hover:bg-blue-700 min-w-24"
                    type="submit"
                    disabled={!selected || navigatingToQuery}
                  >
                    {navigatingToQuery ? (
                      <>
                        <Loader className="h-4 w-4 animate-spin mr-2" />
                        Loading
                      </>
                    ) : (
                      "Query"
                    )}
                  </Button>
                </div>
              </div>
            </form>
          </>
        )}

        <Dialog
          open={!!dbToDelete}
          onOpenChange={() => !deletingDatabase && setDbToDelete(null)}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirm Delete</DialogTitle>
            </DialogHeader>
            <p className="text-sm text-gray-700">
              Are you sure you want to delete <strong>{dbToDelete}</strong>?
              This action cannot be undone.
            </p>
            <DialogFooter className="pt-4">
              <Button
                variant="outline"
                onClick={() => setDbToDelete(null)}
                disabled={deletingDatabase}
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={confirmDelete}
                disabled={deletingDatabase}
              >
                {deletingDatabase ? (
                  <>
                    <Loader className="h-4 w-4 animate-spin mr-2" />
                    Deleting
                  </>
                ) : (
                  "Delete"
                )}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
}
