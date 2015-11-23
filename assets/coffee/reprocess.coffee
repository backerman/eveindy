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

# Constants for this file.
reprocessOutput = [
  "Tritanium"
  "Pyerite"
  "Mexallon"
  "Isogen"
  "Nocxium"
  "Megacyte"
  "Zydrine"
]

# Price calculation functions

buyCalc = (item, mult) ->
  item?.bestBuy * mult

sellCalc = (item, mult) ->
  item?.bestSell * mult

midpointCalc = (item, mult) ->
  switch
    when item?.bestBuy and item?.bestSell
      Math.round((item?.bestBuy + item?.bestSell)/2 * mult * 100) / 100
    when i?.bestBuy
      Math.round(item.bestBuy * mult * 100) / 100
    when i?.bestSell
      Math.round(item.bestSell * mult * 100) / 100
    else undefined

angular.module 'eveindy'
  .controller 'ReprocessCtrl', [
    'Server', 'Session', '$scope', 'numberFilter', '$timeout', '$sce',
    class ReprocessCtrl
      constructor: (
              @Server, @Session, @$scope, @numberFilter, @$timeout, @$sce) ->
        @standing = 0.0
        @taxRate = 0.0
        @scrapSkill = 0
        @$scope.$on 'login-status', @_updateLoginStatus
        @_updateLoginStatus()
        @priceCalc = "midpoint"
        @priceCalcFn = midpointCalc
        @mineralValue = "imputed"
        @reproMultiplier = 1.00
        @imputed = {}
        # We have the separate array because it makes things easier in the view.
        @imputedArray = []
        @Server.getReprocessPrices()
          .then (response) =>
            @jitaPrices = response.data
            @recalculateMineralPrices()

      _updateLoginStatus: () =>
        @authenticated = @Session.authenticated
        @chars = @Session.availableCharacters()

      getStations: (search) ->
        @Server.getAutocomplete(search)
          .then (response) =>
            ((i) =>
              i.class="security-" + switch
                when i.security >= 0.5 then "high"
                when i.security > 0.0 then "low"
                else "null"
              # Server uses 0..1; we use 0..100.
              # FIXME: Converted both here (autocomplete) and on server
              #   (reprocess). Pick one.
              i.name = @$sce.trustAsHtml i.name
              i.constellation = @$sce.trustAsHtml i.constellation
              i.region = @$sce.trustAsHtml i.region
              i.reprocessingEfficiency *= 100
              i
            )(item) for item in response.data

      locationSelected: (loc) ->
        @corporationName = loc.owner
        @stationID = loc.id
        @stationOwnerID = loc.ownerID
        @stationType = if loc.isOutpost then "player" else "npc"
        @reprocessingEfficiency = loc.reprocessingEfficiency
        if @stationType is "npc"
          @selectedToon = @chars[0]
          @characterSelected()

      # Character selection has changed.
      characterSelected: () ->
        id = @selectedToon.id
        @scrapSkill = @selectedToon.skills.filter( (s) ->
          s.name is "Scrapmetal Processing"
          )?[0]?.level || 0
        if @stationType is "npc"
          @Server.getEffectiveStandings id, @stationOwnerID
            .then (response) =>
              @standing = response.data.standing
              @updateTaxRate()

      # Update mineral value calculation method.
      updatePriceCalc: () ->
        @priceCalcFn = switch @priceCalc
          when 'midpoint' then midpointCalc
          when 'buy' then buyCalc
          when 'sell' then sellCalc
          else midpointCalc # default if bad data
        @recalculateMineralPrices()

      # Do the recalculation of imputed reprocessing value.
      recalculateMineralPrices: () ->
        @imputed = {}
        @imputedArray = []
        for name, item of @jitaPrices
          newPrice = @priceCalcFn item, @reproMultiplier
          @imputed[name] = newPrice
          @imputedArray.push {name: name, price: newPrice}
        # Update reprocessing values if we have those results yet.
        if @reprocessingOutput
          @updateReprocessedValues(@reprocessingOutput)

      submitPaste: ->
        @Server.parsePastebin(@pastebin)
          .then (response) =>
            # FIXME: Transition service in ui-bootstrap triggering $digest
            #   error.
            @hidePasteInput = true
            @$timeout (() => @inventory = response.data), 0, false
            # flag to display spinner
            @searchingMarket = true
            # Get market prices
            @Server.searchStationMarket(@stationID, response.data)
              .then (response) =>
                @searchingMarket = false
                # Update items with the best buy order price
                do (parse = (i) =>
                  result = response.data[i.Item.Name]
                  if result and result.bestBuy > 0
                    bi = result.buyInfo
                    i.buyPrice = result.bestBuy * i.Quantity
                    i.buyInfo = @$sce.trustAsHtml "<table class=\"orderinfo\">
                        <tr>
                          <th>\# units:</th>
                          <td>#{bi.quantity}</td>
                        </tr>
                        <tr>
                          <th>Min. qty.:</th>
                          <td>#{bi.minQuantity}</td>
                        </tr>
                        <tr>
                          <th>Price ea.:</th>
                          <td>#{@numberFilter(result.bestBuy, 2)} ISK</td>
                        </tr>
                        <tr>
                          <th>
                            Location:</th>
                          <td>
                            #{bi.station.name} (#{bi.within})</td>
                        </tr>
                       </table>"
                  else
                    i.noSale = true
                    i.buyPrice = 0.0) =>
                  parse item for item in @inventory
            # Get reprocessing value
            @Server.reprocessItems \
              @reprocessingEfficiency, @taxRate, @scrapSkill, response.data
              .then (response) =>
                @reprocessingOutput = response.data
                @updateReprocessedValues(@reprocessingOutput)

      updateReprocessedValues: (response) ->
        for item in @inventory
          result = response.items[item.Item.Name]
          item.reproValue = 0
          if result?.length > 0
            for m in result
              imputedCost = @imputed[m.Item.Name]
              if !imputedCost
                imputedCost = \
                  @priceCalcFn response.prices[m.Item.Name], @reproMultiplier
              item.reproValue += m.Quantity * imputedCost
          else
            item.noReprocess = true

      formatPrice: (item) ->
        switch
          when item.noSale then "-"
          when item.buyPrice > 0 then @numberFilter(item.buyPrice, 2)
          else "-"

      formatRepro: (item) ->
        switch
          when item.noReprocess then "-"
          when item.reproValue > 0 then @numberFilter(item.reproValue, 2)
          else "-"

      # updateTaxRate sets the tax rate based on the user's standing with
      # a station's NPC corporation owners.
      updateTaxRate: ->
        @taxRate = Math.max(0.0, 5.0-0.75*@standing)
      ]
