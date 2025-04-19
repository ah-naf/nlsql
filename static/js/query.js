// Truncate long text for table display
function truncateText(text, maxLength = 30) {
  if (!text || typeof text !== "string") return text;
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + "...";
}

function renderSchema(schema) {
  console.log(schema);
  const $out = $(".schema-scroll").empty();

  if (!schema || Object.keys(schema).length === 0) {
    return $out.append(`
          <div class="flex flex-col items-center justify-center py-12 text-gray-500">
            <i class="fas fa-database text-5xl mb-4 text-gray-400"></i>
            <p class="text-lg font-medium">No schema available</p>
            <p class="text-sm mt-2">Connect to a database to view its structure</p>
          </div>
        `);
  }

  Object.entries(schema).forEach(([tableName, columns]) => {
    const processedColumns = columns.map((col) => {
      // Determine FK and PK from Go's JSON
      const isFK = Boolean(col.foreign_table.Valid && col.foreign_column.Valid);
      const isPK = Boolean(col.is_primary_key);
      const constraints = [];
      if (isPK) constraints.push("PK");
      if (isFK) constraints.push("FK");

      return {
        name: col.name,
        dataType: col.data_type,
        isPrimaryKey: isPK,
        isForeignKey: isFK,
        constraints,
      };
    });

    const $card = $(`
          <div class="mb-6 bg-white rounded-lg shadow-md border border-gray-200 hover:shadow-lg transition-all duration-300 overflow-hidden">
            <div class="table-header bg-gradient-to-r from-blue-50 to-indigo-50 px-4 py-3 cursor-pointer flex items-center justify-between" data-table="${tableName}">
              <div class="flex items-center space-x-2">
                <i class="fas fa-table text-indigo-600"></i>
                <h3 class="font-bold text-gray-800">${tableName}</h3>
                <span class="bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded-full">${
                  processedColumns.length
                } columns</span>
              </div>
              <i class="fas fa-chevron-down text-gray-400 transition-transform duration-300"></i>
            </div>
            <div class="table-content px-3 pt-2 pb-3 hidden">
              <ul class="divide-y divide-gray-100">
                ${processedColumns
                  .map((col) => {
                    let icon = "fa-cube",
                      color = "text-gray-700 bg-gray-50";
                    switch (col.dataType) {
                      case "uuid":
                        icon = "fa-fingerprint";
                        color = "text-purple-700 bg-purple-50";
                        break;
                      case "integer":
                        icon = "fa-hashtag";
                        color = "text-blue-700 bg-blue-50";
                        break;
                      case "text":
                        icon = "fa-font";
                        color = "text-green-700 bg-green-50";
                        break;
                      case "timestamp":
                        icon = "fa-clock";
                        color = "text-amber-700 bg-amber-50";
                        break;
                      case "boolean":
                        icon = "fa-toggle-on";
                        color = "text-cyan-700 bg-cyan-50";
                        break;
                    }

                    return `
                    <li class="py-2 px-3 hover:bg-gray-50 rounded-md flex items-center justify-between">
                      <div class="flex items-center space-x-2">
                        ${
                          col.isPrimaryKey
                            ? '<i class="fas fa-key text-yellow-500 mr-1"></i>'
                            : ""
                        }
                        ${
                          col.isForeignKey
                            ? '<i class="fas fa-link text-indigo-500 mr-1"></i>'
                            : ""
                        }
                        <span class="font-medium text-gray-800">${
                          col.name
                        }</span>
                      </div>
                      <div class="flex items-center space-x-2">
                        <span class="text-xs px-2 py-1 rounded ${color} flex items-center">
                          <i class="fas ${icon} mr-1"></i>${col.dataType}
                        </span>
                      </div>
                    </li>
                  `;
                  })
                  .join("")}
              </ul>
            </div>
          </div>
        `);

    $out.append($card);
  });
}

// CSS additions to add to your styles
const schemaPanelCSS = `
  .schema-scroll {
    scrollbar-width: thin;
    scrollbar-color: rgba(156, 163, 175, 0.5) transparent;
  }

  .schema-scroll::-webkit-scrollbar {
    width: 6px;
  }

  .schema-scroll::-webkit-scrollbar-track {
    background: transparent;
  }

  .schema-scroll::-webkit-scrollbar-thumb {
    background-color: rgba(156, 163, 175, 0.5);
    border-radius: 20px;
  }

  .table-header:hover i.fa-chevron-down {
    color: #4f46e5;
  }

  /* Animations */
  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .schema-panel .table-content li {
    animation: fadeIn 0.2s ease-out forwards;
  }

  /* Color coding for field types */
  .field-uuid { color: #8b5cf6; }
  .field-text { color: #10b981; }
  .field-number { color: #3b82f6; }
  .field-boolean { color: #f59e0b; }
  .field-timestamp { color: #ef4444; }
  .field-relationship { color: #6366f1; }
  `;

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

// Add this to your existing query.js file

$(function () {
  // Schema search functionality
  $("#schema-search").on("input", function () {
    const searchTerm = $(this).val().toLowerCase();

    if (searchTerm.length === 0) {
      // Show all tables and reset
      $(".table-header").parent().show();
      $(".table-content li").show();
      return;
    }

    // Search through tables and columns
    $(".table-header").each(function () {
      const $tableCard = $(this).parent();
      const tableName = $(this).data("table").toLowerCase();
      const $columns = $(this).next(".table-content").find("li");

      // Check if table name matches
      const tableMatches = tableName.includes(searchTerm);

      // Check if any columns match
      let columnMatches = false;
      $columns.each(function () {
        const columnText = $(this).text().toLowerCase();
        const matches = columnText.includes(searchTerm);

        // Show/hide individual columns
        $(this).toggle(matches);

        if (matches) columnMatches = true;
      });

      // Show table if either table name matches or any columns match
      $tableCard.toggle(tableMatches || columnMatches);

      // If table matches the search term, expand it
      if (tableMatches && !$(this).next(".table-content").is(":visible")) {
        $(this).click();
      }
    });
  });

  // Schema collapse/expand all functionality
  function addSchemaControls() {
    const $controls = $(`
        <div class="flex justify-between px-3 py-2 border-t border-b border-gray-200 bg-gray-50 text-sm">
          <button id="collapse-all-tables" class="text-blue-600 hover:text-blue-800">
            <i class="fas fa-compress-alt mr-1"></i> Collapse All
          </button>
          <button id="expand-all-tables" class="text-blue-600 hover:text-blue-800">
            <i class="fas fa-expand-alt mr-1"></i> Expand All
          </button>
        </div>
      `);

    // Insert after the search box
    $controls.insertAfter($("#schema-search").closest(".relative").parent());

    // Add event handlers
    $("#collapse-all-tables").on("click", function () {
      $(".table-content").slideUp(200);
      $(".table-header i.fa-chevron-down").removeClass("transform rotate-180");
    });

    $("#expand-all-tables").on("click", function () {
      $(".table-content").slideDown(200);
      $(".table-header i.fa-chevron-down").addClass("transform rotate-180");
    });
  }

  // Call this function after schema is loaded
  addSchemaControls();

  // Add tooltips for column information (requires tooltip library)
  function addColumnTooltips() {
    $(".table-content li").each(function () {
      const $column = $(this);
      const columnName = $column.find(".font-medium").text();
      const dataType = $column.find('[class*="text-xs px-2 py-1"]').text();

      // Initialize tooltip with detailed information
      $column.attr("title", `${columnName} (${dataType})`);

      // If using a tooltip library like tippy.js:
      // tippy($column[0], {
      //   content: `<div class="p-2">
      //     <div class="font-bold">${columnName}</div>
      //     <div class="text-sm text-gray-500">${dataType}</div>
      //   </div>`,
      //   allowHTML: true
      // });
    });
  }

  // Add visual enhancements for foreign key relationships
  function enhanceForeignKeyRelationships() {
    // This would require knowledge of the actual relationships in your schema
    // Placeholder for future enhancement
  }

  // Style additions to document
  $("<style>").text(schemaPanelCSS).appendTo("head");
});
