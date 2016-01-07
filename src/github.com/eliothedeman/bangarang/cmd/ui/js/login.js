function LoginController($scope, $http, $cookies) {
	$scope.login = function(user_name, password) {
		$http.get("api/auth/user?user="+user_name+"&pass="+password).then(function(resp){
			// set the auth token
			$cookies.put("BANG_SESSION", resp.data.token);
			window.location = "/dashboard"


		}, function(resp){
			alert(resp.data);
			$cookies.remove("BANG_SESSION");
			window.location.reload();
		});
	}

	$scope.new_user = function(name, user_name, password, confirm) {

		if (password != confirm) {
			alert("Passwords don't match");
			return
		}

		body = {
			name:name,
			user_name: user_name,
			password: password
		}

		$http.post("/api/user", body).then(function(resp){
			$scope.login(user_name, password);
		}, function(resp) {
			alert(resp.data);
		});
	}
}
angular.module('bangarang').controller("LoginController", LoginController);