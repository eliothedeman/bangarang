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
		this.comment = ""
		this.match = new Match({})
		this.not_match = new Match({})
		this.crit = null
		this.warn = null
		this.modifiers = [
			{
				name: "simple",
				title: "Simple",
				factory: function() {
					return new Simple()
				}
			},
			{
				name: "derivative",
				title: "Derivative",
				factory: function() {
					return  new Derivative()
				}
			},
			{
				name: "std_dev",
				title: "Standard Deviation",
				factory: function() {
					return new StdDev()
				}

			},
			{
				name: "holt_winters",
				title: "Holt Winters",
				factory: function() {
					return new HoltWinters()
				}
			}
		]
	}

	getPolicy(type) {
		for (var i = 0; i < this.modifiers.length; i++) {
			if (this.modifiers[i].name == type) {
				return this.modifiers[i].factory()
			}
		}

		return new Simple()
	}


	addWarn(type) {
		console.log(type)
		this.warn = this.getPolicy(type)
	}

	addCrit(type) {
		this.crit = this.getPolicy(type)
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
	constructor() {
		this.type = "greater"
		this.value = 0.0
		this.escalation = ""
		this.window_size = 5
		this.occurences = 1
		this.types = [
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
	constructor() {
		super()
		console.log(this)
	}
	data() {
		let d = super.data()
		d.simple = true
		return d
	}
}

class HoltWinters extends Condition {
	constructor() {
		super()
	}

	data() {
		let d = super.data()
		d.holt_winters = true
		return d
	}
}

class Derivative extends Condition {
	constructor() {
		super()
		this.window_size = 2
	}
	data() {
		let d = super.data()
		d.derivative = true
		return d
	}
}

class StdDev extends Condition {
	constructor() {
		super()
	}
	data() {
		let d = super.data()
		d.std_dev = true
		return d
	}
}

function NewPolicyController($scope, $http, $timeout, $mdDialog) {
	$scope.np = new Policy("")
	$scope.escalation_names = []
	$scope.policies = []

	$scope.fetchPolicies = function() {
		$http.get("api/policy/config/*").success(function(data, status) {
			for (key in data) {
				$scope.policies.push(parsePolicy(data[key]))
			}
		});
	}

	$scope.updateCurrent = function(name) {
		for (var i = 0; i < $scope.policies.length; i++) {
			if ($scope.policies[i].name == name) {
				$scope.np = $scope.policies[i]
			}
		}
	}

	$scope.loadEscalationNames = function() {
		$scope.escalation_names = [];
		return $timeout(function() {
			$http.get("api/escalation/config/*").success(function(data, status) {
				if (data == "null") {
					$scope.escalation_names = []
					return
				}
				for (name in data) {
					$scope.escalation_names.push(name);
				}
			}).error(function(data, status) {
				$scope.escalation_names = ["Unable to fetch escalations"];
			});
		}, 650);
	}

	$scope.addNewMatch = function() {
		$scope.np.match.add($scope.nmk, $scope.nmv)
		$scope.nmk = ""
		$scope.nmv = ""
	}

	$scope.addNewNotMatch = function() {
		$scope.np.not_match.add($scope.nnmk, $scope.nnmv)
		$scope.nnmk = ""
		$scope.nnmv = ""
	}

	$scope.showIncompleteDialog = function(message) {
		$mdDialog.show(
			$mdDialog.alert()
				.title("Incomplete config")
				.content(message)
				.ok("I agree to fix this.")
		)

	}

	$scope.newCondition = function(type) {
		switch (type) {
			case "Standard Deviation":
				return new StdDev()

			case "Derivative":
				return new Derivative()

			case "Holt Winters":
				return new HoltWinters()

			default:
				return new Simple()
		}
	}

	$scope.estimateMemFootprint = function(s) {
		crit = 0;
		warn = 0;
		baseConditionFootprint = function() {
			operators = 24
			specials = 4
			options = 16
			status = 80
			return operators + specials + options + status
		}

		dataFrameFootprint = function(size) {
			return 8 + (size * 8)
		}

		if (s.crit) {
			crit += baseConditionFootprint()
			if (s.crit.window_size) {
				crit += dataFrameFootprint(s.crit.window_size)

			} else {
				crit += dataFrameFootprint(100)
			}
		}
		if (s.warn) {
			warn += baseConditionFootprint()
			if (s.warn.window_size) {
				warn += dataFrameFootprint(s.warn.window_size)

			} else {
				warn += dataFrameFootprint(100)
			}
		}

		return {
			crit: crit,
			warn: warn,
			total: crit+warn
		}
	}

	$scope.addPolicy = function() {
		var pol = $scope.np.data();
		if (pol) {
			$http.post("api/policy/config/" + pol.name, pol).success(function() {
				$scope.reset()
			});
		}
	}

	$scope.updatePolicy = function() {
		$scope.removePolicy($scope.np.name)
		var old = $scope.np
		$scope.addPolicy()
		$scope.np = old
	}

	$scope.removePolicy = function(name) {
		$http.delete("api/policy/config/" + name).then(function() {
			console.log("removed policy " + name)
		}, function(resp) {
			alert(resp.data)
		})

	}

	$scope.cancelPolicy = function() {
		$scope.reset();
	}

	$scope.reset = function() {
		$scope.np = new Policy("")
	}

}
angular.module('bangarang').controller("NewPolicyController", NewPolicyController);

function GlobalPolicyController($scope, $http, $cookies, $mdDialog) {
	$scope.addNewMatch = function() {
		$scope.matchChips.push({key: $scope.newMatchKey, val: $scope.newMatchVal})
		$scope.newMatchKey = "";
		$scope.newMatchVal = "";
	}

	$scope.addNewNotMatch = function() {
		$scope.notMatchChips.push({key: $scope.newNotMatchKey, val: $scope.newNotMatchVal})
		$scope.newNotMatchKey = "";
		$scope.newNotMatchVal = "";
	}

	$scope.populateChips = function() {
		for (k in $scope.g.match) {
			$scope.matchChips.push({key:k, val:$scope.g.match[k]})
		}
		for (k in $scope.g.not_match) {
			$scope.notMatchChips.push({key:k, val:$scope.g.not_match[k]})
		}
	}

	var collectPolicy = function() {
		var d = {
			match: {},
			not_match: {}
		};
		for (var i = $scope.matchChips.length - 1; i >= 0; i--) {
			d.match[$scope.matchChips[i].key] = $scope.matchChips[i].val
		};
		for (var i = $scope.notMatchChips.length - 1; i >= 0; i--) {
			d.not_match[$scope.notMatchChips[i].key] = $scope.notMatchChips[i].val
		};

		return d;
	}

	$scope.submit = function() {
		var c = $mdDialog.confirm()
			.title("Submit global policy?")
			.content("Are you sure you want to modify the global policy?")
			.ariaLabel("Global Policy submit")
			.ok("Yes")
			.cancel("No")

		$mdDialog.show(c).then(function() {
			$http.post("api/policy/config/global", collectPolicy()).success(function() {
				$scope.reset();
				$scope.fetchPolicy();
			});
		}, function(){

		})

	}

	$scope.fetchPolicy = function() {
		$http.get("api/policy/config/global").success(function(data){
			$scope.g = data;
			$scope.populateChips();
		})
	}

	$scope.reset = function() {
		$scope.g = {};
		$scope.matchChips = [];
		$scope.notMatchChips = [];
		$scope.newMatchKey = "";
		$scope.newMatchVal = "";
		$scope.newNotMatchKey = "";
		$scope.newNotMatchVal = "";
	}

	$scope.reset();
}

angular.module('bangarang').controller("GlobalPolicyController", GlobalPolicyController);

function PolicyController($scope, $http, $cookies) {
	$scope.policies = null;
	$scope.removeSure = {};
	var t = $scope;
	$scope.tmpEdit = {};
	$scope.tmpEditName = ""

	$scope.selected = 0;
	$scope.getSelected = function() {
		var s = $cookies.get("pol:tab");
		if (s) {
			$scope.selected = s;
		}
		return $scope.selected;
	}

	$scope.updateSelected = function(name) {
		$cookies.put("pol:tab", name);
		$scope.selected = name;
	}

	$scope.showRemoveDialog = function(name) {
		$scope.removeSure[name] = true;
	}

	$scope.hideRemoveDialog = function(name) {
		$scope.removeSure[name] = false;
	}

	$scope.shouldHideRemoveDialog = function(name) {
		var show = $scope.removeSure[name];
		return show != true;
	}

	$scope.removePolicy = function(name) {
		$http.delete("api/policy/config/"+name).success(function(data) {
			t.fetchPolicies();
			$scope.hideRemoveDialog(name);
		});
	}

	$scope.addPolicy = function(name, data) {
		$http.post("api/policy/config/" + name, data).then(function() {
			$scope.fetchPolicies()
		}, function(resp) {
			alert(resp.data);
		})
	}

	$scope.fetchPolicies = function() {
		$http.get("api/policy/config/*").success(function(data, status) {
			$scope.policies = data;
		});
	}

	$scope.init = function() {
		$scope.fetchPolicies();
		$scope.tmp_edit = {};
		$scope.tmpEditName = ""
	}

	$scope.editPolicy = function(name, data) {
		// remove the old policy
		$scope.removePolicy(name);

		// send in the new one
		$scope.addPolicy(name, data);
	}

	$scope.init();
}

angular.module('bangarang').controller("PolicyController", PolicyController);
