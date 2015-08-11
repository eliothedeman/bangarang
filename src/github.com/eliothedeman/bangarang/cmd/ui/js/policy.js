function NewPolicyController($scope, $http, $timeout, $mdDialog) {
	$scope.np = {};
	$scope.compOps = ["greater", "less", "exactly"];

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

	$scope.createPolicyStruct = function() {
		if (!$scope.np.name) {
			$scope.showIncompleteDialog("Must name the policy before submitting.");
			return null;
		}

		var p = {
			name: $scope.np.name
		};

		// set up match
		if ($scope.matchChips.length > 0) {
			p.match = {};
			for (var i = 0; i < $scope.matchChips.length; i++) {
				p.match[$scope.matchChips[i].key] = $scope.matchChips[i].val;
			}
		}
		if ($scope.notMatchChips.length > 0) {
			p.not_match = {
				occurences: $scope.wOcc
			};
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
		}
		if ($scope.warnOpChips.length > 0 && $scope.wEsc) {
			p.warn = {
				occurences: $scope.wOcc,
				escalation: $scope.wEsc
			};
			for (var i = 0; i < $scope.warnOpChips.length; i++) {
				p.warn[$scope.warnOpChips[i].key] = $scope.warnOpChips[i].val;
			}
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
		if ($scope.cOpKey && $scope.cOpVal ) {
			$scope.critOpChips.push({"key": $scope.cOpKey, "val": $scope.cOpVal});
			$scope.cOpKey = "";
			$scope.cOpVal = "";
		}
	}

	$scope.addNewWarnOp = function() {
		if ($scope.wOpKey && $scope.wOpVal ) {
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
		$scope.wEsc = "";
		$scope.cEsc = "";
		$scope.np.name = "";
		$scope.wOcc = 1;
		$scope.cOcc = 1;
		$scope.escalations = [];
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
			.cancel("No");

		$mdDialog.show(confirm);

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
		});
	}

	$scope.fetchPolicies = function() {
		$http.get("api/policy/config/*").success(function(data, status) {
			$scope.policies = data;
		});
	}
	$scope.init = function() {
		$scope.fetchPolicies();
	}

	$scope.init();
}

angular.module('bangarang').controller("PolicyController", PolicyController);
