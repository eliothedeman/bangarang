// Representation of a policy
class Policy {
	constructor(name) {
		this.name = name
		this.match = {}
		this.notMatch = {}
	}

	data() {
		return {

		}
	}
}

class Condition {
	constructor(condition, value, modifier) {
		this.condition = condition
		this.value = value
		this.modifier = modifier
	}

	Modifier() {

	}

	data() {
		return {
			this.condition: this.value,

		}
	}


}

function NewPolicyController($scope, $http, $timeout, $mdDialog) {
	$scope.np = {};
	$scope.compOps = ["greater", "less", "exactly"];
	$scope.specialOps = [
		{	display: "Simple",
			name: "simple"
		},
		{
			display: "Derivative",
			name: "derivative"
		},
		{
			display: "Standard Deviation",
			name: "std_dev"
		},
		{
			display: "Holt Winters",
			name: "holt_winters",
			disabled: true
		}
	]

	$scope.cSpec = "simple"
	$scope.wSpec = "simple"

	$scope.loadEscalationNames = function() {
		$scope.escalation_names = [];
		return $timeout(function() {
			$http.get("api/escalation/config/*").success(function(data, status) {
				for (name in data) {
					$scope.escalation_names.push(name);
				}
			}).error(function(data, status) {
				$scope.escalation_names = ["Unable to fetch escalations"];
			});
		}, 650);
	}

	$scope.showIncompleteDialog = function(message) {
		$mdDialog.show(
			$mdDialog.alert()
				.title("Incomplete config")
				.content(message)
				.ok("I agree to fix this.")
		)

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

	$scope.createPolicyStruct = function() {
		if (!$scope.np.name) {
			$scope.showIncompleteDialog("Must name the policy before submitting.");
			return null;
		}

		var p = {
			name: $scope.np.name,
			comment: $scope.np.comment
		};

		// set up match
		if ($scope.matchChips.length > 0) {
			p.match = {};
			for (var i = 0; i < $scope.matchChips.length; i++) {
				p.match[$scope.matchChips[i].key] = $scope.matchChips[i].val;
			}
		}
		if ($scope.notMatchChips.length > 0) {
			p.not_match = {};
			for (var i = 0; i < $scope.notMatchChips.length; i++) {
				p.not_match[$scope.notMatchChips[i].key] = $scope.notMatchChips[i].val;
			}
		}
		if ($scope.critOpChips.length > 0 && $scope.cEsc) {
			p.crit = {
				occurences: $scope.cOcc,
				escalation: $scope.cEsc
			};
			for (var i = 0; i < $scope.critOpChips.length; i++) {
				p.crit[$scope.critOpChips[i].key] = $scope.critOpChips[i].val;
			}

			p.crit[$scope.cSpec] = true;
			p.crit["window_size"] = $scope.cWinSize;
		}
		if ($scope.warnOpChips.length > 0 && $scope.wEsc) {
			p.warn = {
				occurences: $scope.wOcc,
				escalation: $scope.wEsc
			};
			for (var i = 0; i < $scope.warnOpChips.length; i++) {
				p.warn[$scope.warnOpChips[i].key] = $scope.warnOpChips[i].val;
			}
			p.warn[$scope.wSpec] = true;
			p.warn["window_size"] = $scope.wWinSize;
		}

		return p;
	}

	$scope.addPolicy = function() {
		var pol = $scope.createPolicyStruct();
		if (pol) {
			$http.post("api/policy/config/" + pol.name, $scope.createPolicyStruct()).success(function() {
				$scope.reset()
			});
		}
	}

	$scope.cancelPolicy = function() {
		$scope.reset();
	}

	$scope.addNewMatch = function() {
		if ($scope.matchChips == null ) {
			$scope.matchChips = [];
		}
		$scope.matchChips.push({"key": $scope.newMatchKey, "val": $scope.newMatchVal});
		$scope.newMatchKey = "";
		$scope.newMatchVal = "";
	}

	$scope.addNewNotMatch = function() {

		if ($scope.not_matchChips == null) {
			$scope.not_matchChips = [];
		}
		$scope.notMatchChips.push({"key": $scope.newNotMatchKey, "val": $scope.newNotMatchVal});
		$scope.newNotMatchKey = "";
		$scope.newNotMatchVal = "";
	}

	$scope.addNewCritOp = function() {
		if ($scope.cOpKey && ($scope.cOpVal || $scope.cOpVal === 0) ) {
			$scope.critOpChips.push({"key": $scope.cOpKey, "val": $scope.cOpVal});
			$scope.cOpKey = "";
			$scope.cOpVal = "";
		}
	}

	$scope.addNewWarnOp = function() {
		if ($scope.wOpKey && ($scope.wOpVal || $scope.wOpVal === 0 ) ) {
			$scope.warnOpChips.push({"key": $scope.wOpKey, "val": $scope.wOpVal});
			$scope.wOpVal = "";
			$scope.wOpKey = "";
		}
	}

	$scope.init = function() {
		$scope.cOpVal = "";
		$scope.cOpKey = "";
		$scope.wOpVal = "";
		$scope.wOpKey = "";
		$scope.cSpec = "simple"
		$scope.wSpec = "simple"
		$scope.cWinSize = 100;
		$scope.wWinSize = 100;
		$scope.np.name = "";
		$scope.wOcc = 1;
		$scope.cOcc = 1;
		$scope.escalations = [];
		$scope.np.comment = "";
	}

	$scope.reset = function() {
		$scope.init();
		$scope.matchChips = [];
		$scope.notMatchChips = [];
		$scope.critOpChips = [];
		$scope.warnOpChips = [];
	}

	$scope.reset();
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

	collectPolicy = function() {
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
	t = $scope;
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
