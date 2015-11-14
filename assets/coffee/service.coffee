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

# Constants
jitaID = 60003760 # Jita IV - Moon 4 - Caldari Navy Assembly Plant

angular.module 'eveindy'
  .service 'Server', [ '$http',
    class ServerService
      constructor: (@$http) ->

      getAutocomplete: (searchTerm) ->
        @$http.get "/autocomplete/station/" + encodeURIComponent searchTerm

      parsePastebin: (pasteText) ->
        @$http.post "/pastebin", {"paste": pasteText }

      _queryFromItems: (items) ->
        for item in items
          itemName: item.Item.Name
          quantity: item.Quantity

      _queryFromItemList: (itemNames) ->
        for item in itemNames
          itemName: item
          quantity: 1

      getJitaPrices: (itemNames) ->
        @getLocalPrices jitaID, itemNames

      # Special-purpose endpoint just for the minerals summary.
      getReprocessPrices: () ->
        @$http.get "/market/jita"

      getLocalPrices: (stationID, itemNames) ->
        do (query = @_queryFromItemList itemNames) =>
          @$http.post "/market/station/" + encodeURIComponent(stationID), query

      searchSystemMarket: (systemName, q) ->
        do (query = @_queryFromItems q) =>
          @$http.post "/market/system/" + encodeURIComponent(systemName), query

      searchStationMarket: (stationID, q) ->
        do (query = @_queryFromItems q) =>
          @$http.post "/market/station/" + encodeURIComponent(stationID), query

      reprocessItems: (stationYield, stationTax, scrapSkill, items) ->
        query =
          stationYield: stationYield
          taxRate: stationTax
          scrapmetalReprocessingSkill: scrapSkill
        # Can't put this in do binding - bug in compiler? PEBCAK?
        query.items = @_queryFromItems items
        @$http.post "/reprocess", query

      getLoginStatus: () ->
        @$http.get "/session"

      apiForUser: () ->
        @$http.get "/apikeys/list"

      deleteApiKey: (keyID) ->
        @$http.post "/apikeys/delete/#{keyID}"

      addApiKey: (key) ->
        @$http.post "/apikeys/add", key

      refreshApiKey: (key) ->
        @$http.post "/apikeys/refresh", key

      getEffectiveStandings: (characterID, stationID) ->
        @$http.get "/standings/#{characterID}/#{stationID}"

      getSkills: (characterID, skillGroupID) ->
        @$http.get "/skills/#{characterID}/group/#{skillGroupID}"

      logout: () ->
        @$http.post "/logout"

      # Blueprints!
      getBlueprints: (characterID) ->
        @$http.get "/blueprints/#{characterID}"
      ]
