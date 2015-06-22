function Client(host, port) {
	this.host = host;
	this.port = port;

	this.baseUrl = function() {
		return this.host + ":" + this.port
	}

	this.sync = function(url, data) {
		result = {};
		$.ajax(url, {
			async: false,
			dataType: "json",
			data: data,
			success: function(data) {
				result = data;
			}
		});
		return result;
	}

	this.getAllStats = function() {
		return this.sync(this.baseUrl() + "/api/stats/event", {});
	}

	this.getHosts = function() {
		return this.sync(this.baseUrl() + "/api/stats/hosts", {});
	}

	this.getServices = function(f) {
		return this.sync(this.baseUrl() + "/api/stats/services", {});
	}

}