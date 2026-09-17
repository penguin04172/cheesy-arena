// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Author: ian@yann.io (Ian Thompson)
//
// Client-side logic for the scoring interface.

var websocket;
var stateWebsocket;
let bootstrapRequestId = 0;
let alliance;
let committed = false;
const legacyMatchStateIds = {pre_match: 0, start_match: 1, auto: 2, pause: 3, teleop: 4, post_match: 5, timeout_active: 6, post_timeout: 7};
const legacyControlScore = data => ({Red: {Score: {AutoTowerStatuses: data.red.autoTowerStatuses, EndgameTowerStatuses: data.red.endgameTowerStatuses,
  Fouls: data.red.fouls.map(foul => ({FoulId: foul.foulId, IsMajor: foul.isMajor, TeamId: foul.teamId, RuleId: foul.ruleId}))}},
Blue: {Score: {AutoTowerStatuses: data.blue.autoTowerStatuses, EndgameTowerStatuses: data.blue.endgameTowerStatuses,
  Fouls: data.blue.fouls.map(foul => ({FoulId: foul.foulId, IsMajor: foul.isMajor, TeamId: foul.teamId, RuleId: foul.ruleId}))}}});
const legacyScoringMatch = data => ({Match: {LongName: data.longName, Red1: data.red[0], Red2: data.red[1], Red3: data.red[2], Blue1: data.blue[0], Blue2: data.blue[1], Blue3: data.blue[2]}});
const handleV1MatchClock = data => handleMatchTime({MatchState: legacyMatchStateIds[data.state], MatchTimeSec: data.elapsedSec});

// True when scoring controls in general should be available
let scoringAvailable = false;
// True when the commit button should be available
let commitAvailable = false;

let localFoulCounts = {
  "red-minor": 0,
  "blue-minor": 0,
  "red-major": 0,
  "blue-major": 0,
}

const foulsDialog = $("#fouls-dialog")[0];
const showFoulsDialog = function () {
  foulsDialog.showModal();
}
const closeFoulsDialog = function () {
  foulsDialog.close();
}
const closeFoulsDialogIfOutside = function (event) {
  if (event.target === foulsDialog) {
    closeFoulsDialog();
  }
}

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function (data) {
  $("#matchName").text(data.Match.LongName);
  if (alliance === "red") {
    $(".team-1 .team-num").text(data.Match.Red1);
    $(".team-2 .team-num").text(data.Match.Red2);
    $(".team-3 .team-num").text(data.Match.Red3);
  } else {
    $(".team-1 .team-num").text(data.Match.Blue1);
    $(".team-2 .team-num").text(data.Match.Blue2);
    $(".team-3 .team-num").text(data.Match.Blue3);
  }
};

const renderLocalFoulCounts = function () {
  for (const foulType in localFoulCounts) {
    const count = localFoulCounts[foulType];
    $(`#foul-${foulType} .fouls-local`).text(count)
  }
}

const renderGlobalFoulCounts = function (redFouls, blueFouls) {
  $(`#foul-blue-minor .fouls-global`).text(blueFouls.filter(foul => !foul.IsMajor).length)
  $(`#foul-blue-major .fouls-global`).text(blueFouls.filter(foul => foul.IsMajor).length)
  $(`#foul-red-minor .fouls-global`).text(redFouls.filter(foul => !foul.IsMajor).length)
  $(`#foul-red-major .fouls-global`).text(redFouls.filter(foul => foul.IsMajor).length)
}

const resetFoulCounts = function () {
  localFoulCounts["red-minor"] = 0;
  localFoulCounts["blue-minor"] = 0;
  localFoulCounts["red-major"] = 0;
  localFoulCounts["blue-major"] = 0;
  renderLocalFoulCounts();
}

const addFoul = function (alliance, isMajor) {
  const foulType = `${alliance}-${isMajor ? "major" : "minor"}`;
  localFoulCounts[foulType] += 1;
  renderLocalFoulCounts();
  websocket.send("addFoul", {Alliance: alliance, IsMajor: isMajor});
}

// Handles a websocket message to update the match status.
const handleMatchTime = function (data) {
  switch (matchStates[data.MatchState]) {
    case "AUTO_PERIOD":
    case "PAUSE_PERIOD":
    case "TELEOP_PERIOD":
      scoringAvailable = true;
      commitAvailable = false;
      committed = false;
      break;
    case "POST_MATCH":
      if (!committed) {
        scoringAvailable = true;
        commitAvailable = true;
      }
      break;
    default:
      scoringAvailable = false;
      commitAvailable = false;
      committed = false;
      resetFoulCounts();
  }
  updateUIMode();
};

// Clear any local ephemeral state that is not maintained by the server
const resetLocalState = function () {
  committed = false;
  updateUIMode();
}

// Refresh which UI controls are enabled/disabled
const updateUIMode = function () {
  $(".scoring-button").prop('disabled', !scoringAvailable);
  $(".scoring-tower-button").prop('disabled', !scoringAvailable);
  $("#commit").prop('disabled', !commitAvailable);
}

const endgameStatusNames = [
  "None",
  "Level 1",
  "Level 2",
  "Level 3",
];

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function (data) {
  let realtimeScore;
  if (alliance === "red") {
    realtimeScore = data.Red;
  } else {
    realtimeScore = data.Blue;
  }
  const score = realtimeScore.Score;

  for (let i = 0; i < 3; i++) {
    const i1 = i + 1;
    for (let j = 0; j < endgameStatusNames.length; j++) {
      $(`#auto-input-${i1} .tower-${j}`).attr("data-selected", j == score.AutoTowerStatuses[i]);
      $(`#endgame-input-${i1} .tower-${j}`).attr("data-selected", j == score.EndgameTowerStatuses[i]);
    }
  }

  const redFouls = data.Red.Score.Fouls || [];
  const blueFouls = data.Blue.Score.Fouls || [];
  renderGlobalFoulCounts(redFouls, blueFouls);
};

// Websocket message senders for various buttons
const handleAutoTowerClick = function (teamPosition, autoTowerStatus) {
  websocket.send("autoTower", {TeamPosition: teamPosition, AutoTowerStatus: autoTowerStatus});
}
const handleEndgameClick = function (teamPosition, endgameTowerStatus) {
  websocket.send("endgame", {TeamPosition: teamPosition, EndgameTowerStatus: endgameTowerStatus});
}

// Sends a websocket message to indicate that the score for this alliance is ready.
const commitMatchScore = function () {
  websocket.send("commitMatch");

  committed = true;
  scoringAvailable = false;
  commitAvailable = false;
  updateUIMode();
};

const applyScoringBootstrap = function (data) {
  resetLocalState(); handleMatchLoad(legacyScoringMatch(data.match)); handleV1MatchClock(data.matchClock); handleRealtimeScore(legacyControlScore(data.realtimeScore));
};
const loadScoringBootstrap = function (position) {
  const requestId = ++bootstrapRequestId; const controller = new AbortController(); const timeout = setTimeout(() => controller.abort(), 5000);
  return fetch("/api/v1/admin/scoring/" + position + "/bootstrap", {signal: controller.signal})
    .then(response => { if (!response.ok) { throw new Error("Unable to load scoring bootstrap: " + response.status); } return response.json(); })
    .then(response => { if (requestId === bootstrapRequestId) { applyScoringBootstrap(response.data); } })
    .catch(error => console.error(error)).finally(() => clearTimeout(timeout));
};

$(function () {
  position = window.location.href.split("/").slice(-1)[0];
  alliance = position;
  $(".container").attr("data-alliance", alliance);
  resetLocalState();

  websocket = new CheesyWebsocket("/panels/scoring/" + position + "/websocket", {resetLocalState: resetLocalState});
  loadScoringBootstrap(position).finally(function () {
    stateWebsocket = new CheesyWebsocketV1("/api/v1/streams/admin/scoring/" + position, {
    match: function (event) {
      handleMatchLoad(legacyScoringMatch(event.data));
    },
    matchClock: function (event) {
      handleV1MatchClock(event.data);
    },
    realtimeScore: function (event) {
      handleRealtimeScore(legacyControlScore(event.data));
    },
    }, function () { return loadScoringBootstrap(position); });
  });
});
