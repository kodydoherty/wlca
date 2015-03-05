(function() {
    'use strict';
    var app = angular.module('app', ['app.auth', 'ui.router','flow', 'app.docs']);
    app.constant('API_URL', 'http://localhost:3000');
    app.config(function ($httpProvider, $stateProvider, $urlRouterProvider) {
        $httpProvider.interceptors.push('AuthInterceptor');

        $urlRouterProvider.otherwise('/');
        $stateProvider
        .state('home', {
            url: '/',
            templateUrl: 'Home/home.html',
            controller: 'MainCtrl',
            data: {
                auth: false
            }
        })
        .state('docs', {
            url: '/docs',
            templateUrl: 'Docs/docs.html',
            controller: 'DocCtrl',
            data: {
                auth: true
            }
        });
    });

  //   app.run(['$rootScope','$state', function ($rootScope, $state){
  //     $rootScope.$on('$stateChangeStart', function (event, toState, toParams, fromState, fromParams) {
  //       // if route requires auth and user is not logged in
  //       if (toState.data.auth) {
  //           if ($rootScope.user) {
  //               $state.go('home');
  //           }
  //       }
  //   });
  // }]);



    app.controller('MainCtrl', ['$scope', '$rootScope', 'UserFactory','$state', function MainCtrl($scope , $rootScope, UserFactory, $state) {

        $scope.login = login;
        $scope.logout = logout;
        $scope.register = register;
        function register(username, password) {
            UserFactory.register(username, password).then(function success(response) {
                $rootScope.user = response.data.user;
                alert("Welcome!")
                $state.go('docs');
            }, handleError);
        }

        function login(username, password) {
            UserFactory.login(username, password).then(function success(response) {
                $rootScope.user = response.data.user;
                $state.go('docs');
            }, handleError);
        }

        function logout() {
            UserFactory.logout();
            $rootScope.user = null;
        }

        function handleError(response) {
            $rootScope.flash = response.data;
            $state.go('home');
        }
    }]);

    app.controller('DocCtrl', ['$scope','$rootScope','$state','UserFactory' ,
        'DocFactory', function ($scope, $rootScope, $state, UserFactory, DocFactory) {
        $scope.logout = logout;

        $scope.categories = [
            {'id': 0, 'name': 'Budget'},
            {'id': 1, 'name': 'NewsLetter'},
            {'id': 2, 'name': 'Meeting Minutes'},
            {'id': 3, 'name': 'Welcome To Walden'}
        ];

        UserFactory.getUser().then(function success(response) {
            $rootScope.user = response.data;
        }, handleError);

        DocFactory.getDocs().then(function success(response) {
            $scope.docs = response.data;

        });

        function logout() {
            UserFactory.logout();
            $rootScope.username = null;
            $state.go('home');
        }

        $scope.currentCategory = null;
        function setCurrentCategory(category) {
            $scope.currentCategory = category;
        }

        function isCurrentCategory(category) {
            return $scope.currentCategory !== null && category.name === $scope.currentCategory.name;
        }

        $scope.setCurrentCategory = setCurrentCategory;
        $scope.isCurrentCategory = isCurrentCategory;

        function handleError(response) {
            $rootScope.flash = response.data;
            $state.go('home');
        }
    }]);

})();