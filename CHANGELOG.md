## Changelog

### 26 April 2025

1. **Support Connection via Connection String**  
   You can now connect to a PostgreSQL database using a **full connection string** instead of providing individual host, port, user, and password fields. This simplifies integration for hosted databases and advanced connection setups.

2. **Fix Schema Sidebar Re-Render on Toggle**  
   Fixed an issue where hiding and showing the schema sidebar caused unnecessary re-fetching from the backend.

3. **Loading UI for Schema Fetching and Details**  
   A **loading skeleton UI** has been added for the schema sidebar while fetching table lists and toggling table details.

4. **Loading UI for Chat Bubbles**  
   When a query is sent and the system is awaiting an LLM or database response, a **loading chat bubble** is now displayed.

5. **Schema Refresh After Table Creation**  
   After creating a new table, the frontend now **automatically fetches the updated schema** and refreshes the sidebar.

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
