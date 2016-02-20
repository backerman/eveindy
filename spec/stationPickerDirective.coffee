# Copyright © 2014–6 Brad Ackerman.
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

angular.module 'eveindy'
  .config ($sceProvider) ->
    $sceProvider.enabled false

describe 'Directive: stationPicker', () ->
  compile = null
  timeout = null

  element = null
  searchedTerm = null
  selectedLocation = null
  scope = null

  # Copying this one from uib tests.
  changeInputValueTo = null

  beforeEach () ->
    module 'eveindy'
    module 'directives_test'
    inject ($compile, $rootScope, $sniffer, $timeout) ->
      compile = $compile
      timeout = $timeout
      scope = $rootScope.$new()
      scope.ctrl =
        getStations: (searchTerm) ->
          searchedTerm = searchTerm
          results = fixture.load('stationPicker-autocomplete.json')
          return JSON.parse JSON.stringify results

        location: null

        locationSelected: (station) ->
          selectedLocation = station

      changeInputValueTo = (element, value) ->
        inputEl = element.find 'input'
        inputEl.val value
        inputEl.trigger if $sniffer.hasEvent 'input' then 'input' else 'change'
        scope.$digest()
        timeout.flush()

      element = compile(
        """<div>
        <station-picker search="ctrl.getStations(prefix)"
        location="ctrl.location"
        selected="ctrl.locationSelected(station)">
        </station-picker></div>
        """) scope
      scope.$digest()

  it 'should pass the search term to the parent', () ->
    searchedTerm = null
    changeInputValueTo element, 'xyz'
    expect(searchedTerm).toEqual 'xyz'

  it 'should get an autocomplete list', () ->
    dropdown = element.find 'ul.dropdown-menu'
    expect(dropdown.text().trim()).toBe("")

    changeInputValueTo element, 'poi'
    dropdown = element.find 'ul.dropdown-menu'
    matches = dropdown.find('li')
    expect(matches.length).toBe(10)
