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
  .controller 'SettingsCtrl', [ 'Server', '$scope'
    class SettingsCtrl
      constructor: (@Server, @$scope) ->
        @apikeys = {}
        @newkey = {}
        @forms = {}
        @$scope.$on 'login-status', @_updateLoginStatus
        if @Server.authenticated
          @authenticated = true
          @getApiKeys()

      _updateLoginStatus: (_, isLoggedIn) =>
        if isLoggedIn
          @authenticated = true
          @getApiKeys()
        else
          # Logged out - clear keys
          @authenticated = false
          @apikeys = []

      getApiKeys: () ->
        @Server.apiForUser()
          .then (response) =>
            @apikeys = response.data

      deleteKey: (keyID) =>
        @Server.deleteApiKey keyID
          .then (response) =>
            # We don't actually care about the response; just drop the key
            # from our model.
            @apikeys = @apikeys.filter (key) ->
              key.id != keyID

      addKey: () =>
        @Server.addApiKey @newkey
          .then (response) =>
            @apikeys.push @newkey
            @newkey = {}
            # Ignore nonexistent FormController (for tests)
            @forms.newkey?.$setPristine()
  ]
