// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for the alliance station display.

var station = "";
var blinkInterval;
var currentScreen = "blank";
var websocket;
var bootstrapRequestId = 0;

var legacyMatchStateIds = {pre_match: 0, start_match: 1, auto: 2, pause: 3, teleop: 4, post_match: 5, timeout_active: 6, post_timeout: 7};
var legacyMatchType = function (type) {
  if (type === "qualification") { return matchTypeQualification; }
  if (type === "playoff") { return matchTypePlayoff; }
  return type === "practice" ? 1 : 0;
};
var legacyMatch = function (data) {
  var teams = {};
  Object.keys(data.teams || {}).forEach(function (key) {
    var team = data.teams[key]; teams[key] = team ? {Id: team.id, Nickname: team.nickname} : null;
  });
  return {Match: {Type: legacyMatchType(data.type), PlayoffRedAlliance: data.playoffRedAlliance, PlayoffBlueAlliance: data.playoffBlueAlliance},
    Teams: teams, Rankings: data.rankings || {},
    RedOffFieldTeams: (data.redOffFieldTeams || []).map(function (team) { return {Id: team.id}; }),
    BlueOffFieldTeams: (data.blueOffFieldTeams || []).map(function (team) { return {Id: team.id}; })};
};
var legacyStationStatuses = function (stations) {
  var result = {AllianceStations: {}};
  Object.keys(stations || {}).forEach(function (key) {
    var item = stations[key]; result.AllianceStations[key] = {Bypass: item.bypass, DsConn: {DsLinked: item.dsLinked, RobotLinked: item.robotLinked}};
  });
  return result;
};
var handleV1Timing = function (data) { handleMatchTiming({AutoDurationSec: data.autoDurationSec, PauseDurationSec: data.pauseDurationSec,
  TransitionShiftDurationSec: data.transitionShiftDurationSec, ShiftDurationSec: data.shiftDurationSec,
  EndgameDurationSec: data.endgameDurationSec, TimeoutDurationSec: data.timeoutDurationSec}); };
var handleV1MatchClock = function (data) { handleMatchTime({MatchState: legacyMatchStateIds[data.state], MatchTimeSec: data.elapsedSec}); };

// Handles a websocket message to change which screen is displayed.
var handleAllianceStationDisplayMode = function (targetScreen) {
  currentScreen = targetScreen;
  if (station === "") {
    // Don't do anything if this screen hasn't been assigned a position yet.
  } else {
    var body = $("body");
    body.attr("data-mode", targetScreen);
    if (targetScreen === "timeout") {
      body.attr("data-position", "middle");
    } else {
      switch (station[1]) {
        case "1":
          body.attr("data-position", "right");
          break;
        case "2":
          body.attr("data-position", "middle");
          break;
        case "3":
          body.attr("data-position", "left");
          break;
      }
    }
  }
};

// Handles a websocket message to update the team to display.
var handleMatchLoad = function (data) {
  if (station !== "") {
    var team = data.Teams[station];
    if (team) {
      $("#teamNumber").text(team.Id);
      $("#teamNameText").attr("data-alliance-bg", station[0]).text(team.Nickname);

      var ranking = data.Rankings[team.Id];
      if (ranking && data.Match.Type === matchTypeQualification) {
        $("#teamRank").attr("data-alliance-bg", station[0]).text(ranking);
      } else {
        $("#teamRank").attr("data-alliance-bg", station[0]).text("");
      }
    } else {
      $("#teamNumber").text("");
      $("#teamNameText").attr("data-alliance-bg", station[0]).text("");
      $("#teamRank").attr("data-alliance-bg", station[0]).text("");
    }

    // Populate extra alliance info if this is a playoff match.
    let playoffAlliance = data.Match.PlayoffRedAlliance;
    let offFieldTeams = data.RedOffFieldTeams;
    if (station[0] === "B") {
      playoffAlliance = data.Match.PlayoffBlueAlliance;
      offFieldTeams = data.BlueOffFieldTeams;
    }
    if (playoffAlliance > 0) {
      let playoffAllianceInfo = `Alliance ${playoffAlliance}`;
      if (offFieldTeams.length) {
        playoffAllianceInfo += `&emsp; Not on field: ${offFieldTeams.map(team => team.Id).join(", ")}`;
      }
      $("#playoffAllianceInfo").html(playoffAllianceInfo);
    } else {
      $("#playoffAllianceInfo").text("");
    }
  }
};

// Handles a websocket message to update the team connection status.
var handleArenaStatus = function (data) {
  stationStatus = data.AllianceStations[station];
  var blink = false;
  if (stationStatus && stationStatus.Bypass) {
    $("#match").attr("data-status", "bypass");
  } else if (stationStatus) {
    if (!stationStatus.DsConn || !stationStatus.DsConn.DsLinked) {
      $("#match").attr("data-status", station[0]);
    } else if (!stationStatus.DsConn.RobotLinked) {
      blink = true;
      if (!blinkInterval) {
        blinkInterval = setInterval(function () {
          var status = $("#match").attr("data-status");
          $("#match").attr("data-status", (status === "") ? station[0] : "");
        }, 250);
      }
    } else {
      $("#match").attr("data-status", "");
    }
  }

  if (!blink && blinkInterval) {
    clearInterval(blinkInterval);
    blinkInterval = null;
  }
};

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function (data) {
  translateMatchTime(data, function (matchState, matchStateText, countdownSec) {
    if (station[0] === "N") {
      // Pin the state for a non-alliance display to an in-match state, so as to always show time or score.
      matchState = "TELEOP_PERIOD";
    }
    var countdownString = String(countdownSec % 60);
    if (countdownString.length === 1) {
      countdownString = "0" + countdownString;
    }
    countdownString = Math.floor(countdownSec / 60) + ":" + countdownString;
    $("#timeRemaining").text(countdownString);
    $("#match").attr("data-state", matchState);
  });
};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function (data) {
  $("#redScore").text(data.red);
  $("#blueScore").text(data.blue);
};

var applyAllianceStationBootstrap = function (data) {
  handleMatchLoad(legacyMatch(data.match)); handleRealtimeScore(data.realtimeScore); handleV1Timing(data.timing);
  handleV1MatchClock(data.matchClock); handleAllianceStationDisplayMode(data.displayMode); handleArenaStatus(legacyStationStatuses(data.stations));
};

var loadAllianceStationBootstrap = function () {
  var requestId = ++bootstrapRequestId;
  var controller = new AbortController(); var timeout = setTimeout(function () { controller.abort(); }, 5000);
  return fetch("/api/v1/displays/alliance-station/bootstrap", {signal: controller.signal})
    .then(function (response) { if (!response.ok) { throw new Error("Unable to load alliance station bootstrap: " + response.status); } return response.json(); })
    .then(function (response) { if (requestId === bootstrapRequestId) { applyAllianceStationBootstrap(response.data); } })
    .catch(function (error) { console.error(error); }).finally(function () { clearTimeout(timeout); });
};

$(function () {
  // Read the configuration for this display from the URL query string.
  var urlParams = new URLSearchParams(window.location.search);
  station = urlParams.get("station");

  loadAllianceStationBootstrap().finally(function () {
    websocket = new CheesyWebsocketV1("/api/v1/streams/displays/alliance-station", {
    allianceStationDisplayMode: function (event) {
      handleAllianceStationDisplayMode(event.data);
    },
    stationStatuses: function (event) {
      handleArenaStatus(legacyStationStatuses(event.data));
    },
    match: function (event) {
      handleMatchLoad(legacyMatch(event.data));
    },
    matchClock: function (event) {
      handleV1MatchClock(event.data);
    },
    timing: function (event) {
      handleV1Timing(event.data);
    },
    realtimeScore: function (event) {
      handleRealtimeScore(event.data);
    }
    }, loadAllianceStationBootstrap);
  });
});
