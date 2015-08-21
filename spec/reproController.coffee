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
  $scope = undefined

  beforeEach () ->
    # $timeout is used in the code under test only because of a workaround
    # for angular-ui/bootstrap until it's actually 1.3-compatible. So we
    # just replace it with a mock that executes the callback immediately.
    module ($provide) ->
      $provide.constant '$timeout', (fn) ->
        fn()

    inject ($controller, Server, _$rootScope_) ->
      $scope = _$rootScope_
      jitaMinerals = fixture.load('jitaMinerals.json')
      spyOn Server, 'getJitaPrices'
        .and.returnValue
          then: (callback) ->
            response =
              data: JSON.parse JSON.stringify jitaMinerals
            callback response

      autocomplete = fixture.load('autocomplete.json')
      spyOn Server, 'getAutocomplete'
        .and.returnValue
          then: (callback) ->
            response =
              data: JSON.parse JSON.stringify autocomplete
            callback response

      pastebinTest = fixture.load('pastebin.json')
      spyOn Server, 'parsePastebin'
        .and.returnValue
          then: (callback) ->
            response =
              data: JSON.parse JSON.stringify pastebinTest
            callback response

      marketOutput = fixture.load('marketOutput.json')
      spyOn Server, 'searchStationMarket'
        .and.returnValue
          then: (callback) ->
            response =
              data: JSON.parse JSON.stringify marketOutput
            callback response

      reprocessOutput = fixture.load('reprocessOutput.json')
      spyOn Server, 'reprocessItems'
        .and.returnValue
          then: (callback) ->
            response =
              data: reprocessOutput
            callback response

      serverService = Server
      ctrl = $controller 'ReprocessCtrl'

  it 'should get Jita mineral prices', () ->
    expect(serverService.getJitaPrices).toHaveBeenCalled()
    expect(serverService.getJitaPrices.calls.count()).toEqual 1

  it 'should populate the correct arrays', () ->
    expect(ctrl.jitaPrices).toEqual(jitaMinerals)

  it 'should correctly calculate midpoint values', () ->
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
    expect(ctrl.imputed).toEqual(midpoints)

  it 'should process autocomplete results', () ->
    testSystem = 'Mir'
    response = ctrl.getStations(testSystem)

    expect(response[0].class).toEqual 'security-null'
    expect(response[0].reprocessingEfficiency).toEqual 50
    expect(response[1].class).toEqual 'security-low'
    expect(response[1].reprocessingEfficiency).toEqual 50
    expect(response[7].class).toEqual 'security-high'
    expect(response[7].reprocessingEfficiency).toEqual 30

  it 'should parse station information', () ->
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

    ctrl.locationSelected(testStation)
    ctrl.standing = 0.00
    ctrl.updateTaxRate()
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
    ctrl.locationSelected(testStation)
    ctrl.submitPaste()
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
