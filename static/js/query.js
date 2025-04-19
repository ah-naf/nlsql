$(function () {
  $("#query-form").submit(function (e) {
    e.preventDefault();
    const nl = $("#nl_query").val();

    function submitQuery(confirmed) {
      $.ajax({
        url: "/query",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify({ nl_query: nl, confirmed }),
        success: function (resp) {
          if (resp.needs_confirmation) {
            $("#modal-sql").text(resp.sql_preview);
            $("#confirm-modal").removeClass("hidden");

            $("#modal-confirm").one("click", function () {
              $("#confirm-modal").addClass("hidden");
              submitQuery(true);
            });
            $("#modal-cancel").one("click", function () {
              $("#confirm-modal").addClass("hidden");
            });
          } else {
            console.log("Final response:", resp);
          }
        },
        error: function (err) {
          console.error("Error sending query:", err);
        },
      });
    }

    submitQuery(false);
  });
});
