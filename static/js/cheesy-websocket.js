// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Shared code for initiating websocket connections back to the server for full-duplex communication.

var CheesyWebsocket = function (path, events) {
  var that = this;
  var protocol = "ws://";
  if (window.location.protocol === "https:") {
    protocol = "wss://";
  }
  var url = protocol + window.location.hostname;
  if (window.location.port !== "") {
    url += ":" + window.location.port;
  }
  url += path;

  // Append the page's query string to the websocket URL.
  url += window.location.search;

  // Insert a default error-handling event if a custom one doesn't already exist.
  if (!events.hasOwnProperty("error")) {
    events.error = function (event) {
      // Data is just an error string.
      console.log(event.data);
      alert(event.data);
    };
  }

  // Parse the display parameters that will be present in the query string if this is a display.
  var displayId = new URLSearchParams(window.location.search).get("displayId");

  // Insert an event to allow the server to force-reload the client for any display.
  events.reload = function (event) {
    if (event.data === null || event.data === displayId) {
      location.reload();
    }
  };

  // Insert an event to allow reconfiguration if this is a display.
  if (!events.hasOwnProperty("displayConfiguration")) {
    events.displayConfiguration = function (event) {
      var newUrl = event.data;

      // Reload the display if the configuration has changed.
      if (newUrl !== window.location.pathname + window.location.search) {
        window.location = newUrl;
      }
    };
  }

  this.connect = function () {
    this.websocket = $.websocket(url, {
      open: function () {
        console.log("Websocket connected to the server at " + url + ".")
      },
      close: function () {
        console.log("Websocket lost connection to the server. Reconnecting in 3 seconds...");
        setTimeout(that.connect, 3000);
      },
      events: events
    });
  };

  this.send = function (type, data) {
    this.websocket.send(type, data);
  };

  this.connect();
};

// V1 stream client with per-event sequence tracking. Call onGap when a reconnect or dropped notification requires a
// fresh REST bootstrap; existing pages continue using CheesyWebsocket until their v1 stream migration is complete.
var CheesyWebsocketV1 = function (path, events, onGap) {
  var that = this;
  this.sequences = {};
  this.ready = false;
  this.hasConnected = false;

  this.connect = function () {
    var protocol = window.location.protocol === "https:" ? "wss://" : "ws://";
    var socket = new WebSocket(protocol + window.location.host + path + window.location.search);
    that.websocket = socket;
    socket.onopen = function () {
      console.log("Websocket v1 connected to " + path + ".");
      if (that.hasConnected && onGap) { onGap("reconnect", null, null); }
      that.hasConnected = true;
    };
    socket.onclose = function () {
      that.ready = false;
      console.log("Websocket v1 lost connection. Reconnecting in 3 seconds...");
      setTimeout(that.connect, 3000);
    };
    socket.onmessage = function (event) {
      var message = JSON.parse(event.data);
      if (!message.meta || message.meta.version !== 1) { return; }
      if (message.type === "ready") {
        that.sequences = message.data.sequences || {};
        that.ready = true;
        if (events.ready) { events.ready(message); }
        return;
      }
      if (message.type === "ping") { return; }
      var previous = that.sequences[message.type];
      if (!message.meta.bootstrap && previous !== undefined && message.meta.sequence !== previous + 1) {
        that.ready = false;
        if (onGap) { onGap(message.type, previous, message.meta.sequence); }
      }
      that.sequences[message.type] = message.meta.sequence;
      if (events[message.type]) { events[message.type](message); }
    };
  };

  this.send = function (type, data) {
    this.websocket.send(JSON.stringify({type: type, data: data}));
  };

  this.connect();
};
