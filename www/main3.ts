"use strict";

window.onload = function () {
	const ip = window.location.host
	var currentVideo:HTMLVideoElement|null = null
	var volumeVideo = 1
	var muteVideo = false
	var currentMusic:HTMLAudioElement|null = null
	var volumeMusic = 1
	var muteMusic = false
	var lastMsg = ""
	var ring = new Audio("./upload/Iphone - Message Tone.mp3")

	const tmp = document.getElementById("templatecard") as HTMLTemplateElement
	const content = document.getElementById('content') as HTMLDivElement

	const socket = new WebSocket("ws://" + ip + "/ws");
	const socketUpload = new WebSocket("ws://" + ip + "/uploadws");

	function appendToBody(event: MessageEvent) {
		const d = tmp.content.cloneNode(true) as DocumentFragment
		const post = d.querySelector(".post") as HTMLDivElement
		const comments = d.querySelector(".comments") as HTMLDivElement
		const msginput = d.querySelector(".msginput") as HTMLInputElement
		const wrapper = d.querySelector(".wrapper") as HTMLDivElement

		const path = event.data as string
		wrapper.id = path
		var spl = path.split(".")
		var ext = ""
		// If spl.length is one, it's a visible file with no extension ie. file
		// If spl[0] === "" and spl.length === 2 it's a hidden file with no extension ie. .htaccess
		if( !(spl.length === 1 || ( spl[0] === "" && spl.length === 2 )) ){
			ext = spl.pop()!.toLowerCase()
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
				const buttons = vt.querySelectorAll(".butspeed")
				buttons.forEach(element => {
					(element as HTMLButtonElement).onclick = (e) => {
						const el = e.target as HTMLButtonElement
						switch (el.innerText) {
							case "1.0x":
								v.playbackRate = 1;
							break;
							case "1.25x":
								v.playbackRate = 1.25;
							break;
							case "1.5x":
								v.playbackRate = 1.5;
							break;
							case "2.0x":
								v.playbackRate = 2;
							break;
						}
					}
				});
				const videoname = vt.querySelector(".videoname") as HTMLParagraphElement
				videoname.innerText = path.split("/").pop()!;
				post.append(vt)
				break;
			case "mp3":
				const at = (document.getElementById("templateaudio") as HTMLTemplateElement).content.cloneNode(true) as DocumentFragment
				const ha = at.querySelector("a")!
				ha.setAttribute("href", path)
				ha.innerText = path.split("/").pop()!
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
				ho.innerText = path.split("/").pop()!
				post.append(o)
				break;
		}

		var xhr = new XMLHttpRequest()
		xhr.open('GET', path + "._msg", true)
		xhr.setRequestHeader("Cache-Control", "no-cache")
		xhr.send()
		xhr.onload = () => {
			if(xhr.status==200){
				comments.innerText = xhr.response + "<history restored>\n\n"
				comments.scrollTo(0, comments.scrollHeight)
			}
		}

		msginput.onkeydown = (event) => {
			if(event.key == 'Enter') {
				const el = event.target as HTMLInputElement
				const p = el.parentElement as HTMLDivElement
				const com = p.querySelector(".comments") as HTMLDivElement
				if (el.value.length == 0) {
					return
				}
				lastMsg = el.value
				socket.send(JSON.stringify({
					id: p.id,
					txt: el.value
				}))
				el.value = "";
			}
		}
		content.prepend(d)//делается в конце функции т.к. после выполнения данной команды содержимое d недоступно
	}

	socketUpload.onmessage = appendToBody;
	socket.addEventListener("message", (e)=>{
		const d = JSON.parse(e.data)
		const com = document.getElementById(d.id)!.querySelector(".comments") as HTMLDivElement
		com.innerText += d.txt + "\n"
		com.scrollTo(0, com.scrollHeight)
		//ring.play()
	})

	socket.onclose = () => {
		console.log('Service', "WebSocket Disconnected");
	}
	socket.onerror = () => {
		console.log('Service', "WebSocket Error");
	}
	socket.onopen = () => {
		console.log('Service', "WebSocket Connected");
		// socket.send('Ураааааа!')
	}

	const over = document.querySelector(".over") as HTMLDivElement
	document.body.addEventListener("dragenter", (e) => {
		e.preventDefault();
		over.classList.add("show")
	})
	over.addEventListener("dragleave", (e) => {
		e.preventDefault();
		over.classList.remove("show")
	})

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
			console.log('File size: ' + el.size);
			var formData = new FormData();
			formData.append("file", el);
			var xhr = new XMLHttpRequest()
			xhr.upload.onprogress = (e) => {
				const ev = e as ProgressEvent
				console.log("Upload progress: " + (ev.loaded/ev.total*100).toFixed(0) + "%")
				over.innerText = "Upload progress: " + (ev.loaded/ev.total*100).toFixed(0) + "%"
			}
			xhr.upload.onloadend = (e) => {
				over.classList.remove("show")
				over.innerText = "Drop it!"
			}
			xhr.open('POST', "uploadxml/" + el.name, true)
			xhr.send(formData);
		}
	})
	function snappingOn() {
		document.documentElement.classList.add("snappingOn")
		document.removeEventListener("scroll", snappingOn)
	}
	document.addEventListener("scroll", snappingOn)//при начальной загрузке карточек из-за снаппинга лента сама скролится вниз,
													//поэтому включаю снаппинг когда скролит юзер
}