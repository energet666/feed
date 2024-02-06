'use strict'

class Vector {
	/**
	 * @param {Number} x 
	 * @param {Number} y 
	 */
	constructor(x = 1, y = 0) {
		this.x = x;
		this.y = y;
	}
	/**
	 * @param {Number} val lenght to set
	 */
	set module(val) {
		if (this.module == 0) {
			this.x = val;
			this.y = 0;
		} else {
			let k = val / this.module;
			this.x *= k;
			this.y *= k;
		}
	}
	/**
	 * @returns {Number} lenght of this Vector
	 */
	get module() {
		return Math.sqrt(this.x * this.x + this.y * this.y);
	}
	/**
	 * @returns {Number} lenght of this Vector pow 2
	 */
	getModulePow2() {
		return (this.x * this.x + this.y * this.y);
	}

	/**
	 * @returns {Number} angle in grad betwen this Vector and X axis
	 */
	get angleGrad() {
		return Math.atan2(this.y, this.x) * 180 / Math.PI;
	}

	/**
	 * @returns {Number} projecton this Vector to v Vector
	 * @param {Vector} v
	 */
	projectionScalar(v) {
		return ( this.x * v.x + this.y * v.y ) / v.module;
	}
	projectionVector(v) {
		let k =  (this.x * v.x + this.y * v.y) / v.getModulePow2();
		return v.mul(k);
	}

	/**
	 * @returns {Vector}
	 * @param {Vector} v 
	 */
	add(v) {
		return new Vector(this.x + v.x, this.y + v.y);
	}

	/**
	 * @returns {Vector}
	 * @param {Vector} v 
	 */
	sub(v) {
		return new Vector(this.x - v.x, this.y - v.y);
	}

	mul(k) {
		return new Vector(this.x * k, this.y * k);
	}
}

export default Vector;