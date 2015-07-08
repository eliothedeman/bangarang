angular.module("bangarang", []).controller("RouterController", function($scope){
    
    this.routes = ["home", "policy", "provider", "events"];

    this.route = function(endpoint) {
        console.log(endpoint);
    }

});
