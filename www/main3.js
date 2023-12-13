"use strict";
window.onload = function () {
    var ip = window.location.host;
    var currentVideo = null;
    var volumeVideo = 1;
    var muteVideo = false;
    var currentMusic = null;
    var volumeMusic = 1;
    var muteMusic = false;
    var socket = new WebSocket("ws://" + ip + "/ws");
    var socketUpload = new WebSocket("ws://" + ip + "/upload");
    function appendToBody(event) {
        var d = document.createElement('div');
        d.className = 'card';
        d.innerHTML = event.data;
        var content = document.getElementById('content');
        content.prepend(d);
        d.addEventListener("dragstart", function (e) {
            e.preventDefault();
        });
        var m = d.getElementsByClassName("music")[0];
        if (m != undefined) {
            m.onplaying = function (e) {
                var el = e.target;
                if (el == currentMusic)
                    return;
                el.volume = volumeMusic;
                el.muted = muteMusic;
                if (currentMusic) {
                    currentMusic.classList.remove("musicfix");
                    currentMusic.pause();
                }
                el.classList.add("musicfix");
                currentMusic = el;
            };
            m.onvolumechange = function (e) {
                var el = e.target;
                volumeMusic = el.volume;
                muteMusic = el.muted;
            };
        }
        var v = d.getElementsByClassName("video")[0];
        if (v != undefined) {
            v.onplaying = function (e) {
                var el = e.target;
                if (e.target == currentVideo)
                    return;
                el.volume = volumeVideo;
                el.muted = muteVideo;
                if (currentVideo) {
                    currentVideo.pause();
                }
                currentVideo = el;
            };
            v.onvolumechange = function (e) {
                var el = e.target;
                volumeVideo = el.volume;
                muteVideo = el.muted;
            };
        }
    }
    socketUpload.onmessage = appendToBody;
    socket.onmessage = appendToBody;
    socket.onclose = function () {
        console.log('Service', "WebSocket Disconnected");
    };
    socket.onerror = function () {
        console.log('Service', "WebSocket Error");
    };
    socket.onopen = function () {
        console.log('Service', "WebSocket Connected");
        // socket.send('Ураааааа!')
    };
    var btn = document.getElementById('btn');
    var inp = document.getElementById('inp');
    btn.onclick = function () {
        socket.send(inp.value);
    };
    inp.onkeydown = function (event) {
        if (event.key == 'Enter') {
            socket.send(inp.value);
            inp.value = "";
        }
    };
    document.addEventListener("dragover", function (e) {
        e.preventDefault();
    });
    document.addEventListener("drop", function (e) {
        e.preventDefault();
        var fs = e.dataTransfer.files;
        for (var index = 0; index < fs.length; index++) {
            var element = fs[index];
            console.log('You selected ' + element.name);
            socketUpload.send(element.name);
            socketUpload.send(element.size.toString());
            console.log('File size: ' + element.size);
            socketUpload.send(element);
        }
    });
};
