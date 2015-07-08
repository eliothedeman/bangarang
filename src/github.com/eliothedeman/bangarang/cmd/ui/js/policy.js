angular.module('bangarang', []).controller("PolicyController", function($scope, $http) {
	$scope.policies = {};
	this.fetchPolicies = function() {
		$http.get("api/policy/config/*").success(function(data, status) {
			$scope.policies = data;
		});
	}

	this.addPolicy = function(name, pol) {
		$http.post("api/policy/config/" + name, JSON.stringify(pol))
	}

	this.init = function() {
		this.fetchPolicies();
	}

	this.init();
});
