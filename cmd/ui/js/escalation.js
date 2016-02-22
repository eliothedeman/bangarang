"use strict"
class Escalation {
	constructor(name) {
		this.name = name;
		this.comment = "";
		this.match = new Match([]);
		this.not_match = new Match([]);
		this.crit = false;
		this.warn = false;
		this.ok = false;
		this.configs = [];
		this.type_list = [
			{
				title: "Pagerduty",
				name: "pager_duty",
				opts: [
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
				]
			},
			{
				title: "Email",
				name: "email",
				opts: [
					{
						title: "To",
						name: "recipients",
						value: "",
						format: function() {
							if (typeof this.value == "string") {
								this.value = this.value.split(",");
							}
						}
					},
					{
						title: "From",
						name:"sender",
						value:""
					},
					{
						title: "User",
						name:"user",
						value:""
					},
					{
						title: "Password",
						name:"password",
						value:""
					},
					{
						title:"Host",
						name:"host",
						value: "smtp.gmail.com"
					},
					{
						title:"Port",
						name:"port",
						value: 465
					}
				]
			},
			{
				title: "Console",
				name: "console",
				opts: []
			},
			{
				title: "Grafana Graphite Annotation",
				name: "grafana_graphite_annotation",
				opts: [
					{
						title:"Host",
						name: "host",
						value: ""
					},
					{
						title:"Port",
						name: "port",
						value: 2003
					}	
				]
			}
		]
	}

	getOpts(name) {
		for (var i = this.type_list.length - 1; i >= 0; i--) {
			var opt = this.type_list[i];
			if (opt.name == name) {
				return opt.opts
			}

		};

		return {}
	}

	data() {
		var d = {
			name: this.name,
			comment: this.comment,
			match: this.match.data(),
			not_match: this.not_match.data(),
			ok: this.ok,
			warn: this.warn,
			crit: this.crit,
			configs: this.configs,
		}

		return d
	}

	addEscalation(conf) {
		this.configs.push(conf);
	}
}

function parseEscalation(d) {
	var e = new Escalation(d.name);
	e.match = new Match(d.match);
	e.not_match = new Match(d.not_match);
	e.ok = d.ok;
	e.warn = d.warn;
	e.crit = d.crit;
	e.configs = d.configs;

	return e
}

function EscalationController($scope, $http, $cookies, $mdDialog) {
	$scope.escalations = [];
	$scope.fetchEscalations = function() {
		$http.get("api/escalation/config/*").success(function(data, status) {
			if (data) {
				if (data != "null") {
					$scope.escalations = data;
				}
			}
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
	$scope.esc = new Escalation("");
	$scope.tmp_type = "";
	$scope.escalations = [];

	$scope.updateCurrent = function(name) {
		for (var i = $scope.escalations.length - 1; i >= 0; i--) {
			if($scope.escalations[i].name == name) {
				$scope.esc = $scope.escalations[i];
			}
		};
	}

	$scope.fetchEscalations = function() {
		$http.get("api/escalation/config/*").then(function(resp){
			var data = resp.data;
			if (typeof(data) == "object") {
				$scope.escalations = [];
				for (var key in data) {
					$scope.escalations.push(parseEscalation(data[key]));
				}
			}

		}, function(resp){
			console.log(resp);
		})

	}

	$scope.submitUpdated = function() {
		console.log("update");
		if (!$scope.validate()) {
			alert("Escalation is not complete")
			return
		}

		$http.delete("api/escalation/config/" + $scope.esc.name).then(function(data){
			$scope.submitNew();
		}, function(resp) {
			alert(resp.data);
		});
	}

	$scope.submitNew = function() {
		if (!$scope.validate()) {
			alert("Escalation is not complete")
			return
		}

		$http.post("api/escalation/config/" + $scope.esc.name, $scope.esc.data()).success(function(data) {
			$scope.reset()
		});
	}

	$scope.appendEscalation = function() {
		var d = {};
		var opts = $scope.esc.getOpts($scope.tmp_type);
		for (var i = opts.length - 1; i >= 0; i--) {
			var opt = opts[i];
			d[opt.name] = opt.value;

		};

		d.type = $scope.tmp_type;
		$scope.esc.configs.push(d);
	}

	$scope.validate = function() {
		if (!$scope.esc.name) {
			return false;
		}
		return true;
	}

	$scope.reset = function() {
		$scope.esc = new Escalation("");
	}
}


angular.module("bangarang").controller("NewEscalationController", NewEscalationController);
