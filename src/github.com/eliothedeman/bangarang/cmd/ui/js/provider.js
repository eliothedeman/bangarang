function ProviderController($scope, $http) {
	$scope.providers = {};
	this.fetchProviders = function() {
		$http.get("/api/provider/config/*").success(function(data,status) {
			for (name in data) {
				$scope.providers[name] = data[name];
			}
		});
	}
}
angular.module("bangarang").controller("ProviderController", ProviderController);
function NewProvider($scope,$http) {

	this.init = function() {
		$scope.name = "";
		$scope.type = "";
	}

	this.init();
}

angular.module('bangarang').controller('NewProvider', NewProvider);