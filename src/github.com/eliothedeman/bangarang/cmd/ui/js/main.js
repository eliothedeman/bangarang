function Router($scope, $cookies) {
	this.selected = 0;
	this.getSelected = function() {
		var s = $cookies.get("router:tab");
		if (s) {
			this.selected = s;
		}

		return this.selected;
	}

	this.updateSelected = function(index) {
		$cookies.put("router:tab", index); 
		this.selected = index;
	}
}

angular.module("bangarang").controller("Router", Router);

function Config($scope, $cookies, $http, $mdDialog) {
	$scope.snapshots = [];
	$scope.snapshotsByHash = {};
	$scope.fetchSnapshots = function() {
		$http.get("api/config/version/*").success(function(data) {
			data.sort(function(x,y){
				return new Date(x.time_stamp) - new Date(y.time_stamp)

			});
			data.reverse()
			$scope.snapshotsByHash = {};
			for (var i = data.length - 1; i >= 0; i--) {
				$scope.snapshotsByHash[data[i].hash] = data[i]
			};
			$scope.snapshots = data;
		});
	}
	$scope.showInspect = function(hash) {
		console.log(hash);
		var inspect = $mdDialog.confirm()
			.title(hash)
			.content($scope.snapshotsByHash[hash].app)
			.ariaLabel("Config inspection")
			.cancel("Cancel")
			.ok("Set as current")

		$mdDialog.show(inspect).then(function(){
			$scope.updateCurrent(hash);
		}, function() {
			
		})
	}

	$scope.updateCurrent = function(hash) {
		$http.post("api/config/version/" + hash).success(function(data) {
			$scope.fetchSnapshots();
		})
	}
	this.selected = 0;
	this.getSelected = function() {
		var s = $cookies.get("conf:tab");
		if (s) {
			this.selected = s;
		}
		return this.selected;
	}
	this.updateSelected = function(name) {
		$cookies.put("conf:tab", name);
		this.selected = name;
	}
}

angular.module("bangarang").controller("Config", Config)