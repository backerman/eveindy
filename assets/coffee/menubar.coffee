# Copyright © 2014–5 Brad Ackerman.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

angular.module 'eveindy'
  .controller 'MenubarCtrl', [ '$scope', 'Server', '$route', '$window'
    class MenubarCtrl
      constructor: (@$scope, @Server, @$route, @$window) ->
        @$scope.$on '$routeChangeSuccess', @updateMenubar
        @$scope.$on '$routeChangeStart', @preventNullRoute
        @$scope.view = "reprocess"
        @Server.getLoginStatus()
          .then (response) =>
            @authenticated = response.data.authenticated
        @$route.reload()

        # Put function on window to be called by authentication success screen.
        @$window.hasAuthenticated = () =>
          @$scope.$apply () =>
            @authenticated = true

      updateMenubar: (_, thisRoute, prevRoute) =>
        # Set menu bar active element to the current page.
        switch thisRoute.templateUrl
          when "view/reprocess.html"
            @$scope.view = "reprocess"
          when "view/blueprints.html"
            @$scope.view = "blueprints"
          else
            @$scope.view = "settings"

      preventNullRoute: (evt, next, current) ->
        if !next?.templateUrl?
          # Undefined route - block change.
          evt.preventDefault()

      logout: () =>
        @Server.logoutSessions()
          .then (response) =>
            # FIXME: Should display error if error returned from server,
            # even though that's unlikely and we really can't do anything.
            @authenticated = false
      ]
