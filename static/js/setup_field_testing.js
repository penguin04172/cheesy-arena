// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the Field Testing page.

var websocket;
var stateWebsocket;
var bootstrapRequestId = 0;
var plcOverrideAllowed = false;
var allowedOverrideMatchStates = [0, 5, 6, 7];
var disabledOverrideTooltipText = "Cannot override coil while match is in progress.";
var disabledLedTooltipText = "Cannot override LED Lighting while match is in progress.";
var legacyMatchStateIds = {pre_match: 0, start_match: 1, auto: 2, pause: 3, teleop: 4, post_match: 5, timeout_active: 6, post_timeout: 7};
var legacyPlcIo = data => ({Inputs: data.Inputs || data.inputs, Registers: data.Registers || data.registers, Coils: data.Coils || data.coils, CoilOverrides: data.CoilOverrides || data.coilOverrides});
var legacyLedStatus = data => ({Red: data.red.map(color => ({R: color.r, G: color.g, B: color.b})), Blue: data.blue.map(color => ({R: color.r, G: color.g, B: color.b})), RedMode: data.redMode, BlueMode: data.blueMode});

// Sends a websocket message to play a given game sound on the audience display.
var playSound = function (sound) {
  websocket.send("playSound", sound);
};

// Sends a websocket message to set the selected LED test mode.
var setLedMode = function () {
  websocket.send("setLedMode", {
    RedMode: parseInt(modeSelects["red"].val(), 10),
    BlueMode: parseInt(modeSelects["blue"].val(), 10)
  });
};

var setPlcCoilOverride = function (index, override) {
  websocket.send("setPlcCoilOverride", { Index: index, Override: override });
};

var getNextOverrideState = function (overrideState) {
  switch (overrideState) {
    case "on":
      return "off";
    case "off":
      return "auto";
    default:
      return "on";
  }
};

var updateCoilOverrideTooltips = function () {
  $(".plc-coil-indicator").each(function (_index, element) {
    element.setAttribute("data-plc-override-allowed", plcOverrideAllowed);

    const tooltip = bootstrap.Tooltip.getInstance(element);
    if (plcOverrideAllowed) {
      if (tooltip) {
        tooltip.dispose();
      }
    } else if (!tooltip) {
      new bootstrap.Tooltip(element, { title: disabledOverrideTooltipText });
    }
  });
};

var updateLedOverrideTooltips = function () {
  $(".led-mode-wrapper").each(function (_index, element) {
    const tooltip = bootstrap.Tooltip.getInstance(element);
    if (plcOverrideAllowed) {
      if (tooltip) {
        tooltip.dispose();
      }
    } else if (!tooltip) {
      new bootstrap.Tooltip(element, { title: disabledLedTooltipText });
    }
  });
};

// Handles a websocket message to update the PLC IO status.
var handlePlcIoChange = function (data) {
  $.each(data.Inputs, function (index, input) {
    $("#input" + index).text(input)
    $("#input" + index).attr("data-plc-value", input);
  });

  $.each(data.Registers, function (index, register) {
    $("#register" + index).text(register)
  });

  $.each(data.Coils, function (index, coil) {
    const coilIndicator = $("#coil" + index);
    coilIndicator.text(coil)
    coilIndicator.attr("data-plc-value", coil);
    coilIndicator.attr("data-plc-override", data.CoilOverrides[index]);
  });
};

var handleArenaStatus = function (data) {
  plcOverrideAllowed = allowedOverrideMatchStates.includes(data.MatchState);
  updateCoilOverrideTooltips();
  updateLedOverrideTooltips();
  
  modeSelects["red"].prop("disabled", !plcOverrideAllowed);
  modeSelects["blue"].prop("disabled", !plcOverrideAllowed);
};

var ledContainers = {};
var modeSelects = {};
var lastServerModes = {};

var handleLedStatus = function (data) {
  var renderPixels = function(containerId, pixels) {
    if (!ledContainers[containerId]) {
      var container = $("#" + containerId);
      container.css({"display": "flex", "flex-direction": "column", "gap": "4px"});
      var boxes = [];
      for (var i = 0; i < 4; i++) {
        var row = $("<div></div>").css({"display": "flex", "flex-direction": "row", "gap": "2px"});
        for (var j = 0; j < 8; j++) {
          var box = $("<div></div>").css({
            "width": "12px",
            "height": "12px",
            "background-color": "black",
            "border": "1px solid #333"
          });
          boxes.push(box);
          row.append(box);
        }
        container.append(row);
      }
      ledContainers[containerId] = boxes;
    }
    
    var boxes = ledContainers[containerId];
    for (var i = 0; i < 4; i++) {
      for (var j = 0; j < 8; j++) {
        var pixelIndex = i * 16 + j;
        var color = pixels[pixelIndex];
        boxes[i * 8 + j].css("background-color", "rgb(" + color.R + "," + color.G + "," + color.B + ")");
      }
    }
  };
  
  renderPixels("redHubPixels", data.Red);
  renderPixels("blueHubPixels", data.Blue);
  
  var syncModeSelect = function(alliance, mode) {
    if (lastServerModes[alliance] !== mode || modeSelects[alliance].val() != mode) {
      modeSelects[alliance].val(mode);
      lastServerModes[alliance] = mode;
    }
  };
  
  syncModeSelect("red", data.RedMode);
  syncModeSelect("blue", data.BlueMode);
};

var applyFieldTestingBootstrap = function (data) {
  handlePlcIoChange(legacyPlcIo(data.plcIo)); handleArenaStatus({MatchState: legacyMatchStateIds[data.matchState]}); handleLedStatus(legacyLedStatus(data.ledStatus));
};
var loadFieldTestingBootstrap = function () {
  var requestId = ++bootstrapRequestId; var controller = new AbortController(); var timeout = setTimeout(() => controller.abort(), 5000);
  return fetch("/api/v1/admin/field-testing/bootstrap", {signal: controller.signal})
    .then(response => { if (!response.ok) { throw new Error("Unable to load field testing bootstrap: " + response.status); } return response.json(); })
    .then(response => { if (requestId === bootstrapRequestId) { applyFieldTestingBootstrap(response.data); } })
    .catch(error => console.error(error)).finally(() => clearTimeout(timeout));
};

$(function () {
  modeSelects["red"] = $("select[name=redLedMode]");
  modeSelects["blue"] = $("select[name=blueLedMode]");
  
  updateCoilOverrideTooltips();
  updateLedOverrideTooltips();

  $(document).on("click", ".plc-coil-indicator", function (event) {
    if (!plcOverrideAllowed) {
      return;
    }

    const currentOverrideState = $(event.currentTarget).attr("data-plc-override");
    const nextOverrideState = getNextOverrideState(currentOverrideState);
    setPlcCoilOverride(parseInt($(event.currentTarget).attr("data-coil-index"), 10), nextOverrideState);
  });

  websocket = new CheesyWebsocket("/setup/field_testing/websocket", {});
  loadFieldTestingBootstrap().finally(function () {
    stateWebsocket = new CheesyWebsocketV1("/api/v1/streams/admin/field-testing", {
    plcIo: function (event) {
      handlePlcIoChange(legacyPlcIo(event.data));
    },
    matchState: function (event) {
      handleArenaStatus({MatchState: legacyMatchStateIds[event.data]});
    },
    ledStatus: function (event) {
      handleLedStatus(legacyLedStatus(event.data));
    }
    }, loadFieldTestingBootstrap);
  });
});
