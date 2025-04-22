import React, { useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import axios from "axios";
import { useNavigate } from "react-router-dom";

interface DBForm {
  host: string;
  port: string;
  user: string;
  pass: string;
  dbname: string;
}

export default function Connect() {
  const [form, setForm] = useState<DBForm>({
    host: "",
    port: "",
    user: "",
    pass: "",
    dbname: "",
  });
  const [error, setError] = useState<string>("");
  const navigate = useNavigate();

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setForm((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      const res = await axios.post("http://localhost:8080/connect", form);
      console.log(res);
      localStorage.setItem("dbConfig", JSON.stringify(form));
      localStorage.setItem("databases", JSON.stringify(res.data.databases));
      navigate("/select");
    } catch (err: any) {
      setError(err.response?.data?.error || "Connection failed");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center px-4 py-12 bg-gradient-to-br from-sky-50 to-indigo-100">
      <div className="bg-white w-full max-w-xl rounded-2xl shadow-xl p-8 sm:p-10 border border-gray-200">
        <h1 className="text-3xl font-bold text-center text-gray-800 mb-6">
          Connect to Database
        </h1>

        {error && (
          <div className="mb-6 p-3 rounded-md bg-red-100 border border-red-300 text-red-700 text-sm">
            {error}
          </div>
        )}

        <form className="space-y-5" onSubmit={handleSubmit}>
          {(Object.keys(form) as (keyof DBForm)[]).map((key) => (
            <div key={key} className="space-y-1">
              <label
                htmlFor={key}
                className="block text-sm font-medium text-gray-700"
              >
                {key.charAt(0).toUpperCase() + key.slice(1)}
              </label>
              <Input
                id={key}
                name={key}
                placeholder={key.charAt(0).toUpperCase() + key.slice(1)}
                value={form[key]}
                onChange={handleChange}
                type={key === "pass" ? "password" : "text"}
                required={key !== "dbname"}
                className="text-sm"
              />
            </div>
          ))}

          <div className="flex justify-between items-center pt-4">
            <Button type="submit" className="px-6 py-2">
              Connect
            </Button>
            <Button
              type="button"
              variant="outline"
              className="px-6 py-2"
              onClick={() => {
                localStorage.clear();
                navigate("/");
              }}
            >
              Reset
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
