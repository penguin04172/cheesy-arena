// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the queueing display.

var websocket;

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function (data) {
  fetch("/api/v1/displays/queueing/matches")
    .then(response => {
      if (!response.ok) {
        throw new Error("Unable to load queueing matches: " + response.status);
      }
      return response.json();
    })
    .then(response => renderMatches(response.data))
    .catch(error => console.error(error));
};

var renderTeamAvatars = function (teamIds) {
  var avatars = $("<div>").addClass("col-lg-1 avatars");
  teamIds.forEach(function (teamId, index) {
    if (index > 0) {
      avatars.append($("<br>"));
    }
    avatars.append(
      $("<img>")
        .addClass("avatar")
        .attr("src", "/api/v1/teams/" + teamId + "/avatar")
        .attr("alt", "Team " + teamId + " avatar"),
    );
  });
  return avatars;
};

var renderTeamNumbers = function (alliance, color) {
  var container = $("<div>").addClass("col-lg-2 " + color + "-teams");
  if (color === "blue") {
    container.addClass("text-end");
  }
  if (alliance.teamIds.length === 0) {
    return container;
  }

  var row = $("<div>").addClass("row");
  var teams = $("<div>").addClass("col-lg-8");
  alliance.teamIds.concat(alliance.offFieldTeamIds).forEach(function (teamId, index) {
    if (index > 0) {
      teams.append($("<br>"));
    }
    teams.append(document.createTextNode(teamId));
  });

  var allianceNumber = $("<div>").addClass("col-lg-4");
  if (alliance.playoffAllianceId) {
    var badge = $("<div>").addClass("alliance-container");
    if (alliance.offFieldTeamIds.length > 0) {
      badge.addClass("alliance-tall");
    }
    if (color === "blue") {
      badge.append($("<div>"));
    }
    badge.append($("<div>").addClass("alliance-number").text(alliance.playoffAllianceId));
    allianceNumber.append(badge);
  }
  if (color === "blue") {
    row.append(allianceNumber, teams);
  } else {
    row.append(teams, allianceNumber);
  }
  return container.append(row);
};

var renderMatches = function (matches) {
  var matchesContainer = $("#matches").empty();
  matches.forEach(function (match, index) {
    var details = $("<div>").addClass("col-lg-6");
    details.append(
      $("<div>").addClass("row").append(
        $("<div>").addClass("col-lg-4 ps-4").append($("<h1>").addClass("mt-2").text(match.positionLabel)),
        $("<div>").addClass("col-lg-3").append($("<h1>").addClass("mt-2").text(match.shortName)),
        $("<div>").addClass("col-lg-5").append($("<h1>").addClass("mt-2").text(match.displayTime)),
      ),
    );
    if (index === 0) {
      details.append(
        $("<div>").addClass("row mt-3").append(
          $("<div>").attr("id", "matchState").addClass("col-lg-4 ps-4"),
          $("<div>").attr("id", "matchTime").addClass("col-lg-3"),
        ),
      );
    }

    var redAvatars = renderTeamAvatars(match.red.teamIds).addClass("text-end");
    var cardRow = $("<div>").addClass("row").append(
      details,
      redAvatars,
      renderTeamNumbers(match.red, "red"),
      renderTeamNumbers(match.blue, "blue"),
      renderTeamAvatars(match.blue.teamIds),
    );
    matchesContainer.append(
      $("<div>").addClass("row justify-content-center").append(
        $("<div>").addClass("col-lg-10").append(
          $("<div>").addClass("card card-body").append(cardRow),
        ),
      ),
    );
  });
};

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function (data) {
  translateMatchTime(data, function (matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    var countdownString = String(countdownSec % 60);
    if (countdownString.length === 1) {
      countdownString = "0" + countdownString;
    }
    countdownString = Math.floor(countdownSec / 60) + ":" + countdownString;
    $("#matchTime").text(countdownString);
  });
};

// Handles a websocket message to update the event status message.
var handleEventStatus = function (data) {
  $("#earlyLateMessage").text(data.EarlyLateMessage);
};

$(function () {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/queueing/websocket", {
    eventStatus: function (event) {
      handleEventStatus(event.data);
    },
    matchLoad: function (event) {
      handleMatchLoad(event.data);
    },
    matchTime: function (event) {
      handleMatchTime(event.data);
    },
    matchTiming: function (event) {
      handleMatchTiming(event.data);
    },
  });
});
