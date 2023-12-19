"use strict";

window.onload = function () {
	const ip = window.location.host
	var currentVideo:HTMLVideoElement|null = null
	var volumeVideo = 1
	var muteVideo = false
	var currentMusic:HTMLAudioElement|null = null
	var volumeMusic = 1
	var muteMusic = false

	const tmp = document.getElementById("templatecard") as HTMLTemplateElement
	const content = document.getElementById('content') as HTMLDivElement

	const socket = new WebSocket("ws://" + ip + "/ws");
	const socketUpload = new WebSocket("ws://" + ip + "/upload");

	function appendToBody(event: MessageEvent) {
		const d = tmp.content.cloneNode(true) as HTMLDivElement
		const post = d.querySelector(".post") as HTMLDivElement
		const comments = d.querySelector(".comments") as HTMLDivElement
		const msginput = d.querySelector(".msginput") as HTMLInputElement

		post.innerHTML = event.data
		comments.innerHTML = ""
		msginput.onkeydown = function (event) {
			if(event.key == 'Enter') {
				// socket.send(inp.value);
				if (msginput.value.length == 0) {
					return
				}
				comments.innerHTML = msginput.value+"<br>"+comments.innerHTML;
				msginput.value = "";
			}
		}

		var m = d.querySelector(".music") as HTMLAudioElement
		if(m){
			m.onplaying = (e)=>{
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

		var v = d.querySelector(".video") as HTMLVideoElement
		if(v){
			v.onplaying = (e)=>{
				const el = e.target as HTMLVideoElement
				if(el == currentVideo)
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
		content.prepend(d)//делается в конце функции т.к. после выполнения данной команды содержимое d недоступно
	}

	socketUpload.onmessage = appendToBody;
	socket.addEventListener("message", appendToBody)
	socket.addEventListener("message", (e)=>{
		let src = "/upload/Iphone - Message Tone.mp3"
		let ring = new Audio(src)
		ring.play()
	})

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
	document.addEventListener("dragstart", (e) => {
		e.preventDefault();
	})
	document.addEventListener("drop", (e) => {
		e.preventDefault();
		var fs = e.dataTransfer!.files;
		for (let index = 0; index < fs.length; index++) {
			const el = fs[index];
			console.log('You selected ' + el.name);
			socketUpload.send(el.name);
			socketUpload.send(el.size.toString());
			console.log('File size: ' + el.size);
			socketUpload.send(el);
		}
	})
}