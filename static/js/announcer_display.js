// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the announcer display.

var websocket;
let bootstrapRequestId = 0;

// Handles a websocket message to hide the score dialog once the next match is being introduced.
var handleAudienceDisplayMode = function (targetScreen) {
  // Hide the final results so that they aren't blocking the current teams when the announcer needs them most.
  if (targetScreen === "intro" || targetScreen === "match") {
    $("#matchResult").modal("hide");
  }
};

// Handles a websocket message to update the event status message.
const handleEventStatus = function (data) {
  if (data.cycleTime === "") {
    $("#cycleTimeMessage").text("Last cycle time: Unknown");
  } else {
    $("#cycleTimeMessage").text("Last cycle time: " + data.cycleTime);
  }
  $("#earlyLateMessage").text(data.earlyLateMessage);
};

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function (data) {
  $("#matchName").text(data.longName);
  renderAnnouncerMatch(data);
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
  $("#redScore").text(data.red);
  $("#blueScore").text(data.blue);
};

// Handles a websocket message to populate the final score data.
var handleScorePosted = function (data) {
  const matchResult = document.getElementById("matchResult");
  renderAnnouncerScore(matchResult, data);
  const modal = new bootstrap.Modal(matchResult);
  modal.show();

  // Activate tooltips above the foul listings.
  const tooltipTriggerList = document.querySelectorAll("[data-bs-toggle=tooltip]");
  const tooltipList = [...tooltipTriggerList].map(element => new bootstrap.Tooltip(element));
};

const scoreRow = function (label, value, emphasized) {
  const left = $("<div>").addClass("col-sm-6").text(label);
  const right = $("<div>").addClass("col-sm-4").text(value);
  if (emphasized) { left.wrapInner("<b>"); right.wrapInner("<b>"); }
  return $("<div>").addClass("row justify-content-center").append(left, right);
};

const renderScoreAlliance = function (alliance, matchType, color) {
  const card = $("<div>").addClass("card card-body bg-" + color).append($("<h4>").text("Score"));
  const summary = alliance.summary;
  card.append(scoreRow("Auto Fuel Points", summary.autoFuelPoints), scoreRow("Auto Tower Points", summary.autoTowerPoints),
    scoreRow("Teleop Fuel Points", summary.teleopFuelPoints), scoreRow("Teleop Tower Points", summary.teleopTowerPoints), scoreRow("Foul Points", summary.foulPoints));
  if (matchType !== "playoff") {
    card.append(scoreRow("Energized Bonus RP", summary.energizedBonusRankingPoint ? "Yes" : "No"),
      scoreRow("Supercharged Bonus RP", summary.superchargedBonusRankingPoint ? "Yes" : "No"),
      scoreRow("Traversal Bonus RP", summary.traversalBonusRankingPoint ? "Yes" : "No"));
  }
  card.append(scoreRow("Final Score", summary.score, true));
  if (matchType !== "playoff") { card.append(scoreRow("Ranking Points", alliance.rankingPoints, true)); }
  card.append($("<h4>").addClass("mt-3").text("Fouls"));
  alliance.fouls.forEach(function (foul) {
    const kind = (foul.isMajor ? "Major" : "Minor") + " Foul" + (foul.isRankingPoint ? " + RP" : "");
    card.append($("<div>").addClass("row justify-content-center").append(
      $("<div>").addClass("col-sm-4").text(kind), $("<div>").addClass("col-sm-3").text("Team " + foul.teamId),
      $("<div>").addClass("col-sm-3").attr("data-bs-toggle", "tooltip").attr("title", foul.ruleDescription).text(foul.ruleNumber)));
  });
  card.append($("<h4>").addClass("mt-3").text("Cards"));
  alliance.cards.forEach(function (item) { card.append(scoreRow("Team " + item.teamId, item.card)); });
  card.append($("<h4>").addClass("mt-3").text("Rankings"));
  alliance.rankings.forEach(function (item) {
    let value = String(item.rank);
    if (item.rank > item.previousRank && item.previousRank > 0) { value += " ⬇"; }
    else if (item.rank < item.previousRank) { value += " ⬆"; }
    if (item.previousRank > 0) { value += " (was " + item.previousRank + ")"; }
    card.append(scoreRow("Team " + item.teamId, value));
  });
  return card;
};

const renderAnnouncerScore = function (container, score) {
  const content = $("<div>").addClass("modal-content");
  content.append($("<div>").attr("id", "savedMatchResult").addClass("modal-header").append(
    $("<h4>").addClass("modal-title").text("Final Results – " + score.matchName),
    $("<button>").attr("type", "button").attr("data-bs-dismiss", "modal").addClass("btn-close")));
  content.append($("<div>").addClass("modal-body row").append(
    $("<div>").addClass("col-sm-12 mb-3 text-center").append($("<span>").addClass("badge fs-5 " + score.winnerClass).text("Winner: " + score.winner.charAt(0).toUpperCase() + score.winner.slice(1))),
    $("<div>").addClass("col-sm-6").append(renderScoreAlliance(score.red, score.matchType, "red")),
    $("<div>").addClass("col-sm-6").append(renderScoreAlliance(score.blue, score.matchType, "blue"))));
  content.append($("<div>").addClass("modal-footer").append($("<button>").attr("type", "button").attr("data-bs-dismiss", "modal").addClass("btn btn-secondary").text("Close")));
  $(container).empty().append($("<div>").addClass("modal-dialog modal-xl").append(content));
};

const legacyMatchStateIds = {
  pre_match: 0,
  start_match: 1,
  auto: 2,
  pause: 3,
  teleop: 4,
  post_match: 5,
  timeout_active: 6,
  post_timeout: 7,
};

const handleV1Timing = function (data) {
  handleMatchTiming({
    AutoDurationSec: data.autoDurationSec,
    PauseDurationSec: data.pauseDurationSec,
    TransitionShiftDurationSec: data.transitionShiftDurationSec,
    ShiftDurationSec: data.shiftDurationSec,
    EndgameDurationSec: data.endgameDurationSec,
    TimeoutDurationSec: data.timeoutDurationSec,
  });
};

const handleV1MatchClock = function (data) {
  handleMatchTime({MatchState: legacyMatchStateIds[data.state], MatchTimeSec: data.elapsedSec});
};

const applyAnnouncerBootstrap = function (data) {
  handleMatchLoad(data.match);
  handleRealtimeScore(data.realtimeScore);
  handleV1Timing(data.timing);
  handleV1MatchClock(data.matchClock);
  handleEventStatus(data.event);
  handleAudienceDisplayMode(data.audienceDisplayMode);
};

const loadAnnouncerBootstrap = function () {
  const requestId = ++bootstrapRequestId;
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 5000);
  return fetch("/api/v1/displays/announcer/bootstrap", {signal: controller.signal})
    .then(response => {
      if (!response.ok) { throw new Error("Unable to load announcer bootstrap: " + response.status); }
      return response.json();
    })
    .then(response => {
      if (requestId === bootstrapRequestId) { applyAnnouncerBootstrap(response.data); }
    })
    .catch(error => console.error(error))
    .finally(() => clearTimeout(timeout));
};

$(function () {
  loadAnnouncerBootstrap().finally(function () {
    websocket = new CheesyWebsocketV1("/api/v1/streams/displays/announcer", {
      match: function (event) { handleMatchLoad(event.data); },
      postedScore: function (event) {
        if (!event.meta.bootstrap && event.data !== null) { handleScorePosted(event.data); }
      },
      realtimeScore: function (event) { handleRealtimeScore(event.data); },
      timing: function (event) { handleV1Timing(event.data); },
      matchClock: function (event) { handleV1MatchClock(event.data); },
      eventStatus: function (event) { handleEventStatus(event.data); },
      audienceDisplayMode: function (event) { handleAudienceDisplayMode(event.data); },
    }, loadAnnouncerBootstrap);
  });

  // Make the score blink.
  setInterval(function () {
    var blinkOn = $("#savedMatchResult").attr("data-blink") === "true";
    $("#savedMatchResult").attr("data-blink", !blinkOn);
  }, 500);
});
