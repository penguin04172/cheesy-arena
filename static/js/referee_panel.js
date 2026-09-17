// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the referee interface.

var websocket;
var stateWebsocket;
let bootstrapRequestId = 0;
let redFoulsHashCode = 0;
let blueFoulsHashCode = 0;
let scoreIsReady = false;
let isPostMatch = false;
const legacyMatchStateIds = {pre_match: 0, start_match: 1, auto: 2, pause: 3, teleop: 4, post_match: 5, timeout_active: 6, post_timeout: 7};
const legacyRefereeMatch = function (data) { const teams = {}; Object.keys(data.teams || {}).forEach(key => { const team = data.teams[key]; teams[key] = team ? {Id: team.id, YellowCard: team.yellowCard} : null; }); return {Match: {LongName: data.longName}, Teams: teams}; };
const legacyControlScore = data => ({
  Red: {Score: {AutoTowerStatuses: data.red.autoTowerStatuses, EndgameTowerStatuses: data.red.endgameTowerStatuses, Fouls: data.red.fouls.map(foul => ({FoulId: foul.foulId, IsMajor: foul.isMajor, TeamId: foul.teamId, RuleId: foul.ruleId}))}},
  Blue: {Score: {AutoTowerStatuses: data.blue.autoTowerStatuses, EndgameTowerStatuses: data.blue.endgameTowerStatuses, Fouls: data.blue.fouls.map(foul => ({FoulId: foul.foulId, IsMajor: foul.isMajor, TeamId: foul.teamId, RuleId: foul.ruleId}))}},
  RedCards: data.red.cards, BlueCards: data.blue.cards});
const legacyScoringStatus = data => ({RefereeScoreReady: data.refereeScoreReady, PositionStatuses: Object.fromEntries(Object.entries(data.positions || {}).map(([key, value]) => [key, {Ready: value.ready, NumPanels: value.numPanels, NumPanelsReady: value.numPanelsReady}]))});
const legacyStationStatuses = data => ({AllianceStations: Object.fromEntries(Object.entries(data || {}).map(([key, value]) => [key, {Bypass: value.bypass}]))});
const handleV1MatchClock = data => handleMatchTime({MatchState: legacyMatchStateIds[data.state], MatchTimeSec: data.elapsedSec});

// Sends the foul to the server to add it to the list.
const addFoul = function (alliance, isMajor) {
  websocket.send("addFoul", {Alliance: alliance, IsMajor: isMajor});
}

// Toggles the foul type between minor and major.
const toggleFoulType = function (alliance, index) {
  websocket.send("toggleFoulType", {Alliance: alliance, Index: index});
}

// Updates the team that the foul is attributed to.
const updateFoulTeam = function (alliance, index, teamId) {
  websocket.send("updateFoulTeam", {Alliance: alliance, Index: index, TeamId: teamId});
}

// Updates the rule that the foul is for.
const updateFoulRule = function (alliance, index, ruleId) {
  websocket.send("updateFoulRule", {Alliance: alliance, Index: index, RuleId: ruleId});
}

// Removes the foul with the given parameters from the list.
var deleteFoul = function (alliance, index) {
  websocket.send("deleteFoul", {Alliance: alliance, Index: index});
};

// Cycles through the card options for the selected team.
var cycleCard = function (cardButton) {
  if(isPostMatch) {
    // Cycle card.
    const currentCard = $(cardButton).attr("data-card");
    const hasOldYellowCard = $(cardButton).attr("data-old-yellow-card") === "true";
    let newCard = "";
    if (currentCard === "" && hasOldYellowCard) {
      newCard = "red";
    } else if (currentCard === "") {
      newCard = "yellow";
    } else if (currentCard === "yellow") {
      newCard = "red";
    }
    websocket.send(
      "card",
      {Alliance: $(cardButton).attr("data-alliance"), TeamId: parseInt($(cardButton).attr("data-team")), Card: newCard}
    );
    $(cardButton).attr("data-card", newCard);
    return;
  }

  // Toggle bypass.
  const isDisabled = $(cardButton).hasClass("bypassed-status");
  const team = $(cardButton).attr("data-team");
  $("#confirmBypassTitle").text(`${isDisabled ? "Enable" : "Disable"} ${team}?`);
  $("#confirmBypassAction").text(isDisabled ? "Enable" : "Disable")
  $("#confirmBypass").attr("data-station", $(cardButton).attr("data-station")?.toUpperCase());

  if(team === "0") {
    toggleBypass();
  } else {
    $("#confirmBypass").modal("show");
  }
};

const toggleBypass = function() {
  const station = $("#confirmBypass").attr("data-station");
  websocket.send("toggleBypass", station);
}

// Sends a websocket message to signal to the volunteers that they may enter the field.
var signalVolunteers = function () {
  websocket.send("signalVolunteers");
};

// Sends a websocket message to signal to the teams that they may enter the field.
var signalReset = function () {
  websocket.send("signalReset");
};

// Shows confirmation modal if not all scores are ready, otherwise directly commits and posts.
var confirmCommit = function () {
  if (scoreIsReady) {
    commitAndPost();
    return;
  }

  $("#confirmCommit").modal("show");
};

// Commits the score and posts results to the audience.
var commitAndPost = function () {
  websocket.send("commitAndPost");
};

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function (data) {
  $("#matchName").text(data.Match.LongName);

  setTeamCard("red", 1, data.Teams["R1"]);
  setTeamCard("red", 2, data.Teams["R2"]);
  setTeamCard("red", 3, data.Teams["R3"]);
  setTeamCard("blue", 1, data.Teams["B1"]);
  setTeamCard("blue", 2, data.Teams["B2"]);
  setTeamCard("blue", 3, data.Teams["B3"]);

  $("#redScoreSummary .team-1").text(data.Teams["R1"]?.Id || "");
  $("#redScoreSummary .team-2").text(data.Teams["R2"]?.Id || "");
  $("#redScoreSummary .team-3").text(data.Teams["R3"]?.Id || "");
  $("#blueScoreSummary .team-1").text(data.Teams["B1"]?.Id || "");
  $("#blueScoreSummary .team-2").text(data.Teams["B2"]?.Id || "");
  $("#blueScoreSummary .team-3").text(data.Teams["B3"]?.Id || "");
};

// Handles a websocket message to update the match status.
const handleMatchTime = function (data) {
  isPostMatch = matchStates[data.MatchState] === "POST_MATCH";
  $(".control-button").attr("data-enabled", isPostMatch);

  let title = "Red/Yellow Cards";
  if(!isPostMatch) {
    title = matchStates[data.MatchState] === "PRE_MATCH" ? "Bypass" : "Disable";
  }

  $("#teamTitle").text(title)
};

const towerStatusNames = [
  "None",
  "Level 1",
  "Level 2",
  "Level 3",
];

const setTowerStatus = function (selector, status) {
  $(selector).text(towerStatusNames[status]);
  $(selector).attr("data-status", status);
};

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function (data) {
  for (const [teamId, card] of Object.entries(Object.assign(data.RedCards, data.BlueCards))) {
    $(`[data-team="${teamId}"]`).attr("data-card", card);
  }

  const newRedFoulsHashCode = hashObject(data.Red.Score.Fouls);
  const newBlueFoulsHashCode = hashObject(data.Blue.Score.Fouls);
  if (newRedFoulsHashCode !== redFoulsHashCode || newBlueFoulsHashCode !== blueFoulsHashCode) {
    redFoulsHashCode = newRedFoulsHashCode;
    blueFoulsHashCode = newBlueFoulsHashCode;
    fetch("/api/v1/admin/referee/fouls")
      .then(response => {
        if (!response.ok) {
          throw new Error("Unable to load referee fouls: " + response.status);
        }
        return response.json();
      })
      .then(response => renderRefereeFoulList(response.data))
      .catch(error => console.error(error));
  }

  for (alliance of ["red", "blue"]) {
    let score;
    if (alliance === "red") {
      score = data.Red.Score;
    } else {
      score = data.Blue.Score;
    }

    let scoreRoot = `${alliance}ScoreSummary`;
    setTowerStatus(`#${scoreRoot} .team-1-auto-tower`, score.AutoTowerStatuses[0]);
    setTowerStatus(`#${scoreRoot} .team-2-auto-tower`, score.AutoTowerStatuses[1]);
    setTowerStatus(`#${scoreRoot} .team-3-auto-tower`, score.AutoTowerStatuses[2]);
    setTowerStatus(`#${scoreRoot} .team-1-endgame-tower`, score.EndgameTowerStatuses[0]);
    setTowerStatus(`#${scoreRoot} .team-2-endgame-tower`, score.EndgameTowerStatuses[1]);
    setTowerStatus(`#${scoreRoot} .team-3-endgame-tower`, score.EndgameTowerStatuses[2]);
  }
}

const renderRefereeFoulList = function (data) {
  const container = $("#foulList").empty();
  for (const alliance of ["red", "blue"]) {
    const allianceData = data[alliance];
    allianceData.fouls.forEach(function (foul) {
      container.append(renderRefereeFoul(alliance, allianceData.teamIds, foul, data.rules));
    });
  }
};

const renderRefereeFoul = function (alliance, teamIds, foul, rules) {
  const row = $("<div>").addClass("foul " + alliance + "-foul");
  const typeButton = $("<div>").addClass("type-button").text((foul.isMajor ? "Major" : "Minor") + " Foul");
  typeButton.on("click", function () { toggleFoulType(alliance, foul.index); });

  const teamButtons = $("<div>").addClass("team-buttons");
  teamIds.forEach(function (teamId) {
    const button = $("<div>").addClass("team-button").text(teamId);
    if (foul.teamId === teamId) { button.attr("data-selected", "true"); }
    button.on("click", function () { updateFoulTeam(alliance, foul.index, teamId); });
    teamButtons.append(button);
  });

  const ruleSelect = $("<select>").addClass("rule-select");
  ruleSelect.append($("<option>").attr("value", 0).prop("selected", foul.ruleId === 0).text("No Rule Selected"));
  rules.filter(rule => rule.isMajor === foul.isMajor).forEach(function (rule) {
    const foulLabel = (rule.isMajor ? "Major" : "Minor") + " Foul" + (rule.isRankingPoint ? " + RP" : "");
    ruleSelect.append($("<option>").attr("value", rule.id).prop("selected", foul.ruleId === rule.id)
      .text(rule.ruleNumber + " [" + foulLabel + "]: " + rule.description));
  });
  ruleSelect.on("change", function () { updateFoulRule(alliance, foul.index, parseInt(this.value)); });

  const deleteButton = $("<div>").addClass("delete-button").text("Delete");
  deleteButton.on("click", function () { deleteFoul(alliance, foul.index); });
  return row.append($("<div>").text(foul.index + 1), typeButton, teamButtons, ruleSelect, deleteButton);
};

// Handles a websocket message to update the scoring commit status.
const handleScoringStatus = function (data) {
  if (data.RefereeScoreReady) {
    $("#commitButton").attr("data-enabled", false);
  }
  updateScoreStatus(data, "red", "#redScoreStatus", "Red");
  updateScoreStatus(data, "blue", "#blueScoreStatus", "Blue");

  scoreIsReady = Object.values(data.PositionStatuses).every(status => status.Ready);

  // Make the button visually distinct if not all refs have committed.
  // HR can still press the button with confirm modal.
  if (scoreIsReady) {
    $("#commitButton").removeClass("disabled");
  } else {
    $("#commitButton").addClass("disabled");
  }
}

const handleArenaStatus = function (data) {
  setTeamBypassedStatus("red1", data.AllianceStations["R1"]?.Bypass);
  setTeamBypassedStatus("red2", data.AllianceStations["R2"]?.Bypass);
  setTeamBypassedStatus("red3", data.AllianceStations["R3"]?.Bypass);
  setTeamBypassedStatus("blue1", data.AllianceStations["B1"]?.Bypass);
  setTeamBypassedStatus("blue2", data.AllianceStations["B2"]?.Bypass);
  setTeamBypassedStatus("blue3", data.AllianceStations["B3"]?.Bypass);
};

const setTeamBypassedStatus = function (station, bypassed) {
  const cardButton = $(`#${station}Card`);
  cardButton.toggleClass("bypassed-status", bypassed && !isPostMatch);
}

// Helper function to update a badge that shows scoring panel commit status.
const updateScoreStatus = function (data, position, element, displayName) {
  const status = data.PositionStatuses[position];
  $(element).text(`${displayName} ${status.NumPanelsReady}/${status.NumPanels}`);
  $(element).attr("data-present", status.NumPanels > 0);
  $(element).attr("data-ready", status.Ready);
};

// Populates the red/yellow card button for a given team.
const setTeamCard = function (alliance, position, team) {
  const cardButton = $(`#${alliance}${position}Card`);
  if (team === null) {
    cardButton.text("-");
    cardButton.attr("data-team", 0)
    cardButton.attr("data-old-yellow-card", "");
  } else {
    cardButton.text(team.Id);
    cardButton.attr("data-team", team.Id)
    cardButton.attr("data-old-yellow-card", team.YellowCard);
  }
  cardButton.attr("data-card", "");
}

const applyRefereeBootstrap = function (data) {
  handleMatchLoad(legacyRefereeMatch(data.match)); handleV1MatchClock(data.matchClock);
  handleRealtimeScore(legacyControlScore(data.realtimeScore)); handleScoringStatus(legacyScoringStatus(data.scoringStatus));
  handleArenaStatus(legacyStationStatuses(data.stations));
};
const loadRefereeBootstrap = function () {
  const requestId = ++bootstrapRequestId; const controller = new AbortController(); const timeout = setTimeout(() => controller.abort(), 5000);
  return fetch("/api/v1/admin/referee/bootstrap", {signal: controller.signal})
    .then(response => { if (!response.ok) { throw new Error("Unable to load referee bootstrap: " + response.status); } return response.json(); })
    .then(response => { if (requestId === bootstrapRequestId) { applyRefereeBootstrap(response.data); } })
    .catch(error => console.error(error)).finally(() => clearTimeout(timeout));
};

// Produces a hash code of the given object for use in equality comparisons.
const hashObject = function (object) {
  const s = JSON.stringify(object);
  let h = 0;
  for (let i = 0; i < s.length; i++) {
    h = Math.imul(31, h) + s.charCodeAt(i) | 0;
  }
  return h;
}

$(function () {
  // Read the configuration for this display from the URL query string.
  var urlParams = new URLSearchParams(window.location.search);
  $(".headRef-dependent").attr("data-hr", urlParams.get("hr"));

  websocket = new CheesyWebsocket("/panels/referee/websocket", {});
  loadRefereeBootstrap().finally(function () {
    stateWebsocket = new CheesyWebsocketV1("/api/v1/streams/admin/referee", {
    match: function (event) {
      handleMatchLoad(legacyRefereeMatch(event.data));
    },
    matchClock: function (event) {
      handleV1MatchClock(event.data);
    },
    realtimeScore: function (event) {
      handleRealtimeScore(legacyControlScore(event.data));
    },
    scoringStatus: function (event) {
      handleScoringStatus(legacyScoringStatus(event.data));
    },
    stationStatuses: function (event) {
      handleArenaStatus(legacyStationStatuses(event.data));
    },
    }, loadRefereeBootstrap);
  });
});
