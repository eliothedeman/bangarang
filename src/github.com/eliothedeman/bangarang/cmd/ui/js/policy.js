function PolicyController($scope, $http) {
	$scope.policies = null;
	this.fetchPolicies = function() {
		$http.get("api/policy/config/*").success(function(data, status) {
			$scope.policies = data;
			console.log(data);
		});
	}

	this.addPolicy = function(name, pol) {
		$http.post("api/policy/config/" + name, JSON.stringify(pol))
		this.init();
	}

	this.init = function() {
		this.fetchPolicies();
	}

	this.init();
}

angular.module('bangarang').controller("PolicyController", PolicyController);
