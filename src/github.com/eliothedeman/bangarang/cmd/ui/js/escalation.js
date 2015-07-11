function EscalationController($scope, $http) {
	this.fetchEscalations = function() {
		$http.get("api/escalation/config/*").success(function(data, status) {
			$scope.escalations = data;
		});
	}
}
angular.module("bangarang").controller("EscalationController", EscalationController);
