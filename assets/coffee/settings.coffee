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
  .controller 'SettingsCtrl', [ 'Session', '$scope'
    class SettingsCtrl
      constructor: (@Session, @$scope) ->
        @apikeys = {}
        @newkey = {}
        @forms = {}
        @alerts = []
        @$scope.$on 'login-status', @_updateLoginStatus
        if @Session.authenticated
          @_updateLoginStatus null, true

      _updateLoginStatus: (_, isLoggedIn) =>
        @authenticated = isLoggedIn
        @getApiKeys()
        ((k) -> k.refreshButton = "Refresh")(key) for key in @apikeys

      getApiKeys: () ->
        @apikeys = @Session.apikeys

      deleteKey: (keyID) =>
        @Session.deleteKey keyID

      addKey: () =>
        @Session.addKey @newkey
          .then (response) =>
            @newkey = {}
            # Ignore nonexistent FormController (for tests)
            @forms.newkey?.$setPristine()
          , (_) =>
            @alerts.push
              type: "danger"
              msg: "Internal server error: unable to process key #{@newkey.id}."
            @newkey = {}
            @forms.newkey?.$setPristine()

      refreshKey: (key) ->
        key.processing = true
        key.refreshButton = "Refreshing…"
        @Session.refreshKey key
          .then (_) ->
            key.processing = false
            key.refreshButton = "Refresh"
          , (_) =>
            key.processing = false
            key.refreshButton = "Refresh"
            @alerts.push
              type: "danger"
              msg: "Internal server error: unable to process key #{key.id}."

      closeAlert: (idx) ->
        # Remove the specified alert from the array.
        @alerts.splice(idx, 1)
  ]
