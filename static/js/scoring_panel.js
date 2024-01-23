// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the scoring interface.

var websocket;
let alliance;

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function(data) {
  $("#matchName").text(data.Match.LongName);
  if (alliance === "red") {
    $("#team1").text(data.Match.Red1);
    $("#team2").text(data.Match.Red2);
    $("#team3").text(data.Match.Red3);
  } else {
    $("#team1").text(data.Match.Blue1);
    $("#team2").text(data.Match.Blue2);
    $("#team3").text(data.Match.Blue3);
  }
};

// Handles a websocket message to update the match status.
const handleMatchTime = function(data) {
  switch (matchStates[data.MatchState]) {
    case "PRE_MATCH":
      // Pre-match message state is set in handleRealtimeScore().
      $("#postMatchMessage").hide();
      $("#commitMatchScore").hide();
      break;
    case "POST_MATCH":
      $("#postMatchMessage").hide();
      $("#commitMatchScore").css("display", "flex");
      break;
    default:
      $("#postMatchMessage").hide();
      $("#commitMatchScore").hide();
  }
};

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function(data) {
  let realtimeScore;
  if (alliance === "red") {
    realtimeScore = data.Red;
  } else {
    realtimeScore = data.Blue;
  }
  const score = realtimeScore.Score;

  for (let i = 0; i < 3; i++) {
    const i1 = i + 1;
    $(`#leaveStatus${i1}>.value`).text(score.LeaveStatuses[i] ? "Yes" : "No");
    $("#leaveStatus" + i1).attr("data-value", score.LeaveStatuses[i]);
    $("#trapStatus" + i1).attr("data-value", score.TrapStatuses[i]);
    $("#endgameStatus" + i1 + ">.value").text(getEndgameStatusText(score.EndgameStatuses[i]));
    $("#endgameStatus" + i1).attr("data-value", score.EndgameStatuses[i]);
  }

  $("#coopStatus").attr("data-value", score.Coopertition);
  $("#coopBtn").attr('disabled', score.Coopertition);
  $("#endgameHarmony>.value").text(score.EndgameHarmony ? "Harmony" : "Not Harmony");
  $("#endgameHarmony").attr("data-value", score.EndgameHarmony);

  $(`#autoNoteAmp>.value`).text(score.AutoNoteAmp);
  $(`#autoNoteSpeaker>.value`).text(score.AutoNoteSpeaker);
  $(`#teleopNoteAmp>.value`).text(score.TeleopNoteAmp);
  $(`#teleopNoteSpeaker>.value`).text(score.TeleopNoteSpeaker);
  $(`#teleopNoteAmplifiedSpeaker>.value`).text(score.TeleopNoteAmplifiedSpeaker);
};

// Handles an element click and sends the appropriate websocket message.
const handleClick = function(command, target = 0) {
  websocket.send(command, {Target: target});
};

// Sends a websocket message to indicate that the score for this alliance is ready.
const commitMatchScore = function() {
  websocket.send("commitMatch");
  $("#postMatchMessage").css("display", "flex");
  $("#commitMatchScore").hide();
};

// Returns the display text corresponding to the given integer endgame status value.
const getEndgameStatusText = function(level) {
  switch (level) {
    case 1:
      return "Park";
    case 2:
      return "Onstage";
    case 3:
      return "Spotlit"
    default:
      return "None";
  }
};

$(function() {
  alliance = window.location.href.split("/").slice(-1)[0];
  $("#alliance").attr("data-alliance", alliance);

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/scoring/" + alliance + "/websocket", {
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
  });
});
