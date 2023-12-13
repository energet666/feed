"use strict";

window.onload = function () {
	var ip = window.location.host
	var currentVideo:HTMLVideoElement|null = null
	var volumeVideo = 1
	var muteVideo = false
	var currentMusic:HTMLAudioElement|null = null
	var volumeMusic = 1
	var muteMusic = false

	var socket = new WebSocket("ws://" + ip + "/ws");
	var socketUpload = new WebSocket("ws://" + ip + "/upload");

	function appendToBody(event: MessageEvent) {
		var d = document.createElement('div');
		d.className = 'card';
		d.innerHTML = event.data;
		var content = document.getElementById('content')!
		content.prepend(d)

		d.addEventListener("dragstart", (e)=>{
			e.preventDefault();
		})

		var m = d.getElementsByClassName("music")[0] as HTMLAudioElement
		if(m != undefined){
			m.onplaying = (e: Event)=>{
				const el = e.target as HTMLAudioElement
				if(el == currentMusic)
					return
				el.volume = volumeMusic;
				el.muted = muteMusic;
				if(currentMusic) {
					currentMusic.classList.remove("musicfix");
					currentMusic.pause();
				}
				el.classList.add("musicfix");
				currentMusic = el;
			}
			m.onvolumechange = (e)=>{
				const el = e.target as HTMLAudioElement
				volumeMusic = el.volume;
				muteMusic = el.muted;
			}
		}

		var v = d.getElementsByClassName("video")[0] as HTMLVideoElement
		if(v != undefined){
			v.onplaying = (e)=>{
				const el = e.target as HTMLVideoElement
				if(e.target == currentVideo)
					return
				el.volume = volumeVideo;
				el.muted = muteVideo;
				if (currentVideo) {
					currentVideo.pause();
				}
				currentVideo = el;
			}
			v.onvolumechange = (e) => {
				const el = e.target as HTMLVideoElement
				volumeVideo = el.volume;
				muteVideo = el.muted;
			}
		}
	}

	socketUpload.onmessage = appendToBody;
	socket.onmessage = appendToBody;

	socket.onclose = function () {
		console.log('Service', "WebSocket Disconnected");
	}
	socket.onerror = function () {
		console.log('Service', "WebSocket Error");
	}
	socket.onopen = function () {
		console.log('Service', "WebSocket Connected");
		// socket.send('Ураааааа!')
	}

	var btn = document.getElementById('btn') as HTMLButtonElement
	var inp = document.getElementById('inp') as HTMLInputElement
	btn.onclick = function () {
		socket.send(inp.value);
	}
	inp.onkeydown = function (event) {
		if(event.key == 'Enter') {
			socket.send(inp.value);
			inp.value = "";
		}
	}

	document.addEventListener("dragover", (e) => {
		e.preventDefault();
	})
	document.addEventListener("drop", (e) => {
		e.preventDefault();
		var fs = e.dataTransfer!.files;
		for (let index = 0; index < fs.length; index++) {
			const element = fs[index];
			console.log('You selected ' + element.name);
			socketUpload.send(element.name);
			socketUpload.send(element.size.toString());
			console.log('File size: ' + element.size);
			socketUpload.send(element);
		}
	})
}