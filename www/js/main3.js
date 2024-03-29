import tools from './functions.js';
tools.test("hi from module");
const ip = window.location.host;
let currentVideo = null;
let volumeVideo = 0.25;
let muteVideo = false;
let currentMusic = null;
let volumeMusic = 0.25;
let muteMusic = false;
let lastMsg = "";
//let ring = new Audio("./upload/Iphone - Message Tone.mp3")
let topPost = null;
let bottomPost = null;
let cardTemplate;
let contentBlock;
let uploadingOverlay;
let socket;
let socketUpload;
let socketEvent;
window.onload = () => {
    socket = new WebSocket("ws://" + ip + "/ws");
    socketUpload = new WebSocket("ws://" + ip + "/uploadws");
    socketEvent = new WebSocket("ws://" + ip + "/eventws");
    cardTemplate = document.getElementById("templatecard");
    contentBlock = document.getElementById('content');
    uploadingOverlay = document.querySelector(".over");
    socket.onclose = () => { console.log(`WS "ws" Disconnected`); };
    socket.onerror = () => { console.log(`WS "ws" Error`); };
    socket.onopen = () => { console.log(`WS "ws" Connected`); };
    socket.onmessage = addComment;
    socketUpload.onmessage = (e) => {
        appendToBody(e);
        endContentCheck();
    };
    //При окрытии вебсокета socketUpload запрашиваем файл с индексом "-1"(сервер ответит самым свежим файлом)
    socketUpload.onopen = () => {
        socketUpload.send("-1");
    };
    uploadingOverlay.ondragleave = (e) => {
        e.preventDefault();
        uploadingOverlay.classList.remove("show");
    };
    document.ondragenter = (e) => {
        e.preventDefault();
        uploadingOverlay.innerText = 'Drop it!';
        uploadingOverlay.classList.add("show");
    };
    document.ondragover = (e) => { e.preventDefault(); };
    document.ondragstart = (e) => { e.preventDefault(); };
    document.ondrop = uploadFile;
    document.addEventListener("scroll", function scrollHandler() {
        endContentCheck();
        //ограничение частоты срабатывания scroll Event
        document.removeEventListener("scroll", scrollHandler);
        setTimeout(() => {
            document.addEventListener("scroll", scrollHandler);
        }, 100);
    });
    const snapOff = document.getElementById("snap_off");
    snapOff.checked = false;
    snapOff.onchange = () => {
        if (snapOff.checked) {
            document.documentElement.classList.add("snappingOn");
        }
        else {
            document.documentElement.classList.remove("snappingOn");
        }
    };
};
const uploadFile = (e) => {
    e.preventDefault();
    const files = e.dataTransfer.files;
    //Запрещаем пользователю загружать много файлов за раз
    if (files.length > 1) {
        uploadingOverlay.innerText = "One at a time! Try it again.";
        setTimeout(() => {
            uploadingOverlay.classList.remove("show");
        }, 3000);
        return;
    }
    const file = files[0];
    console.log('You selected ' + file.name);
    console.log('File size: ' + file.size);
    let formData = new FormData();
    formData.append("file", file);
    let xhr = new XMLHttpRequest();
    //Отображаем прогресс загрузки
    xhr.upload.onprogress = (e) => {
        const ev = e;
        const uploadProgress = (ev.loaded / ev.total * 100).toFixed(0);
        console.log("Upload progress: " + uploadProgress + "%");
        uploadingOverlay.innerText = "Upload progress: " + uploadProgress + "%";
    };
    xhr.upload.onloadend = () => {
        uploadingOverlay.classList.remove("show");
    };
    xhr.open('POST', "/uploadxml/" + file.name, true);
    xhr.send(formData);
};
const endContentCheck = () => {
    //Если приближаемся к нижнему концу ленты на высоту окна, то запрашиваем более старые посты
    //и при необходимости удаляем вышедшие из зоны видимости на 3 высоты окна верхние посты
    if (document.documentElement.scrollHeight - window.innerHeight - document.documentElement.scrollTop < window.innerHeight) {
        if (bottomPost > 0) {
            socketUpload.send(String(bottomPost - 1));
            while (document.documentElement.scrollTop > 3 * window.innerHeight) {
                let cards = contentBlock.getElementsByClassName("card");
                //если карточка оказалась последней, то выходим из цикла
                if (cards.length <= 1) {
                    break;
                }
                cards[0].remove();
                topPost--;
            }
        }
    }
    else if (document.documentElement.scrollTop < window.innerHeight) {
        socketUpload.send(String(topPost + 1));
        while (document.documentElement.offsetHeight - window.innerHeight - document.documentElement.scrollTop > 3 * window.innerHeight) {
            let cards = contentBlock.getElementsByClassName("card");
            if (cards.length <= 1) {
                break;
            }
            cards[cards.length - 1].remove();
            bottomPost++;
        }
    }
};
const addComment = (e) => {
    const data = JSON.parse(e.data);
    const target = document.getElementById(data.id);
    if (target) {
        const commentDiv = target.querySelector(".comments");
        commentDiv.innerText += data.txt + "\n";
        commentDiv.scrollTo(0, commentDiv.scrollHeight);
        //ring.play()
    }
};
const appendToBody = (event) => {
    const data = JSON.parse(event.data);
    let direction;
    (function (direction) {
        direction[direction["UP"] = 1] = "UP";
        direction[direction["DOWN"] = 0] = "DOWN";
    })(direction || (direction = {}));
    let dir = direction.DOWN;
    if (topPost != null && bottomPost != null) {
        if (data.N === bottomPost - 1) {
            dir = direction.DOWN;
            bottomPost--;
        }
        else if (data.N === topPost + 1) {
            dir = direction.UP;
            topPost++;
        }
        else {
            return;
        }
    }
    else {
        topPost = data.N;
        bottomPost = data.N;
    }
    const d = cardTemplate.content.cloneNode(true);
    const post = d.querySelector(".post");
    const comments = d.querySelector(".comments");
    const msginput = d.querySelector(".msginput");
    const wrapper = d.querySelector(".wrapper");
    const path = data.Path;
    const lastSlash = path.lastIndexOf("/");
    const pathDir = path.slice(0, lastSlash + 1);
    const pathName = path.slice(lastSlash + 1);
    //Экранируем спец символы в названии файла чтобы получить валидный URL ресурса
    const normalizedPath = pathDir + encodeURIComponent(pathName);
    wrapper.id = path;
    //формируем пост в зависимости от типа контента
    switch (tools.getExt(path)) {
        case "mp4":
            const videoDocFragment = document.getElementById("templatevideo").content.cloneNode(true);
            const v = videoDocFragment.querySelector("video");
            v.setAttribute("src", normalizedPath);
            v.volume = volumeVideo;
            v.onplaying = () => {
                if (v == currentVideo)
                    return;
                v.volume = volumeVideo;
                v.muted = muteVideo;
                if (currentVideo) {
                    currentVideo.pause();
                }
                currentVideo = v;
            };
            v.onvolumechange = () => {
                volumeVideo = v.volume;
                muteVideo = v.muted;
            };
            const buttons = videoDocFragment.querySelectorAll(".butspeed");
            buttons.forEach(button => {
                if (button.innerText == "1.0x") {
                    button.classList.add("butspeedactive");
                }
                button.onclick = () => {
                    v.playbackRate = Number(button.innerText.split("x")[0]);
                    button.classList.add("butspeedactive");
                    buttons.forEach(btn => {
                        if (btn != button) {
                            btn.classList.remove("butspeedactive");
                        }
                    });
                };
            });
            const videoname = videoDocFragment.querySelector(".videoname");
            videoname.innerText = pathName;
            post.append(videoDocFragment);
            break;
        case "mp3":
            const audioDocFragment = document.getElementById("templateaudio").content.cloneNode(true);
            const ha = audioDocFragment.querySelector("a");
            ha.setAttribute("href", path);
            ha.innerText = path.split("/").pop();
            audioDocFragment.querySelector("source").setAttribute("src", normalizedPath);
            const a = audioDocFragment.querySelector("audio");
            a.volume = volumeMusic;
            a.onplaying = () => {
                if (a == currentMusic)
                    return;
                a.volume = volumeMusic;
                a.muted = muteMusic;
                if (currentMusic) {
                    currentMusic.classList.remove("musicfix");
                    currentMusic.pause();
                }
                a.classList.add("musicfix");
                currentMusic = a;
            };
            a.onvolumechange = () => {
                volumeMusic = a.volume;
                muteMusic = a.muted;
            };
            post.append(audioDocFragment);
            break;
        case "jpg":
        case "png":
        case "jpeg":
        case "gif":
            const i = document.getElementById("templateimg").content.cloneNode(true);
            i.querySelector("img").setAttribute("src", normalizedPath);
            post.append(i);
            break;
        default:
            const o = document.getElementById("templateother").content.cloneNode(true);
            const ho = o.querySelector("a");
            ho.setAttribute("href", normalizedPath);
            ho.innerText = pathName;
            post.append(o);
            break;
    }
    //запрашиваем архив комментариев
    let xhr = new XMLHttpRequest();
    xhr.open('GET', normalizedPath + "._msg", true);
    xhr.setRequestHeader("Cache-Control", "no-store");
    xhr.send();
    xhr.onload = () => {
        if (xhr.status == 200) {
            comments.innerText = xhr.response + "<history restored>\n\n";
            comments.scrollTo(0, comments.scrollHeight);
        }
    };
    msginput.onkeydown = (event) => {
        if (event.key == 'Enter') {
            if (msginput.value.length == 0) {
                return;
            }
            lastMsg = msginput.value; //на будущее для вывода последних комментариев
            const msg = {
                id: wrapper.id,
                txt: msginput.value,
            };
            socket.send(JSON.stringify(msg));
            msginput.value = "";
        }
    };
    if (dir == direction.UP) {
        contentBlock.prepend(d);
    }
    else if (dir == direction.DOWN) {
        contentBlock.append(d);
    }
};
