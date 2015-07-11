function PolicyController($scope, $http) {
	$scope.policies = null;
	this.np = {};
	this.compOps = ["greater", "less", "exactly"];
	this.fetchPolicies = function() {
		$http.get("api/policy/config/*").success(function(data, status) {
			$scope.policies = data;
		});
	}

	this.fetchEscalations = function() {
		$http.get("api/escalation/config/*").success(function(data, status) {
			this.escalations = data;
		});
	}

	this.createPolicyStruct = function() {
		if (!this.np.name) {
			return null;
		}

		var p = {
			name: this.np.name
		};

		// set up match
		if (this.matchChips.length > 0) {
			p.match = {};
			for (var i = 0; i < this.matchChips.length; i++) {
				p.match[this.matchChips[i].key] = this.matchChips[i].val;
			}
		}
		if (this.notMatchChips.length > 0) {
			p.not_match = {};
			for (var i = 0; i < this.notMatchChips.length; i++) {
				p.not_match[this.notMatchChips[i].key] = this.notMatchChips[i].val;
			}
		}
		if (this.critOpChips.length > 0) {
			p.crit = {};
			for (var i = 0; i < this.critOpChips.length; i++) {
				p.crit[this.critOpChips[i].key] = this.critOpChips[i].val;
			}
		}
		if (this.warnOpChips.length > 0) {
			p.warn = {};
			for (var i = 0; i < this.warnOpChips.length; i++) {
				p.warn[this.warnOpChips[i].key] = this.warnOpChips[i].val;
			}
		}
		return p;
	}

	this.addPolicy = function() {
		var pol = this.createPolicyStruct();
		if (pol) {
			$http.post("api/policy/config/" + pol.name, this.createPolicyStruct());
			this.reset();
		}
	}

	this.cancelPolicy = function() {
		this.reset();
	}

	this.addNewMatch = function() {
		var np = this.np;
		if (np.match != {}) {
			np.match = {}
		}
		np.match[np.newMatchKey] = np.newMatchValue;
		if (np.matchChips != []) {
			np.matchChips = [];
		}
		this.matchChips.push({"key": np.newMatchKey, "val": np.newMatchValue});
		np.newMatchKey = "";
		np.newMatchVal = "";
	}

	this.addNewNotMatch = function() {
		var np = this.np;
		if (np.not_match != {}) {
			np.match = {}
		}
		np.match[np.newNotMatchKey] = np.newNotMatchValue;

		if (np.not_matchChips != []) {
			np.not_matchChips = [];
		}
		this.notMatchChips.push({"key": np.newNotMatchKey, "val": np.newNotMatchValue});
		np.newNotMatchKey = "";
		np.newNotMatchVal = "";
	}

	this.addNewCritOp = function() {
		var np = this.np;
		if (np.cOpKey && np.cOpVal ) {
			this.critOpChips.push({"key": np.cOpKey, "val": np.cOpVal});
			np.cOpKey = "";
			np.cOpVal = "";
		}
	}
	this.addNewWarnOp = function() {
		var np = this.np;
		if (np.wOpKey && np.wOpVal ) {
			this.warnOpChips.push({"key": np.wOpKey, "val": np.wOpVal});
			np.wOpVal = "";
			np.wOpKey = "";
		}
	}



	this.init = function() {
		this.np = {
			cOpVal: "",
			cOpKey: "",
			wOpVal: "",
			wOpKey: "",
			wOcc: 1,
			cOcc: 1
		};
		this.fetchPolicies();
	}

	this.reset = function() {
		this.init();
		this.matchChips = [];
		this.notMatchChips = [];
		this.critOpChips = [];
		this.warnOpChips = []
	}

	this.reset();
}

angular.module('bangarang').controller("PolicyController", PolicyController);
