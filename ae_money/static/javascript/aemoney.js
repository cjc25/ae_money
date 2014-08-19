function apiUrl(endpoint) {
  return "/api/v0" + endpoint;
}

function toPageFunction(newPage) {
  return function(clickEvent) {
    $("#current > .page").appendTo($("#hidden_pages"));
    $("#" + newPage).appendTo($("#current"));

    // preventDefault if an event was passed in.
    if (typeof clickEvent != "undefined") {
      clickEvent.preventDefault();
    }
  }
}

function toAccountPage(clickEvent) {
  // Migrate the account-specific info we might need.
  $("#account_detail_name").text($(this).text());
  $("#account_detail_name").data("key", $(this).data("key"));
  list_div = $("#account_detail_splits");

  // Kick off a request for the splits.
  $.ajax(apiUrl("/accounts"), {
    type: "GET",
    cache: false,
    data: {key: $(this).data("key")},
    dataType: "json",

    success: function(data) {
      if (data.splits.length == 0) {
        list_div.text("No splits.");
      } else {
        list = $("<ul/>");
        $.each(data.splits, function(i, v) {
          list.append($("<li/>")
            .text(v.amount)
          );
        });

        list_div.html(list);
      }
    }
  });

  // Move to the page immediately while we wait for the splits.
  toPageFunction("account_detail")();
}

function updateAccountsList(sync) {
  list_div = $("#accounts_list");

  // Get the accounts list.
  $.ajax(apiUrl("/accounts"), {
    async: !sync,
    type: "GET",
    dataType: "json",

    success: function(data) {
      if (data.length == 0) {
        list_div.text("No accounts.");
      } else {
        list = $("<ul/>");
        $.each(data, function(i, v) {
          list.append($("<li/>")
            .addClass("account_link")
            .text(v.account.name)
            .data("key", v.key)
            .click(toAccountPage)
          );
        });

        list_div.html(list);
      }
    },
  });
}

function setupRootLinks() {
  $("#root_to_accounts").click(toPageFunction("accounts"));
}

function setupAccountsLinks() {
  // Buttons
  $("#new_account_submit").click(function() {
    submit_button = $(this);
    // Only one creation at a time please.
    submit_button.prop("disabled", true)

    $.ajax(apiUrl("/accounts/new"), {
      type: "POST",
      data: JSON.stringify({
        name: $("#account_creation #new_account_name").val()
      }),
      contentType: "application/json",
      dataType: "json",

      success: function(data) {
        updateAccountsList(true /* synchronous */);
      },
      complete: function() {
        submit_button.prop("disabled", false);
      },
    });
  });
}

function setupAccountDetailLinks() {
  $("#account_detail_to_accounts").click(function(clickEvent) {
    toPageFunction("accounts")();
    $("#account_detail_splits").text("Loading...");
  });
}

$(document).ready(function() {
  // Prepare links.
  setupRootLinks();
  setupAccountsLinks();
  setupAccountDetailLinks();

  // Close enough! Display the front page.
  toPageFunction("root")();

  // Get the accounts list on page load, since the user will probably go there.
  updateAccountsList();
});
