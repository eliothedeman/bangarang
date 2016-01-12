"use strict"
class Incident {
	constructor(status, metric, policy, tags, time) {
		this.status = status;
		this.tags = new Match([]);
		this.metric = metric;
		this.policy = policy;
		this.time = time
		for (var key in tags) {
			this.tags.add(tags[key].key, tags[key].value);
		}
	}
}
function DashboardController($scope, $cookies, $http, $mdDialog) {
	$scope.incidents = [];
	$scope.fetching = false;
	$scope.stats = {};
	$scope.raw_stats = {};
	$scope.byStats = {};
	var hostMap = {};
	$scope.hostMap = hostMap;
	var stopper = null;

	this.selected = 0;
	this.getSelected = function() {
		var s = $cookies.get("dash:tab");
		if (s) {
			this.selected = s;
		}
		return this.selected;
	}
	this.updateSelected = function(name) {
		$cookies.put("dash:tab", name);
		this.selected = name;
	}

	this.showResolveDialog = function($mdOpen,e) {
		$mdOpen();
	}


	$scope.startFetching = function() {
		$scope.fetchIncidents();
		$scope.fetchStats();

		if (!$scope.fetching) {
			$scope.fetching = true;
			stopper = setInterval(function(){
				$scope.fetchIncidents();
				$scope.fetchStats();
			}, 5000)
		}
	}

	$scope.lastTotal = 0;
	$scope.fetchStats = function() {
		$http.get("api/stats/event").success(function(data) {
			$scope.raw_stats = data;
			$scope.stats["Total Events"] = data.total_events
			$scope.stats["Events/s"] = (data.total_events - $scope.lastTotal) / 5
			$scope.lastTotal = data.total_events
		});

		$http.get("api/stats/system").success(function(data) {
			$scope.stats["Uptime"] = data["uptime"];
		})
	}

	$scope.forgetHost = function(hostname) {
		$http.delete("api/host/"+hostname).success(function() {
			$mdDialog.show($mdDialog.alert().title("Removed host" + hostname).content("").ok("Ok"))
		});
	}

	$scope.stopFetching = function() {
		clearInterval(stopper);
		$scope.fetching = false;
	}


	$scope.resolveIncident = function(id) {
		$http.delete("api/incident/" + id, null).success(function(){
			$scope.fetchIncidents();
		});
	}

	$scope.formatDescription = function(incident) {
		var t = incident.tags;
		return t.get("service") + " on " + t.get("host") + " is " + incident.metric.toFixed(2) + " at " + new Date(incident.time * 1000).format("h:MM:ssTT mmmm-dd-yyyy") + " triggered by " + incident.policy;
	}

	var codes = {
		"0": "OK",
		"1": "WARNING",
		"2": "CRITICAL"
	}

	$scope.getStatusCode = function(status) {
		return codes[status];
	}

	var colors = {
		"0": "green",
		"1": "#FFFD82",
		"2": "#FB5C5C"
	}

	$scope.getStatusColor = function(status) {
		return colors[status]
	}

	$scope.fetchIncidents = function() {
		$http.get("api/incident/*").success(function(data) {
			var ins = [];
			for (var k in data) {
				var i = data[k];
				var x = new Incident(i.status, i.metric, i.policy, i.tags, i.time);
				x.key = k;
				ins.push(x);
			}

			ins.sort(function(x,y) {
				if (x.status != y.status) {
					return y.status - x.status
				}
				return y.time - x.time;
			})
			$scope.incidents = ins;
		});
	}
}

angular.module("bangarang").controller("DashboardController", DashboardController);
