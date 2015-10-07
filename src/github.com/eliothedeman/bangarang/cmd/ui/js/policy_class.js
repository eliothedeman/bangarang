"use strict"

class Match {
	constructor(obj) {
		this.match = obj
	}

	add(key, val) {
		this.match[key] = val
	}

	del(key) {
		delete this.match[key]
	}

	chips() {
		var c = []
		for (key in this.match) {
			c.push(key + " : " + this.match[key])
		}

		return c
	}

	data() {
		return this.match
	}
}

function isObject(o) {
	return typeof o === "object"
}

function parsePolicy(raw) {
	var p = new Policy(raw.name)
	if (isObject(raw.match)) {
		p.match = new Match(raw.match)
	}

	if (isObject(raw.not_match)) {
		p.not_match = new Match(raw.not_match)
	}

	if (isObject(raw.crit)) {
		p.crit = parseCondition(raw.crit) 
	}

	if (isObject(raw.warn)) {
		p.warn = parseCondition(raw.warn)
	}

	return p
}

// Representation of a policy
class Policy {
	constructor(name) {
		this.name = name
		this.match = new Match({})
		this.not_match = new Match({})
		this.crit = null
		this.warn = null
	}

	url() {
		return "api/policy/config/" + this.name
	}

	data() {
		let d = {
			name: this.name,
			match: this.match.data(),
			not_match: this.not_match.data()
		}

		if (this.crit) {
			d.crit = this.crit.data()
		}

		if (this.warn) {
			d.warn = this.warn.data()
		}

		return d
	}
}

function parseCondition(raw) {
	var t = function(r) {
		if (raw.greater != null) {
			return "greater"
		}

		if (raw.less != null) {
			return "less"
		}

		if (raw.exactly != null) {
			return "exactly"
		}

		return "greater"
	}

	switch(true) {
		case raw.std_dev:
			return new StdDev(t(raw), raw[t(raw)], raw.escalation)
			break

		case raw.derivative:
			return new Derivative(t(raw), raw[t(raw)], raw.escalation)
			break

		case raw.holt_winters:
			return new HoltWinters(t(raw), raw[t(raw)], raw.escalation)
			break

		default:
			return new Simple(t(raw), raw[t(raw)], raw.escalation)
	}

}

class Condition {
	constructor(type, value, escalation) {
		this.type = type
		this.value = value
		this.escalation = escalation
		this.window_size = 5
		this.occurences = 1
	}

	types() {
		return [
			"greater",
			"less",
			"exactly"
		]
	}

	data() {
		let d = {
			escalation: this.escalation,
			window_size: this.window_size,
			occurences: this.occurences
		}
		d[this.type] = this.value
		return d
	}
}

class Simple extends Condition {
	data() {
		let d = super.data()
		d.simple = true
		return d
	}
}

class HoltWinters extends Condition {
	data() {
		let d = super.data()
		d.holt_winters = true
		return d
	}
}

class Derivative extends Condition {
	data() {
		let d = super.data()
		d.derivative = true
		return d
	}
}

class StdDev extends Condition {
	data() {
		let d = super.data()
		d.std_dev = true
		return d
	}
}

var d = parsePolicy({warn: {std_dev: true, greater: 100}})
console.log(d.data())