function EscalationController($scope, $http, $cookies, $mdDialog) {
	$scope.escalations = null;
	$scope.fetchEscalations = function() {
		$http.get("api/escalation/config/*").success(function(data, status) {
			$scope.escalations = data;
		});
	}
	this.selected = 0;
	this.getSelected = function() {
		var s = $cookies.get("nec:tab");
		if (s) {
			this.selected = s;
		}
		return this.selected;
	}
	this.updateSelected = function(name) {
		$cookies.put("nec:tab", name);
		this.selected = name;
	}

	$scope.removeSure = {}
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

	$scope.removeEscalation = function(name)  {
		$http.delete("api/escalation/config/"+name).success(function(data) {
			$scope.fetchEscalations();
		});
	}

	$scope.fetchEscalations();

}
angular.module("bangarang").controller("EscalationController", EscalationController);


function NewEscalationController($scope, $http, $interval) {
	this.name = "";
	this.type = null;
	this.ots = {};

	this.type_list = [
		{
			title: "Pagerduty",
			name: "pager_duty"
		},
		{
			title: "Email",
			name: "email",
		},
		{
			title: "Console",
			name: "console"
		}
	]

	this.pdOpts = [
		{
			title:"Api Key",
			name: "key",
			value: ""
		},
		{
			title: "Subdomain",
			name: "subdomain",
			value: ""
		}
	];
	this.emailOpts = [];
	this.consoleOpts = [];
	this.chips = [];

	this.getOpts = function(type) {
		switch (type) {
			case "pager_duty":
				return this.pdOpts;

			case "email":
				return this.emailOpts;

			case "console":
				return this.consoleOpts;

			default:
				return [];
		}
	}



	this.newEscalation = function() {

		if (!this.type) {
			return;
		}

		var e = {
			type: this.type
		};

		var opts = this.getOpts(this.type);
		for (var i = 0; i < opts.length; i++) {
			e[opts[i].name] = opts[i].value;
		}

		this.chips.push(e);
	}

	this.submitNew = function() {
		if (!this.name) {
			return;
		}
		$scope.newEscalationProgress = 50;
		a = this;
		$http.post("api/escalation/config/" + this.name, this.chips).success(function(data) {
			a.reset();
		});

	}

	this.reset = function() {
		this.type = null;
		this.name = "";
		this.chips = [];
		this.opts = {};
		$scope.newEscalationProgress = 0;
	}
}


angular.module("bangarang").controller("NewEscalationController", NewEscalationController);
