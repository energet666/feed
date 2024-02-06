'use strict'

import Vector from './ClassVector.js'

function getRandomColorStr() {
	return ('rgb(' + (Math.random()*255).toFixed() + ',' + (Math.random()*255).toFixed() + ',' + (Math.random()*255).toFixed() + ')');
}

class MyObject {
	constructor(r) {
		this.freez = false;
		this.checked = false;
		this._x = 0;
		this._y = 0;
		this._radius = 0;
		this.speedV = new Vector(  MyObject.speedMax * (1 - 2 * Math.random()), MyObject.speedMax * (1 - 2 * Math.random())  );
		
		this.el = document.createElement('div');
		this.el.classList.add('rect');
		this.angleRedraw();
		
		this.radius = r || MyObject.radiusDefault || MyObject.radiusMin + (MyObject.radiusMax - MyObject.radiusMin) * Math.random();
		this.x = this.radius + (MyObject.w - 2 * this.radius) * Math.random();
		this.y = this.radius + (MyObject.h - 2 * this.radius) * Math.random();
		this.color = getRandomColorStr();
	}

	set x(val) {
		this._x = val;
		this.el.style.left = this._x - this._radius;
	}
	get x() {
		return this._x;
	}

	set y(val) {
		this._y = val;
		this.el.style.top = this._y - this._radius;
	}
	get y() {
		return this._y;
	}

	positionRedraw() {
		this.el.style.left = this._x - this._radius;
		this.el.style.top = this._y - this._radius;
	}

	angleRedraw() {
		this.el.style.transform = 'rotate(' + (this.speedV.angleGrad + 45) + 'deg)';
	}

	set radius(val) {
		if (val >= MyObject.h / 2) {
			val = MyObject.h / 2;
		}
		if (val >= MyObject.w / 2) {
			val = MyObject.w / 2;
		}
		this._radius = val;
		this.el.style.height = this._radius * 2 + 'px';
		this.el.style.width = this._radius * 2 + 'px';
	}
	get radius() {
		return this._radius;
	}

	set color(val) {
		this.el.style.borderColor = val;
	}

	get mass() {
		return this.radius * this.radius;
	}

	move() {
		if (this.freez) {
			return;
		}
		this._x += this.speedV.x;
		this._y += this.speedV.y;

		if ( (this._x + this.radius) > MyObject.w ) {
			this.speedV.x = -this.speedV.x;
			this._x = MyObject.w - this.radius;
			this.onContact();
		} else if ( this._x - this.radius < 0 ) {
			this.speedV.x = -this.speedV.x;
			this._x = this.radius;
			this.onContact();
		}

		if ( (this._y + this.radius) > MyObject.h ) {
			this.speedV.y = -this.speedV.y;
			this._y = MyObject.h - this.radius;
			this.onContact();
		} else if ( this._y - this.radius < 0 ) {
			this.speedV.y = -this.speedV.y;
			this._y = this.radius;
			this.onContact();
		}

		this.positionRedraw();

		// this.speed *= 0.999;
		// this.size *= 1.0001;
	}
	onContact() {
		this.angleRedraw();
		// this.color = getRandomColorStr();
		// this.el.style.backgroundColor = getRandomColorStr();
		// document.body.style.backgroundColor = getRandomColorStr();
	}
	insertToHTML() {
		document.body.appendChild(this.el);
	}
}
MyObject.h = window.innerHeight;
MyObject.w = window.innerWidth;

MyObject.speedMax = 5;
MyObject.radiusMin = 10;
MyObject.radiusMax = 100;
MyObject.radiusDefault = 0;

export default MyObject;