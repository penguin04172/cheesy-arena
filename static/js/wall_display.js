// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for the wall display.

if (typeof DisplayShared === "undefined") {
  $.ajax({async: false, cache: true, dataType: "script", url: "/static/js/display_shared.js"});
}

var websocket;
let bootstrapRequestId = 0;
let transitionMap;
const transitionQueue = [];
let transitionInProgress = false;
let currentScreen = "blank";
let redSide;
let blueSide;
let currentMatch;
let messageText = "";
let hasMessage = false;
const hubActiveController = DisplayShared.createHubActiveController(function () {
  return currentScreen;
});

const legacyMatchStateIds = {pre_match: 0, start_match: 1, auto: 2, pause: 3, teleop: 4, post_match: 5, timeout_active: 6, post_timeout: 7};
const legacyMatchType = type => type === "qualification" ? matchTypeQualification : (type === "playoff" ? matchTypePlayoff : (type === "practice" ? 1 : 0));
const legacyScoreSummary = function (summary) {
  const result = {}; Object.keys(summary || {}).forEach(key => { result[key.charAt(0).toUpperCase() + key.slice(1)] = summary[key]; }); return result;
};
const legacyMatch = function (data) {
  const teams = {};
  Object.keys(data.teams || {}).forEach(station => { const team = data.teams[station]; teams[station] = team ? {Id: team.id, Nickname: team.nickname, YellowCard: team.yellowCard} : null; });
  return {Match: {Id: data.id, Type: legacyMatchType(data.type), LongName: data.longName, NameDetail: data.nameDetail,
    PlayoffRedAlliance: data.playoffRedAlliance, PlayoffBlueAlliance: data.playoffBlueAlliance,
    Red1: data.red[0], Red2: data.red[1], Red3: data.red[2], Blue1: data.blue[0], Blue2: data.blue[1], Blue3: data.blue[2]},
  Teams: teams, Rankings: data.rankings || {}, Matchup: data.series ? {NumWinsToAdvance: data.series.numWinsToAdvance,
    RedAllianceWins: data.series.redAllianceWins, BlueAllianceWins: data.series.blueAllianceWins} : null,
  BreakDescription: data.breakDescription, BreakNextMatchName: data.breakNextMatchName};
};
const legacyRealtimeScore = data => ({Red: {ScoreSummary: legacyScoreSummary(data.red.summary), ActiveRemainingSec: data.red.activeRemainingSec, ActiveDurationSec: data.red.activeDurationSec},
  Blue: {ScoreSummary: legacyScoreSummary(data.blue.summary), ActiveRemainingSec: data.blue.activeRemainingSec, ActiveDurationSec: data.blue.activeDurationSec}});
const handleV1Timing = data => handleMatchTiming({AutoDurationSec: data.autoDurationSec, PauseDurationSec: data.pauseDurationSec,
  TransitionShiftDurationSec: data.transitionShiftDurationSec, ShiftDurationSec: data.shiftDurationSec,
  EndgameDurationSec: data.endgameDurationSec, TimeoutDurationSec: data.timeoutDurationSec});
const handleV1MatchClock = data => handleMatchTime({MatchState: legacyMatchStateIds[data.state], MatchTimeSec: data.elapsedSec});

// Constants for overlay positioning. The CSS is the source of truth for the values that represent initial state.
const eventMatchInfoDown = "30px";
const eventMatchInfoUp = $("#eventMatchInfo").css("height");
const logoUp = "35px";
const logoDown = $("#logo").css("top");
const scoreIn = $(".score").css("width");
const scoreMid = "185px";
const scoreOut = "250px";
const scoreFieldsOut = "25px";
const overlayTopOffset = 110;
const timeoutDetailsIn = $("#timeoutDetails").css("width");
const timeoutDetailsOut = "570px";

// Handles a websocket message to change which screen is displayed.
const handleAudienceDisplayMode = function (targetScreen) {
  if (targetScreen === "logoLuma") {
    targetScreen = "logo";
  }
  if (
    targetScreen !== "intro" &&
    targetScreen !== "match" &&
    targetScreen !== "timeout" &&
    targetScreen !== "logo"
  ) {
    targetScreen = "blank";
  }

  transitionQueue.push(targetScreen);
  executeTransitionQueue();
};

// Sequentially executes all transitions in the queue. Returns without doing anything if another invocation is already
// in progress.
const executeTransitionQueue = function () {
  if (transitionInProgress) {
    // There is an existing invocation of this method which will execute all transitions in the queue.
    return;
  }

  if (transitionQueue.length > 0) {
    transitionInProgress = true;
    const targetScreen = transitionQueue.shift();
    const callback = function () {
      // When the current transition is complete, call this method again to invoke the next one in the queue.
      currentScreen = targetScreen;
      transitionInProgress = false;
      setTimeout(executeTransitionQueue, 100);  // A small delay is needed to avoid visual glitches.
    };

    if (targetScreen === currentScreen) {
      callback();
      return;
    }

    let transitions = transitionMap[currentScreen][targetScreen];
    if (transitions !== undefined) {
      transitions(callback);
    } else {
      // There is no direct transition defined; need to go to the blank screen first.
      transitionMap[currentScreen]["blank"](function () {
        transitionMap["blank"][targetScreen](callback);
      });
    }
  }
};

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function (data) {
  currentMatch = DisplayShared.handleMatchLoad(data, redSide, blueSide);
};

// Handles a websocket message to update the match time countdown.
const handleMatchTime = function (data) {
  DisplayShared.handleMatchTime(data);
};

// Handles a websocket message to update the match score.
const handleRealtimeScore = function (data) {
  DisplayShared.handle2026RealtimeScore(
    data,
    currentMatch,
    redSide,
    blueSide,
    hubActiveController.updateHubActiveIndicator
  );
};

const transitionBlankToIntro = function (callback) {
  hideMessage(function () {
    $(".teams").css("display", "flex");
    $(".avatars").css("display", "flex");
    $(".avatars").css("opacity", 1);
    $(".score").transition({queue: false, width: scoreMid}, 500, "ease", function () {
      $("#eventMatchInfo").css("display", "flex");
      $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoDown}, 500, "ease", callback);
    });
  });
};

const transitionBlankToLogo = function (callback) {
  showMessage(callback);
}

const transitionBlankToMatch = function (callback) {
  hideMessage(function () {
    $(".teams").css("display", "flex");
    $(".score-fields").css("display", "flex");
    $(".score-fields").transition({queue: false, width: scoreFieldsOut}, 500, "ease");
    $("#logo").transition({queue: false, top: logoUp}, 500, "ease");
    $(".score").transition({queue: false, width: scoreOut}, 500, "ease", function () {
      $("#eventMatchInfo").css("display", "flex");
      $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoDown}, 500, "ease", callback);
      $(".score-number").transition({queue: false, opacity: 1}, 750, "ease");
      $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease");
      $(".score-fields").transition({queue: false, opacity: 1}, 750, "ease");
      $(".score-aux").transition({queue: false, opacity: 1}, 750, "ease");
      hubActiveController.restartPendingHubActiveIndicators();
    });
  });
};

const transitionBlankToTimeout = function (callback) {
  hideMessage(function () {
    $("#timeoutDetails").transition({queue: false, width: timeoutDetailsOut}, 500, "ease");
    $("#logo").transition({queue: false, top: logoUp}, 500, "ease", function () {
      $(".timeout-detail").transition({queue: false, opacity: 1}, 750, "ease");
      $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
    });
  });
};

const transitionIntroToBlank = function (callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease", function () {
    $("#eventMatchInfo").hide();
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function () {
      $(".avatars").css("opacity", 0);
      $(".avatars").hide();
      $(".teams").hide();
      showMessage(callback);
    });
  });
};

const transitionIntroToMatch = function (callback) {
  $(".avatars").transition({queue: false, opacity: 0}, 500, "ease", function () {
    $(".avatars").hide();
  });
  $(".score-fields").css("display", "flex");
  $(".score-fields").transition({queue: false, width: scoreFieldsOut}, 500, "ease");
  $("#logo").transition({queue: false, top: logoUp}, 500, "ease");
  $(".score").transition({queue: false, width: scoreOut}, 500, "ease", function () {
    $(".score-number").transition({queue: false, opacity: 1}, 750, "ease");
    $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
    $(".score-fields").transition({queue: false, opacity: 1}, 750, "ease");
    $(".score-aux").transition({queue: false, opacity: 1}, 750, "ease");
    hubActiveController.restartPendingHubActiveIndicators();
  });
};

const transitionIntroToTimeout = function (callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease", function () {
    $("#eventMatchInfo").hide();
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function () {
      $(".avatars").css("opacity", 0);
      $(".avatars").hide();
      $(".teams").hide();
      $("#timeoutDetails").transition({queue: false, width: timeoutDetailsOut}, 500, "ease");
      $("#logo").transition({queue: false, top: logoUp}, 500, "ease", function () {
        $(".timeout-detail").transition({queue: false, opacity: 1}, 750, "ease");
        $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
      });
    });
  });
};

const transitionLogoToBlank = function (callback) {
  showMessage(callback);
}

const transitionMatchToBlank = function (callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "ease");
  $(".score-aux").transition({queue: false, opacity: 0}, 750, "ease");
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear", function () {
    $("#eventMatchInfo").hide();
    $(".score-fields").transition({queue: false, width: 0}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease");
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function () {
      $(".teams").hide();
      $(".score-fields").hide();
      showMessage(callback);
    });
  });
};

const transitionMatchToIntro = function (callback) {
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "ease");
  $(".score-aux").transition({queue: false, opacity: 0}, 750, "ease");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function () {
    $(".score-fields").transition({queue: false, width: 0}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease");
    $(".score").transition({queue: false, width: scoreMid}, 500, "ease", function () {
      $(".score-fields").hide();
      $(".avatars").css("display", "flex");
      $(".avatars").transition({queue: false, opacity: 1}, 500, "ease", callback);
    });
  });
};

const transitionTimeoutToBlank = function (callback) {
  $(".timeout-detail").transition({queue: false, opacity: 0}, 300, "linear");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function () {
    $("#timeoutDetails").transition({queue: false, width: timeoutDetailsIn}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease", function () {
      showMessage(callback);
    });
  });
};

const transitionTimeoutToIntro = function (callback) {
  $(".timeout-detail").transition({queue: false, opacity: 0}, 300, "linear");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function () {
    $("#timeoutDetails").transition({queue: false, width: timeoutDetailsIn}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease", function () {
      $(".avatars").css("display", "flex");
      $(".avatars").css("opacity", 1);
      $(".teams").css("display", "flex");
      $(".score").transition({queue: false, width: scoreMid}, 500, "ease", function () {
        $("#eventMatchInfo").show();
        $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoDown}, 500, "ease", callback);
      });
    });
  });
};

const showMessage = function (callback) {
  if (!hasMessage) {
    if (callback) {
      callback();
    }
    return;
  }
  $("#message").show();
  $("#message").transition({queue: false, opacity: 1}, 750, "ease", callback);
};

const hideMessage = function (callback) {
  if (!hasMessage) {
    if (callback) {
      callback();
    }
    return;
  }
  $("#message").transition({queue: false, opacity: 0}, 750, "ease", function () {
    $("#message").hide();
    if (callback) {
      callback();
    }
  });
};

const applyWallBootstrap = function (data) {
  handleMatchLoad(legacyMatch(data.match)); handleRealtimeScore(legacyRealtimeScore(data.realtimeScore));
  handleV1Timing(data.timing); handleV1MatchClock(data.matchClock); handleAudienceDisplayMode(data.displayMode);
};
const loadWallBootstrap = function () {
  const requestId = ++bootstrapRequestId; const controller = new AbortController(); const timeout = setTimeout(() => controller.abort(), 5000);
  return fetch("/api/v1/displays/wall/bootstrap", {signal: controller.signal})
    .then(response => { if (!response.ok) { throw new Error("Unable to load wall bootstrap: " + response.status); } return response.json(); })
    .then(response => { if (requestId === bootstrapRequestId) { applyWallBootstrap(response.data); } })
    .catch(error => console.error(error)).finally(() => clearTimeout(timeout));
};

$(function () {
  // Read the configuration for this display from the URL query string.
  const urlParams = new URLSearchParams(window.location.search);
  document.body.style.backgroundColor = urlParams.get("background");
  const sides = DisplayShared.applyDisplaySides(urlParams);
  redSide = sides.redSide;
  blueSide = sides.blueSide;

  // Adjust position and size of display contents.
  const overlayCentering = $("#overlayCentering");
  overlayCentering.css("top", parseInt(urlParams.get("topSpacingPx")) + overlayTopOffset + "px");
  overlayCentering.css("transform", `scale(${urlParams.get("zoomFactor")})`);

  messageText = urlParams.get("message") || "";
  hasMessage = messageText !== "";
  const messageDiv = $("#message");
  messageDiv.text(messageText);
  messageDiv.toggle(hasMessage);
  if (hasMessage) {
    showMessage();
  }

  loadWallBootstrap().finally(function () {
    websocket = new CheesyWebsocketV1("/api/v1/streams/displays/wall", {
    audienceDisplayMode: function (event) {
      handleAudienceDisplayMode(event.data);
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
      handleRealtimeScore(legacyRealtimeScore(event.data));
    },
    }, loadWallBootstrap);
  });

  // Map how to transition from one screen to another. Missing links between screens indicate that first we
  // must transition to the blank screen and then to the target screen.
  transitionMap = {
    blank: {
      intro: transitionBlankToIntro,
      logo: transitionBlankToLogo,
      match: transitionBlankToMatch,
      timeout: transitionBlankToTimeout,
    },
    intro: {
      blank: transitionIntroToBlank,
      match: transitionIntroToMatch,
      timeout: transitionIntroToTimeout,
    },
    logo: {
      blank: transitionLogoToBlank,
    },
    match: {
      blank: transitionMatchToBlank,
      intro: transitionMatchToIntro,
    },
    timeout: {
      blank: transitionTimeoutToBlank,
      intro: transitionTimeoutToIntro,
    },
  }
});
