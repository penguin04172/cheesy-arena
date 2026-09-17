// Copyright 2026 Team 254. All Rights Reserved.

// Client-side methods for the unpicked teams display.

$(function () {
  // Read the configuration for this display from the URL query string.
  const urlParams = new URLSearchParams(window.location.search);

  if (urlParams.get("inverted") === "true") {
    $("#unpickedTeamsContainer").css("transform", "rotate(180deg)");
    $("#gameLogoContainer").css("transform", "rotate(180deg)");
  }

  const $unpickedTeams = $("#unpickedTeams");
  const $container = $("#unpickedTeamsContainer");
  const $logoContainer = $("#gameLogoContainer");
  const unpickedTeamsTemplate = Handlebars.compile($("#unpickedTeamsTemplate").html());

  let currentRankedTeams = [];
  let isAllianceSelection = false;
  let bootstrapRequestId = 0;

  const updateDisplay = function() {
    if (isAllianceSelection && currentRankedTeams.length > 0) {
      $container.css("display", "flex");
      $logoContainer.hide();
      
      const unpickedTeams = currentRankedTeams.filter(team => !team.Picked);
      $unpickedTeams.html(unpickedTeamsTemplate(unpickedTeams));
    } else {
      $container.hide();
      $logoContainer.css("display", "flex");
    }
  };

  const applyAllianceSelection = function (data) {
    currentRankedTeams = (data.rankedTeams || []).map(team => ({Rank: team.rank, TeamId: team.teamId, Picked: team.picked}));
    updateDisplay();
  };
  const applyBootstrap = function (data) {
    applyAllianceSelection(data.allianceSelection);
    isAllianceSelection = data.displayMode === "allianceSelection";
    updateDisplay();
  };
  const loadBootstrap = function () {
    const requestId = ++bootstrapRequestId;
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 5000);
    return fetch("/api/v1/displays/unpicked/bootstrap", {signal: controller.signal})
      .then(response => { if (!response.ok) { throw new Error("Unable to load unpicked bootstrap: " + response.status); } return response.json(); })
      .then(response => { if (requestId === bootstrapRequestId) { applyBootstrap(response.data); } })
      .catch(error => console.error(error)).finally(() => clearTimeout(timeout));
  };

  loadBootstrap().finally(function () {
    new CheesyWebsocketV1("/api/v1/streams/displays/unpicked", {
    allianceSelection: function (event) {
      applyAllianceSelection(event.data);
    },
    audienceDisplayMode: function (event) {
      isAllianceSelection = event.data === "allianceSelection";
      updateDisplay();
    }
    }, loadBootstrap);
  });
});


