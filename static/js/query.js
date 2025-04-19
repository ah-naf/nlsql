// Escape HTML to prevent XSS
function escapeHtml(str) {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

// Truncate long text for table display
function truncateText(text, maxLength = 30) {
  if (!text || typeof text !== "string") return text;
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + "...";
}

$(function () {
  // Make chat container responsive initially
  function adjustChatContainerHeight() {
    const windowHeight = $(window).height();
    const headerHeight = $("header").outerHeight();
    const footerHeight = $("footer").outerHeight();
    const chatHeight = windowHeight - headerHeight - footerHeight;
    $("#chat-container").css("height", chatHeight + "px");
  }

  // Run on load and resize
  $(window).on("resize", adjustChatContainerHeight);
  adjustChatContainerHeight();

  $("#query-form").submit(function (e) {
    e.preventDefault();
    const nl = $("#nl_query").val().trim();
    if (!nl) return;

    // 1) hide welcome message
    $("#welcome-message").hide();

    // 2) append user bubble
    $("#chat-container").append(`
          <div class="flex justify-end mb-4">
            <div class="bg-blue-600 text-white px-4 py-2 rounded-lg shadow max-w-lg break-words">
              ${escapeHtml(nl)}
            </div>
          </div>
        `);
    // scroll to bottom
    $("#chat-container").scrollTop($("#chat-container")[0].scrollHeight);

    let confirmed = false;

    function submitQuery() {
      $.ajax({
        url: "/query",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify({ nl_query: nl, confirmed }),
        success: function (resp) {
          // 3) confirmation flow
          if (resp.needs_confirmation) {
            $("#modal-sql").text(resp.sql_preview);
            $("#confirm-modal").removeClass("hidden");
            $("#modal-confirm").one("click", () => {
              $("#confirm-modal").addClass("hidden");
              confirmed = true;
              submitQuery();
            });
            $("#modal-cancel").one("click", () => {
              $("#confirm-modal").addClass("hidden");
            });
            return;
          }

          // 4) build server bubble - constrained width
          let bubble = `<div class="flex justify-start mb-4">
                          <div class="bg-white p-3 rounded-lg shadow max-w-full space-y-2">`;

          // 4a) SQL preview
          bubble += `<div class="font-mono text-xs text-gray-500 break-words">SQL: ${escapeHtml(
            resp.sql_preview
          )}</div>`;

          // 4b) error?
          if (resp.error) {
            bubble += `<div class="text-red-600 font-semibold">${escapeHtml(
              resp.error
            )}</div>`;
          }
          // 4c) simple message?
          else if (resp.message) {
            bubble += `<div class="text-gray-800">${escapeHtml(
              resp.message
            )}</div>`;
          }
          // 4d) table result?
          else if (resp.table && resp.table.length) {
            const rows = resp.table;
            const cols = Object.keys(rows[0]);

            // Add row count info at top
            bubble += `<div class="text-xs text-gray-500 mb-1">${
              rows.length
            } row${rows.length !== 1 ? "s" : ""} returned</div>`;

            // Compact table with fixed layout
            bubble += `<div class="overflow-x-auto border border-gray-200 rounded-md shadow-sm">
                         <table class="w-full text-sm table-fixed">`;

            // header - make it sticky
            bubble += `<thead class="bg-gray-100 sticky top-0">
                          <tr>`;
            cols.forEach((col) => {
              // Calculate appropriate column width based on content type
              let colWidth = "150px"; // default width
              if (col.toLowerCase().includes("email")) colWidth = "180px";
              else if (
                col.toLowerCase().includes("date") ||
                col.toLowerCase().includes("_at")
              )
                colWidth = "160px";
              else if (
                col.toLowerCase().includes("name") ||
                col.toLowerCase() === "first" ||
                col.toLowerCase() === "last"
              )
                colWidth = "100px";
              else if (col.toLowerCase().includes("id")) colWidth = "80px";

              bubble += `<th class="px-2 py-1 text-left font-medium text-gray-600 uppercase tracking-wider" 
                              style="width: ${colWidth}; max-width: ${colWidth};">
                              ${escapeHtml(col)}
                           </th>`;
            });
            bubble += `</tr></thead>`;

            // body - with truncated text
            bubble += `<tbody class="bg-white divide-y divide-gray-200">`;
            rows.forEach((row, rowIndex) => {
              bubble += `<tr class="${
                rowIndex % 2 === 0 ? "bg-white" : "bg-gray-50"
              }">`;
              cols.forEach((col) => {
                let cellContent = row[col] === null ? "NULL" : String(row[col]);
                let truncated = cellContent;
                let title =
                  cellContent !== truncated
                    ? `title="${escapeHtml(cellContent)}"`
                    : "";

                bubble += `<td class="px-2 py-1 text-gray-700 whitespace-normal break-words" ${title}>
                              ${escapeHtml(truncated)}
                            </td>`;
              });
              bubble += `</tr>`;
            });
            bubble += `</tbody></table></div>`;
          }

          bubble += `</div></div>`;

          // 5) append server bubble
          $("#chat-container").append(bubble);
          $("#chat-container").scrollTop($("#chat-container")[0].scrollHeight);
        },
        error: function (xhr) {
          const msg = xhr.responseJSON?.error || xhr.statusText;
          $("#chat-container").append(`
                <div class="flex justify-start mb-4">
                  <div class="bg-red-100 text-red-800 px-4 py-2 rounded-lg shadow max-w-lg">
                    ${escapeHtml(msg)}
                  </div>
                </div>
              `);
          $("#chat-container").scrollTop($("#chat-container")[0].scrollHeight);
        },
      });
    }

    submitQuery();
    // clear input
    $("#nl_query").val("");
  });
});
