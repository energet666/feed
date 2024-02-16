import { test } from './functions.js';
window.onload = function () {
    test("hi from module");
    const ip = window.location.host;
    let currentVideo = null;
    let volumeVideo = 0.25;
    let muteVideo = false;
    let currentMusic = null;
    let volumeMusic = 0.25;
    let muteMusic = false;
    let lastMsg = "";
    let ring = new Audio("./upload/Iphone - Message Tone.mp3");
    const cardTemplate = document.getElementById("templatecard");
    const contentBlock = document.getElementById('content');
    const socket = new WebSocket("ws://" + ip + "/ws");
    const socketUpload = new WebSocket("ws://" + ip + "/uploadws");
    const socketEvent = new WebSocket("ws://" + ip + "/eventws");
    function appendToBody(event) {
        const d = cardTemplate.content.cloneNode(true);
        const post = d.querySelector(".post");
        const comments = d.querySelector(".comments");
        const msginput = d.querySelector(".msginput");
        const wrapper = d.querySelector(".wrapper");
        const path = event.data;
        const pathName = path.split("/").pop();
        wrapper.id = path;
        let pathNameSplit = pathName.split(".");
        let pathNameExt = "";
        // If spl.length is one, it's a visible file with no extension ie. file
        // If spl[0] === "" and spl.length === 2 it's a hidden file with no extension ie. .htaccess
        if (!(pathNameSplit.length === 1 || (pathNameSplit[0] === "" && pathNameSplit.length === 2))) {
            pathNameExt = pathNameSplit.pop().toLowerCase();
        }
        switch (pathNameExt) {
            case "mp4":
                const videoDocFragment = document.getElementById("templatevideo").content.cloneNode(true);
                const v = videoDocFragment.querySelector("video");
                v.setAttribute("src", path);
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
                    button.onclick = () => {
                        v.playbackRate = Number(button.innerText.split("x")[0]);
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
                audioDocFragment.querySelector("source").setAttribute("src", path);
                const a = audioDocFragment.querySelector("audio");
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
                const i = document.getElementById("templateimg").content.cloneNode(true);
                i.querySelector("img").setAttribute("src", path);
                post.append(i);
                break;
            default:
                const o = document.getElementById("templateother").content.cloneNode(true);
                const ho = o.querySelector("a");
                ho.setAttribute("href", path);
                ho.innerText = pathName;
                post.append(o);
                break;
        }
        let xhr = new XMLHttpRequest();
        xhr.open('GET', path + "._msg", true);
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
                const p = msginput.parentElement;
                if (msginput.value.length == 0) {
                    return;
                }
                lastMsg = msginput.value; //на будущее для вывода последних комментариев
                socket.send(JSON.stringify({
                    id: p.id,
                    txt: msginput.value
                }));
                msginput.value = "";
            }
        };
        contentBlock.prepend(d); //делается в конце функции т.к. после выполнения данной команды содержимое d недоступно
    }
    socketUpload.onmessage = appendToBody;
    socket.addEventListener("message", (e) => {
        const data = JSON.parse(e.data);
        const com = document.getElementById(data.id).querySelector(".comments");
        com.innerText += data.txt + "\n";
        com.scrollTo(0, com.scrollHeight);
        //ring.play()
    });
    socket.onclose = () => {
        console.log('Service', "WebSocket Disconnected");
    };
    socket.onerror = () => {
        console.log('Service', "WebSocket Error");
    };
    socket.onopen = () => {
        console.log('Service', "WebSocket Connected");
        // socket.send('Ураааааа!')
    };
    const uploadingOverlay = document.querySelector(".over");
    document.addEventListener("dragenter", (e) => {
        e.preventDefault();
        uploadingOverlay.classList.add("show");
    });
    uploadingOverlay.addEventListener("dragleave", (e) => {
        e.preventDefault();
        uploadingOverlay.classList.remove("show");
    });
    document.addEventListener("dragover", (e) => {
        e.preventDefault();
    });
    document.addEventListener("dragstart", (e) => {
        e.preventDefault();
    });
    document.addEventListener("drop", (e) => {
        e.preventDefault();
        let files = e.dataTransfer.files;
        for (let index = 0; index < files.length; index++) {
            const file = files[index];
            console.log('You selected ' + file.name);
            console.log('File size: ' + file.size);
            let formData = new FormData();
            formData.append("file", file);
            let xhr = new XMLHttpRequest();
            xhr.upload.onprogress = (e) => {
                const ev = e;
                const uploadProgress = (ev.loaded / ev.total * 100).toFixed(0);
                console.log("Upload progress: " + uploadProgress + "%");
                uploadingOverlay.innerText = "Upload progress: " + uploadProgress + "%";
            };
            xhr.upload.onloadend = () => {
                uploadingOverlay.classList.remove("show");
                uploadingOverlay.innerText = "Drop it!";
            };
            xhr.open('POST', "/uploadxml/" + file.name, true);
            xhr.send(formData);
        }
    });
    function snappingOn() {
        document.documentElement.classList.add("snappingOn");
        document.removeEventListener("scroll", snappingOn);
    }
    document.addEventListener("scroll", snappingOn); //при начальной загрузке карточек из-за снаппинга лента сама скролится вниз,
    //поэтому включаю снаппинг когда скролит юзер
    const snapOff = document.getElementById("snap_off");
    snapOff.checked = true;
    snapOff.onchange = () => {
        if (snapOff.checked) {
            document.documentElement.classList.add("snappingOn");
        }
        else {
            document.documentElement.classList.remove("snappingOn");
        }
    };
    // let q = document.createElement("div")
    // q.id = "point"
    // q.classList.add("rect")
    // let w = 30
    // q.style.width = w + "px"
    // q.style.height = w + "px"
    // document.body.appendChild(q)
    // document.addEventListener("mousemove", (e)=>{
    //     const ev = e
    //     console.log(ev.clientX, ev.clientY)
    //     socketEvent.send(JSON.stringify({
    //         X: ev.clientX,
    //         Y: ev.clientY
    //     }))
    // })
    // socketEvent.onmessage = (e) => {
    //     const d = JSON.parse(e.data)
    //     q.style.top = d.Y - w/2 + "px"
    //     q.style.left = d.X - w/2 + "px"
    // }
};
