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

# Skill group for refining skills.
RESOURCE_PROCESSING = 1218

# The Session service is a single repository for storing the user's session
# information.
angular.module 'eveindy'
  .service 'Session', [ 'Server', '$rootScope', '$window'
    class SessionService
      constructor: (@Server, @$rootScope, @$window) ->
        @apikeys = []
        @authenticated = false
        @_getSessionStatus()

        # Put function on window to be called by authentication success screen.
        @$window.hasAuthenticated = () =>
          @$rootScope.$apply () =>
            @_getSessionStatus()

      logout: () ->
        @Server.logout()
          .then (response) =>
            @authenticated = false
            @apikeys = []
            @$rootScope.$broadcast('login-status', @authenticated)

      # Get session status (authenticated, API keys)
      _getSessionStatus: () ->
        @Server.getLoginStatus()
          .then (response) =>
            @authenticated = response.data.authenticated
            if @authenticated
              @apikeys = response.data.apiKeys
              @_getSkills()
            @_keysUpdated()

      # Get skills for the current characters.
      _getSkills: () ->
        for c in @availableCharacters()
          do (c) =>
            @Server.getSkills(c.id, RESOURCE_PROCESSING)
              .then (response) ->
                c.skills = response.data

      # Broadcast API key change notification.
      _keysUpdated: () ->
        @$rootScope.$broadcast('login-status', @authenticated)

      deleteKey: (keyID) =>
        @Server.deleteApiKey keyID
          .then (response) =>
            # We don't actually care about the response; just drop the key
            # from our model. FIXME: we should care.
            @apikeys = @apikeys.filter (key) ->
              key.id != keyID
            @_keysUpdated()

      addKey: (newKey) =>
        @Server.addApiKey newKey
          .then (response) =>
            newKey.characters = response.data.characters
            @apikeys.push newKey
            @_keysUpdated()

      refreshKey: (key) =>
        @Server.refreshApiKey key
          .then (response) =>
            key.characters = response.data.characters
            @_getSkills()
            @_keysUpdated()

      # List the characters available for this user.
      availableCharacters: () ->
        reduceFn = (chars, key) ->
          if key.characters then [chars..., key.characters...] else chars
        @apikeys.reduce reduceFn, []

      standing: (character, corporation) ->
        42
  ]
