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
  .controller 'BlueprintsCtrl', [ 'Session', '$scope', 'DTOptionsBuilder'
    class BlueprintsCtrl
      constructor: (@Session, @$scope, @DTOptionsBuilder) ->
        @characters = []
        @blueprints = []
        @selectedToon = null
        @dtOptions = @DTOptionsBuilder.newOptions()
        @dtColumnDefs = [
          type: 'numeric-inf'
          targets: 2
        ]

        @$scope.$on 'login-status', @_updateLoginStatus
        if @Session.authenticated
          @_updateLoginStatus null, true

      _updateLoginStatus: (_, isLoggedIn) =>
        @authenticated = isLoggedIn
        @getCharacters()
        if !@selectedToon
          @selectedToon = @characters[0]
          @characterSelected()

      _getBlueprints: (toon) ->
        @Session.blueprints toon
          .then (resp) =>
            for bp in resp.blueprints
              bp.NumRuns = '∞' if bp.IsOriginal
              bp.Station = resp.stations[bp.StationID]
            @blueprints = resp.blueprints

      getCharacters: () ->
        @characters = @Session.availableCharacters()

      characterSelected: () ->
        @_getBlueprints @selectedToon
  ]
