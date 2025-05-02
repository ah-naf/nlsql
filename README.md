# NL→SQL

A simple web application that lets you ask natural-language questions about your PostgreSQL database and have them converted into SQL queries by an LLM. It includes schema browsing, query confirmation for destructive statements, and result display with pagination and hover-tooltips.

---

## ✨ Features

- **Natural Language → SQL**: Describe what you want in plain English, and the app generates a SQL statement.
- **Schema Browser**: View tables, columns, data types, primary/foreign key badges, and search/filter tables.
- **Confirmation Flow**: Destructive operations (INSERT, UPDATE, DELETE, etc.) require confirmation.
- **Result Rendering**: Paginated, responsive table with hover popovers for long content.
- **Database Connection**: Connect to or create databases directly from the UI.

---

## 🔧 Environment Variables

Place a `.env` file at the project root with the following:

```ini
# For Together AI (default)
LLM_API_KEY=your-together-api-key
LLM_API_URL=https://api.together.xyz/v1/chat/completions
LLM_API_MODEL_NAME=meta-llama/Llama-3-70B-Instruct

# OR for Gemini (optional alternative)
GEMINI_API_URL=https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=your-gemini-api-key
````

## 🚀 Local Development

### 1. Clone the repository

```bash
git clone https://github.com/ah-naf/nlsql.git
cd nlsql
```

### 2. Backend setup

```bash
cd backend
go mod tidy
```

### 3. Frontend setup

```bash
cd ../frontend
npm install
npm run build
```

### 4. Run the application

```bash
cd ../backend
go run cmd/main.go
```

Open `http://localhost:8080` in your browser.

---

## 🐳 Run with Docker

This project supports multi-stage Docker builds (frontend + backend). Here's how to build and run:

### 1. Build the image

```bash
docker build -t nlsql-app .
```

### 2. Run the container

```bash
docker run -p 8080:8080 nlsql-app
```

> Make sure your PostgreSQL instance is accessible from the container.

### 3. Open the app

Visit: [http://localhost:8080](http://localhost:8080)

---

## 💡 Example Prompts

* `Show all orders placed in the last 24 hours`
* `Add a new user named Alice with email alice@example.com`
* `Delete all rows from temp_sessions table`

---

## 📄 License

This project is open-source and available under the [MIT License](LICENSE).
