# NL→SQL Chat

A simple web application that lets you ask natural-language questions about your PostgreSQL database and have them converted into SQL queries by an LLM. It includes schema browsing, query confirmation for destructive statements, and result display with pagination and hover-tooltips.

<center>
<img src="nlsql.gif" alt="nlsql" width="800"/>
</center>


## Updates

### 25 April 2025

1. **React UI Migration**  
   The frontend has been fully migrated from vanilla JavaScript to **React** for a more modern, modular UI experience. You can now build the frontend using:

   ```bash
   npm run build
   ```

   The Go server serves the bundled frontend seamlessly.

2. **Sidebar Virtualization**  
   Previously, rendering all table names and details in the sidebar worked fine for small databases (10–15 tables), but caused performance issues with larger schemas. This has been resolved by implementing **virtualized scrolling** using the `@tanstack/virtual` package, significantly improving performance for large databases.

3. **Smarter Prompt Construction for LLM**  
   Originally, the entire schema (all tables and their details) was sent to the LLM for SQL generation. This approach consumed a large context window for big databases. In this update, a **two-step prompt strategy** has been introduced:
   - The first prompt extracts **relevant table names** based on the user's question.
   - The second prompt uses only those tables' schemas to generate the final SQL.

   This drastically reduces token usage and improves response relevance and latency for large databases.

4. **Transactional DML Execution with Context Cancellation**  
   Destructive queries such as `ALTER`, `DELETE`, `UPDATE`, etc. are now executed within a **database transaction**.  
   If the client cancels the request (e.g., via timeout or manual cancel), the system **rolls back** to the previous state automatically using Go's context management. This ensures database integrity and safe handling of partial or aborted operations.


## Features

- **Natural Language → SQL**: Describe what you want in plain English, and the app generates a SQL statement.
- **Schema Browser**: View tables, columns, data types, primary/foreign key badges, and search/filter tables.
- **Confirmation Flow**: Destructive operations (INSERT/UPDATE/DELETE/ALTER/CREATE/DROP) are flagged and require user confirmation.
- **Result Rendering**: SELECT results show in a responsive, truncated table with hover popovers for long text.
- **Database Selection & Creation**: Connect to an existing database or create a new one from the UI.

## How It Works

1. **Frontend (HTML/CSS/JS)**
   - _query.html/query.js_: Handles user input, conversation history, AJAX calls, and renders chat bubbles and schema cards.
   - _select.html_: Lists available databases and lets you create a new one.
   - _favicon.svg_, _CSS_, and _JS_ live in `/static/`.
2. **Backend (Go / Gin)**
   - **Session Middleware** stores connection credentials.
   - **Handlers**:
     - `ShowConnectForm` / `ConnectDB` → initial Postgres connection.
     - `ShowDBForm` / `SelectDB` → list or create databases and save `connection_string` in session.
     - `ShowQueryPage` → serve the chat interface with preloaded schema.
     - `HandleNLQuery` → bind JSON, rebuild prompt with schema, call Together AI LLM, return SQL preview, execute if `SELECT`, and refresh schema for DML.
   - **Schema Module** (`models/db.go`) inspects `information_schema` for columns, PK/FK metadata.
3. **LLM Integration**
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
   go mod tidy
   ```

4. **Set environment variables**
   Create a `.env` file in the project root (or export in your shell):

   ```ini
   LLM_API_KEY=your-together-ai-key
   ```

5. **Run the server**
   ```bash
   go run main.go
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
