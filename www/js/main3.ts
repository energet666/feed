import tools from './functions.js'

tools.test("hi from module")

const ip = window.location.host

let currentVideo:HTMLVideoElement|null = null
let volumeVideo = 0.25
let muteVideo = false

let currentMusic:HTMLAudioElement|null = null
let volumeMusic = 0.25
let muteMusic = false

let lastMsg = ""
//let ring = new Audio("./upload/Iphone - Message Tone.mp3")
let topPost:number|null = null
let bottomPost:number|null = null

let cardTemplate:HTMLTemplateElement
let contentBlock:HTMLDivElement
let uploadingOverlay:HTMLDivElement

let socket:WebSocket
let socketUpload:WebSocket
let socketEvent:WebSocket

type msgStruct = {
	id: string,
	txt: string,
}

window.onload = () => {
	socket       = new WebSocket("ws://" + ip + "/ws")
	socketUpload = new WebSocket("ws://" + ip + "/uploadws")
	socketEvent  = new WebSocket("ws://" + ip + "/eventws")

	cardTemplate = document.getElementById("templatecard") as HTMLTemplateElement
	contentBlock = document.getElementById('content') as HTMLDivElement
	uploadingOverlay = document.querySelector(".over") as HTMLDivElement

	socket.onclose   = () => {console.log(`WS "ws" Disconnected`)}
	socket.onerror   = () => {console.log(`WS "ws" Error`)}
	socket.onopen    = () => {console.log(`WS "ws" Connected`)}
	socket.onmessage = addComment

	socketUpload.onmessage = (e) => {
		appendToBody(e)
		endContentCheck()
	}
	//При окрытии вебсокета socketUpload запрашиваем файл с индексом "-1"(сервер ответит самым свежим файлом)
	socketUpload.onopen = () => {
		socketUpload.send("-1")
	}

	uploadingOverlay.ondragleave = (e) => {
		e.preventDefault()
		uploadingOverlay.classList.remove("show")
	}

	document.ondragenter = (e) => {
		e.preventDefault()
		uploadingOverlay.innerText = 'Drop it!'
		uploadingOverlay.classList.add("show")
	}
	document.ondragover = (e) => {e.preventDefault()}
	document.ondragstart = (e) => {e.preventDefault()}
	document.ondrop = uploadFile
	document.addEventListener("scroll", function scrollHandler() {
		endContentCheck()
		//ограничение частоты срабатывания scroll Event
		document.removeEventListener("scroll", scrollHandler)
		setTimeout( () => {
			document.addEventListener("scroll", scrollHandler)
		}, 100)
	})

	const snapOff = document.getElementById("snap_off") as HTMLInputElement
	snapOff.checked = false
	snapOff.onchange = () => {
		if (snapOff.checked) {
			document.documentElement.classList.add("snappingOn")
		} else {
			document.documentElement.classList.remove("snappingOn")
		}
	}
}

const uploadFile = (e:DragEvent) => {
	e.preventDefault()
	const files = e.dataTransfer!.files
	
	//Запрещаем пользователю загружать много файлов за раз
	if(files.length > 1) {
		uploadingOverlay.innerText = "One at a time! Try it again."
		setTimeout(()=>{
			uploadingOverlay.classList.remove("show")
		}, 3000)
		return
	}

	const file = files[0]
	console.log('You selected ' + file.name)
	console.log('File size: ' + file.size)
	let formData = new FormData()
	formData.append("file", file)
	let xhr = new XMLHttpRequest()
	//Отображаем прогресс загрузки
	xhr.upload.onprogress = (e) => {
		const ev = e as ProgressEvent
		const uploadProgress = (ev.loaded/ev.total*100).toFixed(0)
		console.log("Upload progress: " + uploadProgress + "%")
		uploadingOverlay.innerText = "Upload progress: " + uploadProgress + "%"
	}
	xhr.upload.onloadend = () => {
		uploadingOverlay.classList.remove("show")
	}
	xhr.open('POST', "/uploadxml/" + file.name, true)
	xhr.send(formData)
}

const endContentCheck = () => {
	//Если приближаемся к нижнему концу ленты на высоту окна, то запрашиваем более старые посты
	//и при необходимости удаляем вышедшие из зоны видимости на 3 высоты окна верхние посты
	if (document.documentElement.scrollHeight - window.innerHeight - document.documentElement.scrollTop < window.innerHeight) { 
		if(bottomPost! > 0) {
			socketUpload.send(String(bottomPost! - 1))
			while (document.documentElement.scrollTop > 3 * window.innerHeight) {
				let cards = contentBlock.getElementsByClassName("card")
				//если карточка оказалась последней, то выходим из цикла
				if(cards.length <= 1) {
					break
				}
				cards[0].remove()
				topPost!--
			}
		}
	} else if (document.documentElement.scrollTop < window.innerHeight) {
		socketUpload.send(String(topPost! + 1))
		while (document.documentElement.offsetHeight - window.innerHeight - document.documentElement.scrollTop > 3 * window.innerHeight) {
			let cards = contentBlock.getElementsByClassName("card")
			if(cards.length <= 1) {
				break
			}
			cards[cards.length - 1].remove()
			bottomPost!++
		}
	}
}

const addComment = (e:MessageEvent)=>{
	const data = JSON.parse(e.data) as msgStruct
	const target = document.getElementById(data.id)
	if(target){
		const commentDiv = target.querySelector(".comments") as HTMLDivElement
		commentDiv.innerText += data.txt + "\n"
		commentDiv.scrollTo(0, commentDiv.scrollHeight)
		//ring.play()
	}
}

const appendToBody = (event: MessageEvent) => {
	type fileInfo = {
		N: number,
		Path: string,
	}
	const data = JSON.parse(event.data) as fileInfo

	enum direction {
		UP = 1,
		DOWN = 0,
	}
	let dir = direction.DOWN
	if(topPost != null && bottomPost != null) {
		if(data.N === bottomPost - 1) {
			dir = direction.DOWN
			bottomPost--
		} else if(data.N === topPost + 1) {
			dir = direction.UP
			topPost++
		} else {
			return
		}
	} else {
		topPost = data.N
		bottomPost = data.N
	}

	const d = cardTemplate.content.cloneNode(true) as DocumentFragment
	const post = d.querySelector(".post") as HTMLDivElement
	const comments = d.querySelector(".comments") as HTMLDivElement
	const msginput = d.querySelector(".msginput") as HTMLInputElement
	const wrapper = d.querySelector(".wrapper") as HTMLDivElement
	const path = data.Path
	const pathName = path.split("/").pop()!

	wrapper.id = path

	//формируем пост в зависимости от типа контента
	switch (tools.getExt(path)) {
		case "mp4":
			const videoDocFragment = (document.getElementById("templatevideo") as HTMLTemplateElement).content.cloneNode(true) as DocumentFragment
			const v = videoDocFragment.querySelector("video")! as HTMLVideoElement
			v.setAttribute("src", path)
			v.volume = volumeVideo
			v.onplaying = () => {
				if(v == currentVideo)
					return
				v.volume = volumeVideo
				v.muted = muteVideo
				if (currentVideo) {
					currentVideo.pause()
				}
				currentVideo = v
			}
			v.onvolumechange = () => {
				volumeVideo = v.volume
				muteVideo = v.muted
			}
			const buttons = videoDocFragment.querySelectorAll(".butspeed") as NodeListOf<HTMLButtonElement>
			buttons.forEach(button => {
				if (button.innerText == "1.0x"){
					button.classList.add("butspeedactive")
				}
				button.onclick = () => {
					v.playbackRate = Number(button.innerText.split("x")[0])
					button.classList.add("butspeedactive")
					buttons.forEach(btn => {
						if(btn != button){
							btn.classList.remove("butspeedactive")
						}
					})
				}
			})
			const videoname = videoDocFragment.querySelector(".videoname") as HTMLParagraphElement
			videoname.innerText = pathName
			post.append(videoDocFragment)
			break
		case "mp3":
			const audioDocFragment = (document.getElementById("templateaudio") as HTMLTemplateElement).content.cloneNode(true) as DocumentFragment
			const ha = audioDocFragment.querySelector("a")!
			ha.setAttribute("href", path)
			ha.innerText = path.split("/").pop()!
			audioDocFragment.querySelector("source")!.setAttribute("src", path)
			const a = audioDocFragment.querySelector("audio")!
			a.volume = volumeMusic
			a.onplaying = () => {
				if(a == currentMusic)
					return
				a.volume = volumeMusic
				a.muted = muteMusic
				if(currentMusic) {
					currentMusic.classList.remove("musicfix")
					currentMusic.pause()
				}
				a.classList.add("musicfix")
				currentMusic = a
			}
			a.onvolumechange = () => {
				volumeMusic = a.volume
				muteMusic = a.muted
			}
			post.append(audioDocFragment)
			break
		case "jpg":
		case "png":
		case "jpeg":
			const i = (document.getElementById("templateimg") as HTMLTemplateElement).content.cloneNode(true) as DocumentFragment
			i.querySelector("img")!.setAttribute("src", path)
			post.append(i)
			break
		default:
			const o = (document.getElementById("templateother") as HTMLTemplateElement).content.cloneNode(true) as DocumentFragment
			const ho = o.querySelector("a")!
			ho.setAttribute("href", path)
			ho.innerText = pathName
			post.append(o)
			break
	}

	//запрашиваем архив комментариев
	let xhr = new XMLHttpRequest()
	xhr.open('GET', path + "._msg", true)
	xhr.setRequestHeader("Cache-Control", "no-store")
	xhr.send()
	xhr.onload = () => {
		if(xhr.status==200){
			comments.innerText = xhr.response + "<history restored>\n\n"
			comments.scrollTo(0, comments.scrollHeight)
		}
	}

	msginput.onkeydown = (event) => {
		if(event.key == 'Enter') {
			if (msginput.value.length == 0) {
				return
			}
			lastMsg = msginput.value//на будущее для вывода последних комментариев
			const msg:msgStruct = {
					id: wrapper.id,
					txt: msginput.value,
				}
			socket.send(JSON.stringify(msg))
			msginput.value = ""
		}
	}
	if(dir == direction.UP) {
		contentBlock.prepend(d)//делается в конце функции т.к. после выполнения данной команды содержимое d недоступно
	} else if(dir == direction.DOWN) {
		contentBlock.append(d)//делается в конце функции т.к. после выполнения данной команды содержимое d недоступно
	}
}