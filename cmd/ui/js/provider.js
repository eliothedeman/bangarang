function ProviderController($scope, $http, $cookies) {
	$scope.providers = {};
	this.fetchProviders = function() {
		$http.get("/api/provider/config/*").success(function(data,status) {
			for (name in data) {
				$scope.providers[name] = data[name];
			}
		});
	}
	this.selected = 0;
	this.getSelected = function() {
		var s = $cookies.get("prov:tab");
		if (s) {
			this.selected = s;
		}
		return this.selected;
	}
	this.updateSelected = function(name) {
		$cookies.put("prov:tab", name);
		this.selected = name;
	}
}
angular.module("bangarang").controller("ProviderController", ProviderController);
function NewProvider($scope, $http) {

	this.getOpts = function(name) {
		for (var i = this.type_list.length - 1; i >= 0; i--) {
			if (this.type_list[i].name == name || this.type_list[i].title == name)  {
				return this.type_list[i].opts
			}
		};

		return []
	}

	this.type_list = [
		{
			title: "TCP",
			name: "tcp",
			opts: [
				{
					title: "Listen",
					name: "listen"
				},
				{
					title: "Encoding",
					name: "encoding"
				}
			]
		},
		{
			title: "HTTP",
			name: "HTTP",
			opts: [
				{
					title: "Listen",
					name: "listen"
				},
				{
					tittle: "Encoding",
					name: "encoding"
				}
			]
		}
	]
	var t = this

	this.submitNew = function() {
		console.log(this)

		this.opts.type = this.type;
		$http.post("api/provider/config/" + this.name, this.opts).then(function(){
			t.init();
			alert("Success")
		}, function(resp) {
			alert(resp.data);
		})
	}

	this.init = function() {
		this.name = "";
		this.type = "";
		this.opts = {};
	}

	this.init();
}

angular.module('bangarang').controller('NewProvider', NewProvider);