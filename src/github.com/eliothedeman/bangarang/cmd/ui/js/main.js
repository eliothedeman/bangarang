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
