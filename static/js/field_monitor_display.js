// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the field monitor display.

let websocket;
let commandWebsocket;
let bootstrapRequestId = 0;
let currentMatchId;
let redSide;
let blueSide;
const lowBatteryThreshold = 8;
const highBtuThreshold = 7.0;

const legacyMatchStateIds = {pre_match: 0, start_match: 1, auto: 2, pause: 3, teleop: 4, post_match: 5, timeout_active: 6, post_timeout: 7};
const legacyArenaStatus = function (data) {
  const stations = {};
  Object.keys(data.stations || {}).forEach(function (key) {
    const item = data.stations[key];
    stations[key] = {Team: item.team ? {Id: item.team.id, FtaNotes: item.team.ftaNotes || ""} : null,
      DsConn: item.connection ? {WrongStation: item.connection.wrongStation, DsLinked: item.connection.dsLinked,
        RadioLinked: item.connection.radioLinked, RioLinked: item.connection.rioLinked, RobotLinked: item.connection.robotLinked,
        BatteryVoltage: item.connection.batteryVoltage, DsRobotTripTimeMs: item.connection.dsRobotTripTimeMs,
        MissedPacketCount: item.connection.missedPacketCount, SecondsSinceLastRobotLink: item.connection.secondsSinceLastRobotLink} : null,
      WifiStatus: {TeamId: item.wifi.teamId, RadioLinked: item.wifi.radioLinked, MBits: item.wifi.mBits, ConnectionQuality: item.wifi.connectionQuality},
      Ethernet: item.ethernet, AStop: item.aStop, EStop: item.eStop, Bypass: item.bypass};
  });
  return {MatchId: data.matchId, AccessPointStatus: data.accessPointStatus, SwitchStatus: data.switchStatus, AllianceStations: stations};
};
const legacyScoreSummary = function (summary) { const result = {}; Object.keys(summary || {}).forEach(key => { result[key.charAt(0).toUpperCase() + key.slice(1)] = summary[key]; }); return result; };
const legacyRealtimeScore = data => ({Red: {ScoreSummary: legacyScoreSummary(data.red.summary)}, Blue: {ScoreSummary: legacyScoreSummary(data.blue.summary)}});
const handleV1Timing = data => handleMatchTiming({AutoDurationSec: data.autoDurationSec, PauseDurationSec: data.pauseDurationSec,
  TransitionShiftDurationSec: data.transitionShiftDurationSec, ShiftDurationSec: data.shiftDurationSec,
  EndgameDurationSec: data.endgameDurationSec, TimeoutDurationSec: data.timeoutDurationSec});
const handleV1MatchClock = data => handleMatchTime({MatchState: legacyMatchStateIds[data.state], MatchTimeSec: data.elapsedSec});


const handleArenaStatus = function (data) {
  // If getting data for the wrong match (e.g. after a server restart), reload the page.
  if (currentMatchId == null) {
    currentMatchId = data.MatchId;
  } else if (currentMatchId !== data.MatchId) {
    location.reload();
  }

  $.each(data.AllianceStations, function (station, stationStatus) {
    // Select the DOM elements corresponding to the team station.
    let teamElementPrefix;
    if (station[0] === "R") {
      teamElementPrefix = "#" + redSide + "Team" + station[1];
    } else {
      teamElementPrefix = "#" + blueSide + "Team" + station[1];
    }
    const teamIdElement = $(teamElementPrefix + "Id");
    const teamNotesElement = $(teamElementPrefix + "Notes");
    const teamNotesTextElement = $(teamElementPrefix + "Notes div");
    const teamEthernetElement = $(teamElementPrefix + "Ethernet");
    const teamDsElement = $(teamElementPrefix + "Ds");
    const teamRadioElement = $(teamElementPrefix + "Radio");
    const teamRadioIconElement = $(teamElementPrefix + "Radio i");
    const teamRobotElement = $(teamElementPrefix + "Robot");
    const teamBatteryElement = $(teamElementPrefix + "Battery");
    const teamBypassElement = $(teamElementPrefix + "Bypass");
    const teamStatsElement = $(teamElementPrefix + "Stats");
    const teamBandwidthElement = $(teamElementPrefix + "Bandwidth");
    const teamTripTimeElement = $(teamElementPrefix + "TripTime");
    const teamMissedPacketsElement = $(teamElementPrefix + "MissedPackets");

    teamNotesTextElement.attr("data-station", station);

    if (stationStatus.Team) {
      // Set the team number and status.
      teamIdElement.text(stationStatus.Team.Id);
      let status = "no-link";
      if (stationStatus.Bypass) {
        status = "";
      } else if (stationStatus.DsConn) {
        if (stationStatus.DsConn.WrongStation) {
          status = "wrong-station";
        } else if (stationStatus.DsConn.RobotLinked) {
          status = "robot-linked";
        } else if (stationStatus.DsConn.RioLinked) {
          status = "rio-linked";
        } else if (stationStatus.DsConn.RadioLinked) {
          status = "radio-linked";
        } else if (stationStatus.DsConn.DsLinked) {
          status = "ds-linked";
        }
      }
      teamIdElement.attr("data-status", status);
      teamNotesTextElement.text(stationStatus.Team.FtaNotes);
      teamNotesElement.attr("data-status", status);
    } else {
      // No team is present in this position for this match; blank out the status.
      teamIdElement.text("");
      teamNotesTextElement.text("");
      teamNotesElement.attr("data-status", "");
    }

    // Format the Ethernet status box.
    teamEthernetElement.attr("data-status-ok", stationStatus.Ethernet ? "true" : "");
    if (stationStatus.DsConn && stationStatus.DsConn.DsRobotTripTimeMs > 0) {
      teamEthernetElement.text(stationStatus.DsConn.DsRobotTripTimeMs);
    } else {
      teamEthernetElement.text("ETH");
    }

    const wifiStatus = stationStatus.WifiStatus;
    teamRadioIconElement.attr("class", `bi-reception-${wifiStatus.ConnectionQuality}`);

    $("#accessPointStatus").attr("data-status", data.AccessPointStatus);
    $("#switchStatus").attr("data-status", data.SwitchStatus);

    if (stationStatus.DsConn) {
      // Format the driver station status box.
      const dsConn = stationStatus.DsConn;
      teamDsElement.attr("data-status-ok", dsConn.DsLinked);
      teamDsElement.text(dsConn.MissedPacketCount);

      // Format the radio status box according to the connection status of the robot radio.
      const radioOkay = stationStatus.Team && stationStatus.Team.Id === wifiStatus.TeamId &&
        (wifiStatus.RadioLinked || dsConn.RobotLinked);
      teamRadioElement.attr("data-status-ok", radioOkay);

      // Format the robot status box.
      const rioOkay = dsConn.RobotLinked;
      teamRobotElement.attr("data-status-ok", rioOkay);
      if (stationStatus.DsConn.SecondsSinceLastRobotLink > 1 && stationStatus.DsConn.SecondsSinceLastRobotLink < 1000) {
        teamRobotElement.text(stationStatus.DsConn.SecondsSinceLastRobotLink.toFixed());
      } else {
        teamRobotElement.text("RIO");
      }
      const batteryOkay = dsConn.BatteryVoltage > lowBatteryThreshold && dsConn.RobotLinked;
      teamBatteryElement.attr("data-status-ok", batteryOkay);
      teamBatteryElement.text(dsConn.BatteryVoltage.toFixed(1) + "V");

      const btuOkay = wifiStatus.MBits < highBtuThreshold && dsConn.RobotLinked;
      teamStatsElement.attr("data-status-ok", btuOkay);
      if (wifiStatus.MBits >= 0.01) {
        teamBandwidthElement.text(wifiStatus.MBits.toFixed(2));
        teamTripTimeElement.text(dsConn.DsRobotTripTimeMs);
        teamMissedPacketsElement.text(dsConn.MissedPacketCount);
      } else {
        teamBandwidthElement.text("-");
        teamTripTimeElement.text("-");
        teamMissedPacketsElement.text("-");
      }
    } else {
      teamDsElement.attr("data-status-ok", "");
      teamDsElement.text("DS");
      teamRobotElement.attr("data-status-ok", "");
      teamRobotElement.text("RIO");
      teamBatteryElement.text("0.0V");
      teamBandwidthElement.text("-");
      teamTripTimeElement.text("-");
      teamMissedPacketsElement.text("-");

      // Format the robot status box according to whether the AP is configured with the correct SSID.
      const expectedTeamId = stationStatus.Team ? stationStatus.Team.Id : 0;
      if (wifiStatus.TeamId === expectedTeamId) {
        if (wifiStatus.RadioLinked) {
          teamRadioElement.attr("data-status-ok", true);
        } else {
          teamRadioElement.attr("data-status-ok", "");
        }
      } else {
        teamRadioElement.attr("data-status-ok", false);
      }
    }

    if (stationStatus.EStop) {
      teamBypassElement.attr("data-status-ok", false);
      teamBypassElement.text("E-STP");
    } else if (stationStatus.AStop) {
      teamBypassElement.attr("data-status-ok", true);
      teamBypassElement.text("A-STP");
    } else if (stationStatus.Bypass) {
      teamBypassElement.attr("data-status-ok", false);
      teamBypassElement.text("BYP");
    } else {
      teamBypassElement.attr("data-status-ok", true);
      teamBypassElement.text("");
    }
  });
};

// Handles a websocket message to update the match time countdown.
const handleMatchTime = function (data) {
  translateMatchTime(data, function (matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    $("#matchTime").text(countdownSec);
    if (matchStateText === "PRE-MATCH" || matchStateText === "POST-MATCH") {
      $(".ds-dependent").attr("data-preMatch", "true");
    } else {
      $(".ds-dependent").attr("data-preMatch", "false");
    }
  });
};

// Handles a websocket message to update the match score.
const handleRealtimeScore = function (data, reversed) {

  if (reversed === "true") {
    $("#rightScore").text(data.Red.ScoreSummary.Score);
    $("#leftScore").text(data.Blue.ScoreSummary.Score);
  } else {
    $("#rightScore").text(data.Blue.ScoreSummary.Score);
    $("#leftScore").text(data.Red.ScoreSummary.Score);

  }
};

// Handles a websocket message to update current match
const handleMatchLoad = function (data) {
  $("#matchName").text(data.Match.LongName);
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

// Makes the team notes section editable and handles saving edits to the server.
const editFtaNotes = function (element) {
  const teamNotesTextElement = $(element);
  const textArea = $("<textarea />");
  textArea.val(teamNotesTextElement.text());
  teamNotesTextElement.replaceWith(textArea);
  textArea.focus();
  textArea.blur(function () {
    textArea.replaceWith(teamNotesTextElement);
    if (textArea.val() !== teamNotesTextElement.text()) {
      if (commandWebsocket) {
        commandWebsocket.send("updateTeamNotes", {station: teamNotesTextElement.attr("data-station"), notes: textArea.val()});
      }
    }
  });
};

const applyFieldMonitorBootstrap = function (data, reversed) {
  handleArenaStatus(legacyArenaStatus(data.arenaStatus));
  handleEventStatus({CycleTime: data.event.cycleTime, EarlyLateMessage: data.event.earlyLateMessage});
  handleMatchLoad({Match: {LongName: data.match.longName}});
  handleV1Timing(data.timing); handleV1MatchClock(data.matchClock);
  handleRealtimeScore(legacyRealtimeScore(data.realtimeScore), reversed);
};

const loadFieldMonitorBootstrap = function (isFta, reversed) {
  const requestId = ++bootstrapRequestId; const controller = new AbortController(); const timeout = setTimeout(() => controller.abort(), 5000);
  return fetch("/api/v1/displays/field-monitor/bootstrap?fta=" + isFta, {signal: controller.signal})
    .then(response => { if (!response.ok) { throw new Error("Unable to load field monitor bootstrap: " + response.status); } return response.json(); })
    .then(response => { if (requestId === bootstrapRequestId) { applyFieldMonitorBootstrap(response.data, reversed); } })
    .catch(error => console.error(error)).finally(() => clearTimeout(timeout));
};

$(function () {
  // Read the configuration for this display from the URL query string.
  const urlParams = new URLSearchParams(window.location.search);
  const reversed = urlParams.get("reversed");
  if (reversed === "true") {
    redSide = "right";
    blueSide = "left";
  } else {
    redSide = "left";
    blueSide = "right";
  }

  //Read if display to be used in a Driver Station, ignore FTA flag if so.
  const driverStation = urlParams.get("ds");
  const isFta = driverStation !== "true" && urlParams.get("fta") === "true";
  if (driverStation === "true") {
    $(".fta-dependent").attr("data-fta", "false");
    $(".ds-dependent").attr("data-ds", driverStation);
  } else {
    $(".fta-dependent").attr("data-fta", urlParams.get("fta"));
    $(".ds-dependent").attr("data-ds", driverStation);
  }

  $(".reversible-left").attr("data-reversed", reversed);
  $(".reversible-right").attr("data-reversed", reversed);


  const isFms = window.location.pathname.indexOf("fms_field_monitor") !== -1;
  loadFieldMonitorBootstrap(isFta, reversed).finally(function () {
    websocket = new CheesyWebsocketV1("/api/v1/streams/displays/field-monitor?fta=" + isFta + "&fms=" + isFms, {
    arenaStatus: function (event) {
      handleArenaStatus(legacyArenaStatus(event.data));
    },
    eventStatus: function (event) {
      handleEventStatus({CycleTime: event.data.cycleTime, EarlyLateMessage: event.data.earlyLateMessage});
    },
    match: function (event) {
      handleMatchLoad({Match: {LongName: event.data.longName}});
    },
    timing: function (event) {
      handleV1Timing(event.data);
    },
    matchClock: function (event) {
      handleV1MatchClock(event.data);
    },
    realtimeScore: function (event) {
      handleRealtimeScore(legacyRealtimeScore(event.data), reversed);
    },
    }, function () { return loadFieldMonitorBootstrap(isFta, reversed); });
  });

  if (isFta) {
    const commandPath = isFms ? "/displays/fms_field_monitor/websocket" : "/displays/field_monitor/websocket";
    commandWebsocket = new CheesyWebsocket(commandPath, {});
  }
});
