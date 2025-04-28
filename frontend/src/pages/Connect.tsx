import React, { useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent } from "@/components/ui/card";
import axios from "axios";
import { useNavigate } from "react-router-dom";
import {
  Database,
  Server,
  KeyRound,
  User,
  FileText,
  Lock,
  Loader,
} from "lucide-react";

interface DBConfig {
  host: string;
  port: string;
  user: string;
  pass: string;
  dbname: string;
  provider: string;
  sslmode: string;
}

export default function Connect() {
  const [form, setForm] = useState<DBConfig>({
    host: "",
    port: "",
    user: "",
    pass: "",
    dbname: "",
    provider: "postgresql",
    sslmode: "disable",
  });

  const [connectionString, setConnectionString] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [activeTab, setActiveTab] = useState<string>("form");
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);

  const providers = [
    { value: "postgresql", label: "PostgreSQL" },
    { value: "mysql", label: "MySQL" },
    // { value: "mongodb", label: "MongoDB" },
    // { value: "mssql", label: "SQL Server" },
    // { value: "oracle", label: "Oracle" },
  ];

  const sslModes = [
    { value: "disable", label: "Disable" },
    { value: "require", label: "Require" },
    // { value: "verify-ca", label: "Verify CA" },
    // { value: "verify-full", label: "Verify Full" },
  ];

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setForm((prev) => ({ ...prev, [name]: value }));
  };

  const handleSelectChange = (name: string, value: string) => {
    setForm((prev) => ({ ...prev, [name]: value }));
  };

  const handleConnStringChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setConnectionString(e.target.value);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      setLoading(true);
      const payload = { ...form, connectionString };
      const res = await axios.post("http://localhost:8080/connect", payload);
      localStorage.setItem(
        "dbConfig",
        JSON.stringify({ ...form, connectionString })
      );
      localStorage.setItem("databases", JSON.stringify(res.data.databases));
      window.location.href = "/select";

      // eslint-disable-next-line
    } catch (err: any) {
      setError(err.response?.data?.error || "Connection failed");
    } finally {
      setLoading(false);
    }
  };

  // Icons for form fields
  const fieldIcons = {
    provider: <Database className="h-4 w-4 text-gray-500" />,
    host: <Server className="h-4 w-4 text-gray-500" />,
    port: <FileText className="h-4 w-4 text-gray-500" />,
    user: <User className="h-4 w-4 text-gray-500" />,
    pass: <Lock className="h-4 w-4 text-gray-500" />,
    dbname: <Database className="h-4 w-4 text-gray-500" />,
    sslmode: <KeyRound className="h-4 w-4 text-gray-500" />,
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-50">
      <Card className="w-full max-w-md shadow-lg border-0">
        <h1 className="text-center text-2xl font-bold">Connect To Database</h1>

        <CardContent className="p-6 pt-3">
          {error && (
            <div className="mb-6 p-3 rounded-md bg-red-50 border border-red-200 text-red-600 text-sm">
              {error}
            </div>
          )}

          <Tabs
            defaultValue="form"
            value={activeTab}
            onValueChange={setActiveTab}
            className="w-full"
          >
            <TabsList className="grid grid-cols-2 mb-6 rounded-lg bg-gray-100">
              <TabsTrigger
                value="form"
                className="rounded-md data-[state=active]:bg-white data-[state=active]:shadow-sm"
              >
                Connection Form
              </TabsTrigger>
              <TabsTrigger
                value="string"
                className="rounded-md data-[state=active]:bg-white data-[state=active]:shadow-sm"
              >
                Connection String
              </TabsTrigger>
            </TabsList>

            <TabsContent value="form">
              <form className="space-y-4" onSubmit={handleSubmit}>
                <div className="space-y-1.5">
                  <label className="text-sm font-medium text-gray-700 flex items-center gap-2">
                    {fieldIcons.provider} Database Provider
                  </label>
                  <Select
                    value={form.provider}
                    onValueChange={(value) =>
                      handleSelectChange("provider", value)
                    }
                  >
                    <SelectTrigger className="w-full bg-white border border-gray-200 rounded-md h-10">
                      <SelectValue placeholder="Select database provider" />
                    </SelectTrigger>
                    <SelectContent>
                      {providers.map((provider) => (
                        <SelectItem key={provider.value} value={provider.value}>
                          {provider.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {["host", "port", "user", "pass", "dbname"].map((key) => (
                  <div key={key} className="space-y-1.5">
                    <label
                      htmlFor={key}
                      className="text-sm font-medium text-gray-700 flex items-center gap-2"
                    >
                      {fieldIcons[key as keyof typeof fieldIcons]}{" "}
                      {key.charAt(0).toUpperCase() + key.slice(1)}
                    </label>
                    <div className="relative">
                      <Input
                        id={key}
                        name={key}
                        placeholder={key.charAt(0).toUpperCase() + key.slice(1)}
                        value={form[key as keyof DBConfig]}
                        onChange={handleChange}
                        type={key === "pass" ? "password" : "text"}
                        required={key !== "dbname"}
                        className="h-10 pl-3 bg-white border border-gray-200 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      />
                    </div>
                  </div>
                ))}

                <div className="space-y-1.5">
                  <label className="text-sm font-medium text-gray-700 flex items-center gap-2">
                    {fieldIcons.sslmode} SSL Mode
                  </label>
                  <Select
                    value={form.sslmode}
                    onValueChange={(value) =>
                      handleSelectChange("sslmode", value)
                    }
                  >
                    <SelectTrigger className="w-full bg-white border border-gray-200 rounded-md h-10">
                      <SelectValue placeholder="Select SSL mode" />
                    </SelectTrigger>
                    <SelectContent>
                      {sslModes.map((mode) => (
                        <SelectItem key={mode.value} value={mode.value}>
                          {mode.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className="pt-4 flex justify-between gap-3">
                  <Button
                    type="submit"
                    disabled={loading}
                    className="px-5 w-1/2 bg-blue-600 hover:bg-blue-700 text-white"
                  >
                    Connect
                    {loading && <Loader className="animate-spin" />}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    className="px-5 w-1/2 border-gray-300 text-gray-700 hover:bg-gray-50"
                    onClick={() => {
                      localStorage.clear();
                      navigate("/");
                    }}
                  >
                    Reset
                  </Button>
                </div>
              </form>
            </TabsContent>

            <TabsContent value="string">
              <form className="space-y-5" onSubmit={handleSubmit}>
                <div className="space-y-1.5">
                  <label className="text-sm font-medium text-gray-700 flex items-center gap-2">
                    {fieldIcons.provider} Database Provider
                  </label>
                  <Select
                    value={form.provider}
                    onValueChange={(value) =>
                      handleSelectChange("provider", value)
                    }
                  >
                    <SelectTrigger className="w-full bg-white border border-gray-200 rounded-md h-10">
                      <SelectValue placeholder="Select database provider" />
                    </SelectTrigger>
                    <SelectContent>
                      {providers.map((provider) => (
                        <SelectItem key={provider.value} value={provider.value}>
                          {provider.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-1.5">
                  <label
                    htmlFor="connectionString"
                    className="text-sm font-medium text-gray-700 flex items-center gap-2"
                  >
                    <Database className="h-4 w-4 text-gray-500" /> Connection
                    String
                  </label>
                  <Input
                    id="connectionString"
                    name="connectionString"
                    placeholder="Enter your database connection string"
                    value={connectionString}
                    onChange={handleConnStringChange}
                    required
                    className="h-10 bg-white border border-gray-200 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>

                <div className="pt-4 flex justify-between gap-3">
                  <Button
                    type="submit"
                    disabled={loading}
                    className="px-5 w-1/2 bg-blue-600 hover:bg-blue-700 text-white"
                  >
                    Connect
                    {loading && <Loader className="animate-spin" />}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    className="px-5 w-1/2 border-gray-300 text-gray-700 hover:bg-gray-50"
                    onClick={() => {
                      localStorage.clear();
                      navigate("/");
                    }}
                  >
                    Reset
                  </Button>
                </div>
              </form>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}
