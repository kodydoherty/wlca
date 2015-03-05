!function(){"use strict";var t=angular.module("app",["app.auth","ui.router","flow","app.docs"]);t.constant("API_URL","http://localhost:3000"),t.config(function(t,o,e){t.interceptors.push("AuthInterceptor"),e.otherwise("/"),o.state("home",{url:"/",templateUrl:"Home/home.html",controller:"MainCtrl",data:{auth:!1}}).state("docs",{url:"/docs",templateUrl:"Docs/docs.html",controller:"DocCtrl",data:{auth:!0}})}),t.controller("MainCtrl",["$scope","$rootScope","UserFactory","$state",function o(t,e,n,r){function a(t,o){n.login(t,o).then(function a(t){e.user=t.data.user,r.go("docs")},u)}function c(){n.logout(),e.user=null}function u(t){e.flash=t.data,r.go("home")}t.login=a,t.logout=c}]),t.controller("DocCtrl",["$scope","$rootScope","$state","UserFactory","DocFactory",function(t,o,e,n,r){function a(){n.logout(),o.username=null,e.go("home")}function c(o){t.currentCategory=o}function u(o){return null!==t.currentCategory&&o.name===t.currentCategory.name}function l(t){o.flash=t.data,e.go("home")}t.logout=a,t.categories=[{id:0,name:"Budget"},{id:1,name:"NewsLetter"},{id:2,name:"Meeting Minutes"},{id:3,name:"Welcome To Walden"}],n.getUser().then(function s(t){o.user=t.data},l),r.getDocs().then(function i(o){t.docs=o.data}),t.currentCategory=null,t.setCurrentCategory=c,t.isCurrentCategory=u}])}();