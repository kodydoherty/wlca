 'use strict';
 var app = angular.module('app.auth', ['app']);

 app.factory('UserFactory', ['$http', 'API_URL','AuthTokenFactory','$q', function ($http, API_URL, AuthTokenFactory, $q) {

    function login(username, password) {
        return $http.post(API_URL + '/login', {
            username: username,
            password: password
        }).then(function success(response) {
            AuthTokenFactory.setToken(response.data.Token);
            return response;
        });
    }
    function logout() {
        AuthTokenFactory.setToken();
    }

    function getUser() {
        if (AuthTokenFactory.getToken()) {
            return $http.get(API_URL + '/api/me');
        } else {
            return $q.reject({ data: 'client has no auth token'});
        }
    }

    return {
        login: login,
        logout: logout ,
        getUser: getUser

    };
}]);

 app.factory('AuthTokenFactory', ['$window', function ($window) {

    var store = $window.localStorage;
    var key = 'auth-token';
    function getToken() {
        return store.getItem(key);
    }
    function setToken(token) {
        if (token) {
            store.setItem(key, token);
        } else {
            store.removeItem(key);
        }
    }
    return {
        getToken: getToken,
        setToken: setToken
    };
}]);

 app.factory('AuthInterceptor', ['AuthTokenFactory', function (AuthTokenFactory) {

    function addToken(config) {
        var token = AuthTokenFactory.getToken();
        if (token) {
            config.headers = config.headers || {};
            config.headers.Authorization = 'Bearer ' + token;
        }
        return config;
    }

    return {
        request: addToken
    };

}]);