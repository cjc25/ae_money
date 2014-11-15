function apiUrl(/* any number of arguments, encoded and separated by "/" */) {
  result = "/api/v0";
  for (i = 0; i < arguments.length; i++) {
    if (arguments[i]) {
      result += "/" + encodeURIComponent(arguments[i]);
    } else {
      break;
    }
  }
  return result;
}

function toPageFunction(newPage) {
  return function(clickEvent) {
    $(".page.current").addClass("hidden").removeClass("current");
    $("#" + newPage).addClass("current").removeClass("hidden");

    // preventDefault if an event was passed in.
    if (typeof clickEvent != "undefined") {
      clickEvent.preventDefault();
    }
  }
}

function updateAccountsList(sync) {
  list_div = $("#accounts_list");

  // Get the accounts list.
  $.ajax(apiUrl("accounts"), {
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
            .append($("<div/>")
              .addClass("account_link")
              .text(v.account.name)
              .data("key", v.key)
              .click(toAccountPage)
            )
            .append($("<div/>")
              .addClass("total")
              .text(v.account.total)
            )
          );
        });

        list_div.html(list);
      }
    },
  });
}

function toAccountListPage(clickEvent) {
  updateAccountsList();
  toPageFunction("accounts")();
}

function toAccountPage(clickEvent) {
  // Migrate the account-specific info we might need.
  $("#account_detail_name").text($(this).text());
  $("#account_detail_name").data("key", $(this).data("key"));
  list_div = $("#account_detail_splits");

  // Kick off a request for the splits.
  $.ajax(apiUrl("accounts", $(this).data("key")), {
    type: "GET",
    cache: false,
    dataType: "json",

    success: function(data) {
      if (data.splits.length == 0) {
        list_div.text("No splits.");
      } else {
        list = $("<ul/>");
        list.append(
          $("<li/>").addClass("header")
            .append($("<div/>").addClass("date").text("Date"))
            .append($("<div/>").addClass("memo").text("Memo"))
            .append($("<div/>").addClass("amount").text("Amount"))
        );

        $.each(data.splits, function(i, v) {
          line = $("<li/>").addClass("entry");
          line.append($("<div/>")
            .addClass("date")
            .text(v.date.split("T")[0])  // Take the date portion of the time.
          );
          line.append($("<div/>")
            .addClass("memo")
            .text(v.memo)
          );
          line.append($("<div/>")
            .addClass("amount")
            .text(v.amount)
          );
          list.append(line);
        });

        list.append(
          $("<li/>").addClass("total")
            .append($("<div/>").addClass("date").text("Total"))
            .append($("<div/>").addClass("memo"))
            .append($("<div/>").addClass("amount").text(data.account.total))
        );

        list_div.html(list);
      }
    }
  });

  // Move to the page immediately while we wait for the splits.
  toPageFunction("account_detail")();
}


function setupRootLinks() {
  $("#root_to_accounts").click(toPageFunction("accounts"));
}

function setupAccountsLinks() {
  $("#accounts_to_new_transaction").click(function() {
    // Add 2 transaction splits up front.
    addNewTransactionSplit();
    addNewTransactionSplit();

    toPageFunction("new_transaction")();
  });

  // Buttons
  $("#new_account_submit").click(function() {
    submit_button = $(this);
    // Only one creation at a time please.
    submit_button.prop("disabled", true)

    $.ajax(apiUrl("accounts", "new"), {
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

function buildAccountSelector() {
  selector = $("<select/>");
  $("#accounts_list .account_link").each(function() {
    selector.append($("<option/>")
      .text($(this).text())
      .data("key", $(this).data("key"))
    );
  })
  return selector
}

function addNewTransactionSplit() {
  $("#new_transaction_splits").append(
    $("<div/>").append(
      $("<input/>").prop({
        type: "submit",
        class: "new_transaction_remove_split",
        value: "-",
      })
    ).append(
      $("<input/>").prop({
        type: "number",
        class: "new_transaction_amount",
        placeholder: "Amount",
      })
    ).append(
      buildAccountSelector()
    )
  );
}

function newTransactionToAccounts() {
  toAccountListPage();

  $("#new_transaction_memo").val("");
  // Remove the split selectors, in case accounts change.
  $("#new_transaction_splits").children().remove();
}

function submitNewTransaction() {
  submit_button = $(this);
  submit_button.prop("disabled", true);

  request = {amounts: [], accounts: []};
  $("#new_transaction_splits .new_transaction_amount").each(function() {
    request.amounts.push(parseInt($(this).val(), 10));
  });
  $("#new_transaction_splits option:selected").each(function() {
    request.accounts.push($(this).data("key"));
  });
  request.memo = $("#new_transaction_memo").val();
  request.date = $("#new_transaction_date").val();

  $.ajax(apiUrl("transactions", "new"), {
    type: "POST",
    data: JSON.stringify(request),
    contentType: "application/json",

    success: newTransactionToAccounts,
    error: function(jqXHR, textStatus) {
      alert("Failed to commit transaction:\n" + jqXHR.responseText);
    },
    complete: function() {
      submit_button.prop("disabled", false);
    },
  });
}

function setupNewTransactionLinks() {
  $("#new_transaction_splits").on("click", ".new_transaction_remove_split",
    function() {
      $(this).parent().remove()
    }
  );
  $("#new_transaction_add_split").click(addNewTransactionSplit);
  $("#new_transaction_submit").click(submitNewTransaction);
  $("#new_transaction_cancel").click(newTransactionToAccounts);
}

$(document).ready(function() {
  // Prepare links.
  setupRootLinks();
  setupAccountsLinks();

  // Close enough! Display the front page.
  toPageFunction("root")();

  // Get the accounts list on page load, since the user will probably go there.
  updateAccountsList();

  // Links for pages that a user won't get to until later.
  setupAccountDetailLinks();
  setupNewTransactionLinks();
});
