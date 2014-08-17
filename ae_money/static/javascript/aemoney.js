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
	alert($(this).data("key"));
}

function updateAccountsList(async) {
  list_div = $("#accounts_list");

  // Get the accounts list.
  $.ajax(apiUrl("/accounts"), {
		async: async,
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
					  .click(toAccountPage));
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
        console.log(data.key + " " + data.name);
        updateAccountsList(false);
      },
			complete: function() {
			  submit_button.prop("disabled", false);
		  },
		});
	});
}

$(document).ready(function() {
  // Prepare links.
	setupRootLinks();
	setupAccountsLinks();

  // Close enough! Display the front page.
  toPageFunction("root")();

  // Get the accounts list on page load, since the user will probably go there.
  updateAccountsList(true);
});
