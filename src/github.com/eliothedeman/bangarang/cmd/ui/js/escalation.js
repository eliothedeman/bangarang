function EscalationController($scope, $http, $cookies) {
	this.fetchEscalations = function() {
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
}
angular.module("bangarang").controller("EscalationController", EscalationController);


function NewEscalationController($scope, $http) {
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
		$http.post("api/escalation/config/" + this.name, this.chips);
	}

	this.reset = function() {
		this.type = null;
	}
}


angular.module("bangarang").controller("NewEscalationController", NewEscalationController);
