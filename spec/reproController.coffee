# Copyright © 2014–5 Brad Ackerman.
#
# Licensed under the Apache License, Version 2.0  the "License";
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

describe 'Controller: ReproController', () ->
  beforeEach () -> module 'eveindy'
  ctrl = undefined
  serverService = undefined
  jitaMinerals = undefined
  pastebinTest = undefined
  scope = undefined

  beforeEach () ->
    # $timeout is used in the code under test only because of a workaround
    # for angular-ui/bootstrap until it's actually 1.3-compatible. So we
    # just replace it with a mock that executes the callback immediately.
    module ($provide) ->
      $provide.constant '$timeout', (fn) ->
        fn()

    inject ($controller, Server, _$rootScope_, $q) ->

      # Hook a call on our Server to return static information.
      fakeReturn = (serverMethod, jsonReturned) ->
        spyOn Server, serverMethod
          .and.callFake () ->
            deferred = $q.defer()
            deferred.resolve
              data: JSON.parse JSON.stringify jsonReturned
            deferred.promise

      scope = _$rootScope_.$new()

      jitaMinerals = fixture.load('jitaMinerals.json')
      autocomplete = fixture.load('autocomplete.json')
      pastebinTest = fixture.load('pastebin.json')
      marketOutput = fixture.load('marketOutput.json')
      reprocessOutput = fixture.load('reprocessOutput.json')
      sessionInfo = fixture.load('session.json')

      fakeReturn 'getReprocessPrices', jitaMinerals
      fakeReturn 'getAutocomplete', autocomplete
      fakeReturn 'parsePastebin', pastebinTest
      fakeReturn 'searchStationMarket', marketOutput
      fakeReturn 'reprocessItems', reprocessOutput
      fakeReturn 'getLoginStatus', sessionInfo

      # getSkills and getEffectiveStandings will be called indirectly.
      spyOn Server, 'getSkills'
        .and.callFake (charID, skillGroup) ->
          deferred = $q.defer()
          deferred.resolve
            data: []
          deferred.promise

      spyOn Server, 'getEffectiveStandings'
        .and.callFake () ->
          deferred = $q.defer()
          deferred.resolve
            data:
              standing: 0.00
          deferred.promise

      serverService = Server
      ctrl = $controller 'ReprocessCtrl',
        $scope: scope

  it 'should get Jita mineral prices', () ->
    scope.$apply()
    expect(serverService.getReprocessPrices).toHaveBeenCalled()
    expect(serverService.getReprocessPrices.calls.count()).toEqual 1

  it 'should populate the correct arrays', () ->
    scope.$apply()
    expect(ctrl.jitaPrices).toEqual(jitaMinerals)

  it 'should correctly calculate midpoint values', () ->
    scope.$apply()
    midpoints =
      Isogen: 137.93
      Megacyte: 717.00
      Mexallon: 59.89
      Nocxium: 665.57
      Pyerite: 12.22
      Tritanium: 5.86
      Zydrine: 407.11
    expect(ctrl.imputed).toEqual(midpoints)

  it 'should correctly calculate buy values', () ->
    buy =
      Isogen: 136.33
      Megacyte: 704.03
      Mexallon: 59.36
      Nocxium: 642.17
      Pyerite: 12.18
      Tritanium: 5.72
      Zydrine: 400.01
    ctrl.priceCalc = 'buy'
    ctrl.updatePriceCalc()
    scope.$apply()
    expect(ctrl.imputed).toEqual(buy)

  it 'should correctly calculate sell values', () ->
    sell =
      Isogen: 139.53
      Megacyte: 729.96
      Mexallon: 60.41
      Nocxium: 688.97
      Pyerite: 12.25
      Tritanium: 5.99
      Zydrine: 414.20
    ctrl.priceCalc = 'sell'
    ctrl.updatePriceCalc()
    scope.$apply()
    expect(ctrl.imputed).toEqual(sell)

  it 'should correctly calculate midpoint values with a multiplier', () ->
    ctrl.reproMultiplier = 10.0
    ctrl.updatePriceCalc()
    midpoints =
      Isogen: 1379.30
      Megacyte: 7169.95
      Mexallon: 598.85
      Nocxium: 6655.70
      Pyerite: 122.15
      Tritanium: 58.55
      Zydrine: 4071.05
    scope.$apply()
    expect(ctrl.imputed).toEqual(midpoints)

  it 'should process autocomplete results', () ->
    testSystem = 'Mir'
    response = []
    ctrl.getStations(testSystem)
      .then (result) ->
        response = result
    scope.$apply()
    expect(response[0].class).toEqual 'security-null'
    expect(response[0].reprocessingEfficiency).toEqual 50
    expect(response[1].class).toEqual 'security-low'
    expect(response[1].reprocessingEfficiency).toEqual 50
    expect(response[7].class).toEqual 'security-high'
    expect(response[7].reprocessingEfficiency).toEqual 30

  it 'should parse station information', () ->
    scope.$apply()
    testStation =
      constellation: "Ambrye"
      id: 60002479
      isOutpost: false
      name: "Mirilene VII - Moon 3 - Expert Distribution Warehouse"
      owner: "Expert Distribution"
      region: "Sinq Laison"
      reprocessingEfficiency: 50
      security: 0.779691
      systemName: "Mirilene"
      class: "security-high"

    ctrl.locationSelected(testStation)
    scope.$apply()
    expect(ctrl.corporationName).toEqual 'Expert Distribution'
    expect(ctrl.stationID).toEqual 60002479
    expect(ctrl.stationType).toEqual 'npc'
    expect(ctrl.reprocessingEfficiency).toEqual 50

  it 'should calculate NPC tax rate based on standings', () ->
    testStation =
      constellation: "Ambrye"
      id: 60002479
      isOutpost: false
      name: "Mirilene VII - Moon 3 - Expert Distribution Warehouse"
      owner: "Expert Distribution"
      region: "Sinq Laison"
      reprocessingEfficiency: 50
      security: 0.779691
      systemName: "Mirilene"
      class: "security-high"
    scope.$apply()
    ctrl.locationSelected(testStation)
    scope.$apply()
    ctrl.standing = 0.00
    ctrl.updateTaxRate()
    scope.$apply()
    expect(ctrl.taxRate).toEqual 5.00

    ctrl.standing = 3.00
    ctrl.updateTaxRate()
    expect(ctrl.taxRate).toEqual 2.75

    ctrl.standing = 8.00
    ctrl.updateTaxRate()
    expect(ctrl.taxRate).toEqual 0.00

  it 'should parse inventory', () ->
    testStation =
      constellation: "Ambrye"
      id: 60011854
      isOutpost: false
      name: "Mirilene IV - Moon 1 - Federation Navy Logistic Support"
      owner: "Expert Distribution"
      region: "Sinq Laison"
      reprocessingEfficiency: 50
      security: 0.779691
      systemName: "Mirilene"
      class: "security-high"
    ctrl.pastebin = "This is ignored so do whatever."
    scope.$apply()
    ctrl.locationSelected(testStation)
    ctrl.submitPaste()
    scope.$apply()
    # Create copy of inventory with only the bits we care about.
    myInventory = []
    for item in ctrl.inventory
      newItem =
        Quantity: item.Quantity
        Item:
          BatchSize: item.Item.BatchSize
          Category: item.Item.Category
          Group: item.Item.Group
          ID: item.Item.ID
          Name: item.Item.Name
          Type: item.Item.Type
      myInventory.push newItem
    expect(myInventory).toEqual pastebinTest
