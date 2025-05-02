# NL→SQL Chat

A simple web application that lets you ask natural-language questions about your PostgreSQL database and have them converted into SQL queries by an LLM. It includes schema browsing, query confirmation for destructive statements, and result display with pagination and hover-tooltips.

## Features

- **Natural Language → SQL**: Describe what you want in plain English, and the app generates a SQL statement.
- **Schema Browser**: View tables, columns, data types, primary/foreign key badges, and search/filter tables.
- **Confirmation Flow**: Destructive operations (INSERT/UPDATE/DELETE/ALTER/CREATE/DROP) are flagged and require user confirmation.
- **Result Rendering**: SELECT results show in a responsive, truncated table with hover popovers for long text.
- **Database Selection & Creation**: Connect to an existing database or create a new one from the UI.

## How It Works

1. **Backend (Go / Gin)**
   - **Session Middleware** stores connection credentials.
   - **Handlers**:
     - `ShowConnectForm` / `ConnectDB` → initial Postgres connection.
     - `ShowDBForm` / `SelectDB` → list or create databases and save `connection_string` in session.
     - `ShowQueryPage` → serve the chat interface with preloaded schema.
     - `HandleNLQuery` → bind JSON, rebuild prompt with schema, call Together AI LLM, return SQL preview, execute if `SELECT`, and refresh schema for DML.
   - **Schema Module** (`models/db.go`) inspects `information_schema` for columns, PK/FK metadata.
2. **LLM Integration**
   - Uses Together AI’s API (`meta-llama/Llama-3.3-70B-Instruct-Turbo-Free` model) to translate English into SQL.
   - Requires an environment variable `LLM_API_KEY` with your Together AI API key.

## Installation & Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/ah-naf/nlsql.git
   cd nlsql
   ```

2. **Install Go (1.20+)**
   Make sure `go` is in your PATH and version is at least 1.20.

3. **Fetch dependencies**

   ```bash
   cd backend
   go mod tidy
   ```

4. **Set environment variables**
   Create a `.env` file in the project root (or export in your shell):

   ```ini
   LLM_API_KEY=your-together-ai-key
   ```

5. **Setup Frontend**
   ```bash
   cd frontend
   npm install
   npm run build
   ```
6. **Run the server**
   ```bash
   cd backend
   go run cmd/main.go
   ```
   By default, the app listens on `http://localhost:8080`.

## Usage

1. **Browse to** `http://localhost:8080`.
2. **Connect** to your Postgres server using the form.
3. **Select** an existing database or **create** a new one.
4. **Ask** questions in natural language, e.g.:
   - "Show me all users who signed up in the last 7 days"
   - "Add a new product named 'Gadget' with price 19.99"
5. **Review** the generated SQL in the chat bubble.
6. **Confirm** if it’s a destructive query.
7. **View** results directly in the chat or see a success message for updates.

## License

This project is open-source and available under the MIT License.
