 'use strict';
 var app = angular.module('app.docs', ['app']);

app.factory('DocFactory', function DocFactory($http, API_URL){
    return {
        getDocs : getDocs
    };

    function getDocs() {
        return $http.get(API_URL + '/api/docs/')
        .then(function success(response) {
            return response;
        });
    }
});
