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
		const d = tmp.content.cloneNode(true) as DocumentFragment
		const post = d.querySelector(".post") as HTMLDivElement
		const comments = d.querySelector(".comments") as HTMLDivElement
		const msginput = d.querySelector(".msginput") as HTMLInputElement

		const path = event.data as string
		var spl = path.split(".")
		var ext = ""
		if (spl.length > 1){
			ext = spl.pop()!
		}
		switch (ext) {
			case "mp4":
				const vt = (document.getElementById("templatevideo") as HTMLTemplateElement).content.cloneNode(true) as DocumentFragment
				const v = vt.querySelector("video")!
				v.setAttribute("src", path)

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
				post.append(vt)
				break;
			case "mp3":
				const at = (document.getElementById("templateaudio") as HTMLTemplateElement).content.cloneNode(true) as DocumentFragment
				const ha = at.querySelector("a")!
				ha.setAttribute("href", path)
				ha.innerText = path
				at.querySelector("source")!.setAttribute("src", path)
				
				const a = at.querySelector("audio")!
				a.onplaying = (e)=>{
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
				a.onvolumechange = (e)=>{
					const el = e.target as HTMLAudioElement
					volumeMusic = el.volume;
					muteMusic = el.muted;
				}
				post.append(at)
				break;
			case "jpg":
			case "png":
			case "jpeg":
				const i = (document.getElementById("templateimg") as HTMLTemplateElement).content.cloneNode(true) as DocumentFragment
				i.querySelector("img")!.setAttribute("src", path)
				post.append(i)
				break;
			default:
				const o = (document.getElementById("templateother") as HTMLTemplateElement).content.cloneNode(true) as DocumentFragment
				const ho = o.querySelector("a")!
				ho.setAttribute("href", path) 
				ho.innerText = path
				post.append(o)
				break;
		}
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