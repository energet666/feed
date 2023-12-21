"use strict";
window.onload = function () {
    var ip = window.location.host;
    var currentVideo = null;
    var volumeVideo = 1;
    var muteVideo = false;
    var currentMusic = null;
    var volumeMusic = 1;
    var muteMusic = false;
    var tmp = document.getElementById("templatecard");
    var content = document.getElementById('content');
    var socket = new WebSocket("ws://" + ip + "/ws");
    var socketUpload = new WebSocket("ws://" + ip + "/upload");
    function appendToBody(event) {
        var d = tmp.content.cloneNode(true);
        var post = d.querySelector(".post");
        var comments = d.querySelector(".comments");
        var msginput = d.querySelector(".msginput");
        var path = event.data;
        var spl = path.split(".");
        var ext = "";
        if (spl.length > 1) {
            ext = spl.pop();
        }
        switch (ext) {
            case "mp4":
                var vt = document.getElementById("templatevideo").content.cloneNode(true);
                var v = vt.querySelector("video");
                v.setAttribute("src", path);
                v.onplaying = function (e) {
                    var el = e.target;
                    if (el == currentVideo)
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
                post.append(vt);
                break;
            case "mp3":
                var at = document.getElementById("templateaudio").content.cloneNode(true);
                var ha = at.querySelector("a");
                ha.setAttribute("href", path);
                ha.innerText = path;
                at.querySelector("source").setAttribute("src", path);
                var a = at.querySelector("audio");
                a.onplaying = function (e) {
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
                a.onvolumechange = function (e) {
                    var el = e.target;
                    volumeMusic = el.volume;
                    muteMusic = el.muted;
                };
                post.append(at);
                break;
            case "jpg":
            case "png":
            case "jpeg":
                var i = document.getElementById("templateimg").content.cloneNode(true);
                i.querySelector("img").setAttribute("src", path);
                post.append(i);
                break;
            default:
                var o = document.getElementById("templateother").content.cloneNode(true);
                var ho = o.querySelector("a");
                ho.setAttribute("href", path);
                ho.innerText = path;
                post.append(o);
                break;
        }
        comments.innerHTML = "";
        msginput.onkeydown = function (event) {
            if (event.key == 'Enter') {
                // socket.send(inp.value);
                if (msginput.value.length == 0) {
                    return;
                }
                comments.innerHTML = msginput.value + "<br>" + comments.innerHTML;
                msginput.value = "";
            }
        };
        content.prepend(d); //делается в конце функции т.к. после выполнения данной команды содержимое d недоступно
    }
    socketUpload.onmessage = appendToBody;
    socket.addEventListener("message", appendToBody);
    socket.addEventListener("message", function (e) {
        var src = "/upload/Iphone - Message Tone.mp3";
        var ring = new Audio(src);
        ring.play();
    });
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
    document.addEventListener("dragstart", function (e) {
        e.preventDefault();
    });
    document.addEventListener("drop", function (e) {
        e.preventDefault();
        var fs = e.dataTransfer.files;
        for (var index = 0; index < fs.length; index++) {
            var el = fs[index];
            console.log('You selected ' + el.name);
            socketUpload.send(el.name);
            socketUpload.send(el.size.toString());
            console.log('File size: ' + el.size);
            socketUpload.send(el);
        }
    });
};
