// Truncate long text for table display
function truncateText(text, maxLength = 30) {
  if (!text || typeof text !== "string") return text;
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + "...";
}

function renderSchema(schema) {
  const $out = $("#schema-panel .schema-scroll").empty();
  Object.entries(schema).forEach(([table, cols]) => {
    const $tableCard = $(`
        <div class="mb-6 bg-gray-50 rounded-lg p-3 shadow schema-table">
          <h3 class="font-semibold text-gray-800 mb-2 flex items-center justify-between table-header">
            <div class="flex items-center">
              <i class="fas fa-table mr-2 text-blue-500"></i>${table}
            </div>
            <i class="fas fa-chevron-right text-gray-400 rotate-icon"></i>
          </h3>
          <ul class="space-y-1 text-gray-600 hidden table-content">
            ${cols
              .map(
                (c) => `
              <li class="flex items-center py-1 border-b border-gray-100">
                <span class="w-2 h-2 bg-blue-400 rounded-full mr-2"></span>
                ${c}
              </li>
            `
              )
              .join("")}
          </ul>
        </div>
      `);
    $out.append($tableCard);
  });

  // re‑bind the expand/collapse click handler
  $out
    .find(".table-header")
    .off("click")
    .on("click", function () {
      const $content = $(this).next(".table-content");
      const $icon = $(this).find(".rotate-icon");
      $content.slideToggle(150);
      $icon.toggleClass("down");
    });
}

$(function () {
  // Initialize conversation history
  let conversationHistory = [
    {
      role: "system",
      content: "You are a helpful assistant. Only output SQL.",
    },
  ];

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
                ${nl}
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
        data: JSON.stringify({
          nl_query: nl,
          confirmed: confirmed,
          history: conversationHistory, // Send current history with each request
        }),
        success: function (resp) {
          console.log(resp);

          if (resp.schema) {
            renderSchema(resp.schema);
          }

          // Update the conversation history from the response
          if (resp.history && Array.isArray(resp.history)) {
            conversationHistory = resp.history;
          }

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
          bubble += `<div class="font-mono text-xs text-gray-500 break-words">SQL: ${resp.sql_preview}</div>`;

          // 4b) error?
          if (resp.error) {
            bubble += `<div class="text-red-600 font-semibold">${resp.error}</div>`;
          }
          // 4c) simple message?
          else if (resp.message) {
            bubble += `<div class="text-gray-800">${resp.message}</div>`;
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
                                ${col}
                             </th>`;
            });
            bubble += `</tr></thead>`;

            // body - with truncated text
            bubble += `<tbody class="bg-white divide-y divide-gray-200">`;
            const maxLength = 30;

            rows.forEach((row, rowIndex) => {
              bubble += `<tr class="${
                rowIndex % 2 === 0 ? "bg-white" : "bg-gray-50"
              }">`;

              cols.forEach((col) => {
                const raw = row[col] == null ? "NULL" : String(row[col]);
                const truncated = truncateText(raw, maxLength);
                const isTruncated = raw.length > maxLength;

                bubble += `
        <td class="px-2 py-1 text-gray-700 relative group">
          <!-- single-line, truncated preview -->
          <div class="truncate whitespace-nowrap">
            ${truncated}
          </div>
          ${
            isTruncated
              ? `
            <!-- full-text popover on hover -->
            <div
              class="absolute left-0 top-full mt-1 hidden group-hover:block
                     bg-gray-800 text-white text-xs rounded p-2 z-20
                     whitespace-normal break-words max-w-xs"
            >
              ${raw}
            </div>
          `
              : ""
          }
        </td>
      `;
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
                      ${msg}
                    </div>
                  </div>
                `);
          $("#chat-container").scrollTop($("#chat-container")[0].scrollHeight);

          // If there's history in the error response, update our history
          if (
            xhr.responseJSON?.history &&
            Array.isArray(xhr.responseJSON.history)
          ) {
            conversationHistory = xhr.responseJSON.history;
          }
        },
      });
    }

    submitQuery();
    // clear input
    $("#nl_query").val("");
  });

  // Optional: Add a button to clear history
  if ($("#clear-history").length === 0) {
    $("#query-form").after(`
        <button id="clear-history" class="mt-2 text-sm text-gray-600 hover:text-gray-800">
          Clear conversation history
        </button>
      `);

    $("#clear-history").click(function () {
      // Reset history to initial state
      conversationHistory = [
        {
          role: "system",
          content: "You are a helpful assistant. Only output SQL.",
        },
      ];

      // Clear chat container except for the welcome message
      $("#chat-container").empty();
      $("#welcome-message").show();

      alert("Conversation history cleared");
    });
  }
});
