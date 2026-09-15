// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the announcer display.

var websocket;
let isFirstScorePosted = true;

// Handles a websocket message to hide the score dialog once the next match is being introduced.
var handleAudienceDisplayMode = function (targetScreen) {
  // Hide the final results so that they aren't blocking the current teams when the announcer needs them most.
  if (targetScreen === "intro" || targetScreen === "match") {
    $("#matchResult").modal("hide");
  }
};

// Handles a websocket message to update the event status message.
const handleEventStatus = function (data) {
  if (data.CycleTime === "") {
    $("#cycleTimeMessage").text("Last cycle time: Unknown");
  } else {
    $("#cycleTimeMessage").text("Last cycle time: " + data.CycleTime);
  }
  $("#earlyLateMessage").text(data.EarlyLateMessage);
};

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function (data) {
  $("#matchName").text(data.Match.LongName);

  const teams = $("#teams");
  teams.empty();

  fetch("/api/v1/displays/announcer/match")
    .then(response => {
      if (!response.ok) {
        throw new Error("Unable to load announcer match data: " + response.status);
      }
      return response.json();
    })
    .then(response => renderAnnouncerMatch(response.data))
    .catch(error => console.error(error));
};

const renderAnnouncerTeam = function (team) {
  const row = $("<div>").addClass("row");
  if (team === null) {
    return row.append($("<div>").addClass("col-sm-12").append($("<h3>").append($("<b>").text("No team present"))));
  }

  const number = $("<h2>").append($("<b>").text(team.id));
  if (team.isOffField) {
    number.append($("<span>").css("font-size", "0.5em").text(" (not on field)"));
  }
  const detailsId = "team" + team.id + "Details";
  const moreButton = $("<button>").attr("type", "button").addClass("btn btn-secondary btn-sm").text("More");
  moreButton.on("click", function () { $("#" + detailsId).modal("show"); });
  row.append(
    $("<div>").addClass("col-sm-2").append(number),
    $("<div>").addClass("col-sm-4").append($("<h2>").text(team.nickname)),
    $("<div>").addClass("col-sm-2").append($("<h5>").text(team.schoolName)),
    $("<div>").addClass("col-sm-3").append($("<div>").append($("<h5>").text([team.city, team.stateProv, team.country].join(", ")))),
    $("<div>").addClass("col-sm-1").append($("<div>").addClass("row").append(
      $("<div>").addClass("col-sm-6").text(team.rank === null ? "" : team.rank),
      $("<div>").addClass("col-sm-6").append(moreButton),
    )),
  );

  const modalBody = $("<div>").addClass("modal-body").append(
    $("<div>").addClass("mb-3").append($("<b>").text("Rookie Year: "), document.createTextNode(team.rookieYear)),
    $("<div>").addClass("mb-3").append($("<b>").text("Robot Name: "), document.createTextNode(team.robotName)),
    $("<div>").addClass("mb-1").append($("<b>").text("Recent Accomplishments:")),
    $("<div>").text(team.accomplishments),
  );
  row.append(
    $("<div>").attr("id", detailsId).addClass("modal").append(
      $("<div>").addClass("modal-dialog").append($("<div>").addClass("modal-content").append(
        $("<div>").addClass("modal-header").append(
          $("<h4>").addClass("modal-title").text("Team " + team.id),
          $("<button>").attr("type", "button").attr("data-bs-dismiss", "modal").addClass("btn-close"),
        ),
        modalBody,
        $("<div>").addClass("modal-footer").append(
          $("<button>").attr("type", "button").attr("data-bs-dismiss", "modal").addClass("btn btn-secondary").text("Close"),
        ),
      )),
    ),
  );
  return row;
};

const renderAnnouncerAlliance = function (alliance, color, isPlayoff) {
  const card = $("<div>").addClass("row card card-body bg-" + color);
  if (isPlayoff) {
    card.append($("<h4>").append($("<b>").text("Alliance " + alliance.playoffAllianceId)));
  }
  alliance.teams.forEach(function (team) {
    card.append(renderAnnouncerTeam(team));
  });
  return card;
};

const renderAnnouncerMatch = function (match) {
  const isPlayoff = match.type === "playoff";
  $("#teams").empty().append(
    renderAnnouncerAlliance(match.red, "red", isPlayoff),
    renderAnnouncerAlliance(match.blue, "blue", isPlayoff),
  );
};

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function (data) {
  translateMatchTime(data, function (matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    $("#matchTime").text(getCountdown(data.MatchState, data.MatchTimeSec));
  });
};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function (data) {
  $("#redScore").text(data.Red.ScoreSummary.Score - data.Red.ScoreSummary.PostMatchPoints);
  $("#blueScore").text(data.Blue.ScoreSummary.Score - data.Blue.ScoreSummary.PostMatchPoints);
};

// Handles a websocket message to populate the final score data.
var handleScorePosted = function (data) {
  if (isFirstScorePosted) {
    // Don't show the final score dialog when the page is first loaded.
    isFirstScorePosted = false;
    return;
  }

  const matchResult = document.getElementById("matchResult");
  fetch("/displays/announcer/score_posted")
    .then(response => response.text())
    .then(html => {
      matchResult.innerHTML = html;
      const modal = new bootstrap.Modal(matchResult);
      modal.show();

      // Activate tooltips above the foul listings.
      const tooltipTriggerList = document.querySelectorAll("[data-bs-toggle=tooltip]");
      const tooltipList = [...tooltipTriggerList].map(element => new bootstrap.Tooltip(element));
    });
};

$(function () {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/announcer/websocket", {
    audienceDisplayMode: function (event) {
      handleAudienceDisplayMode(event.data);
    },
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
    realtimeScore: function (event) {
      handleRealtimeScore(event.data);
    },
    scorePosted: function (event) {
      handleScorePosted(event.data);
    }
  });

  // Make the score blink.
  setInterval(function () {
    var blinkOn = $("#savedMatchResult").attr("data-blink") === "true";
    $("#savedMatchResult").attr("data-blink", !blinkOn);
  }, 500);
});
