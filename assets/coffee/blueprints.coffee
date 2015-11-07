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
  .controller 'BlueprintsCtrl', [ 'Session', '$scope'
    class BlueprintsCtrl
      constructor: (@Session, @$scope) ->
        @characters = []
        @selectedToon = null
        @$scope.$on 'login-status', @_updateLoginStatus
        if @Session.authenticated
          @_updateLoginStatus null, true

      _updateLoginStatus: (_, isLoggedIn) =>
        @authenticated = isLoggedIn
        @getCharacters()
        @selectedToon = @characters[0] if !@selectedToon

      _getBlueprints: (toon) ->
        

      getCharacters: () ->
        @characters = @Session.availableCharacters()

      charactersSelected: () ->
        _getBlueprints @selectedToon
  ]
