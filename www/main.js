"use strict";
var ip = window.location.host

window.onload = function () {

	var socket = new WebSocket("ws://" + ip + "/ws");
	var socketUpload = new WebSocket("ws://" + ip + "/upload");

	function appendToBody(event) {
		var d = document.createElement('div');
		d.className = 'card';
		d.innerHTML = event.data;
		// document.body.append(d);
		var content = document.getElementById('content');
		content.prepend(d)

		// var imgLs = document.getElementsByTagName("img")
		// for(const element of imgLs) {
		// 	element.addEventListener("dragstart", (e)=>{
		// 		e.preventDefault();
		// 	})
		// }
		d.addEventListener("dragstart", (e)=>{
			e.preventDefault();
		})
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

	var btn = document.getElementById('btn');
	var inp = document.getElementById('inp');
	btn.onclick = function () {
		socket.send(inp.value);
	}
	inp.onkeydown = function (event) {
		if(event.key == 'Enter') {
			socket.send(inp.value);
			inp.value = "";
		}
	}

	var sf = document.getElementById("sendfile");
	sf.addEventListener('change', (e) => {
		if (e.target.files[0]) {
			console.log('You selected ' + e.target.files[0].name);
			socketUpload.send(e.target.files[0].name);
			socketUpload.send(e.target.files[0].size);
			console.log('File size: ' + e.target.files[0].size);
			socketUpload.send(e.target.files[0]);
		}
	});

	document.addEventListener("dragover", (e) => {
		e.preventDefault();
	})
	document.addEventListener("drop", (e) => {
		e.preventDefault();
		var fs = e.dataTransfer.files;
		for (let index = 0; index < fs.length; index++) {
			const element = fs[index];
			console.log('You selected ' + element.name);
			socketUpload.send(element.name);
			socketUpload.send(element.size);
			console.log('File size: ' + element.size);
			socketUpload.send(element);
		}
	})
}