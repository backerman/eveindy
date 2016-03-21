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

describe 'Directive: charPicker', () ->
  compile = null
  timeout = null

  element = null
  element2 = null
  scope = null

  beforeEach () ->
    module 'eveindy'
    module 'controller_test'
    module 'directives_test'
    inject ($compile, $rootScope, $timeout, $sniffer) ->
      compile = $compile
      timeout = $timeout
      scope = $rootScope.$new()
      scope.characters = JSON.parse JSON.stringify [
        name: "Arjun Kansene"
        id:   94319654
        corporation: "Center for Advanced Studies"
        corporationID: 1000169
        alliance: ""
        allianceID: 0
      ,
        name: "All reps on Cain"
        id:   123456
        corporation: "Yes, this is test data"
        corporationID: 78910
        alliance: "Some Alliance"
        allianceID: 494949
      ]

      scope.selectedToon = scope.characters[0]
      element = compile(
        """
         <char-picker characters="characters"
            char-selected="selectedToon"
            name="sampleName">
        """) scope

      element2 = compile(
        """
         <char-picker characters="characters"
            char-selected="selectedToon">
        """) scope

      scope.$digest()

  it 'should get the list of characters', () ->
    options = element.find 'option'
    expect(options.length).toBe 2

  it 'should default to the value of charSelected', () ->
    selectedOption = element.find 'option:selected'
    expect(parseInt selectedOption.val()).toBe scope.characters[0].id
    scope.$apply (scope) ->
      scope.selectedToon = scope.characters[1]
    selectedOption = element.find 'option:selected'
    expect(parseInt selectedOption.val()).toBe scope.characters[1].id

  it 'should list characters in alphabetical order', () ->
    options = element.find 'option'
    expect(options.eq(0).val()).toBe '123456'
    expect(options.eq(1).val()).toBe '94319654'

  it 'should change the model when a new choice is selected', () ->
    selectedName = element.find('option:selected').html()
    expect(selectedName).toBe 'Arjun Kansene'

    scope.selectedToon = scope.characters[1]
    scope.$digest()
    selectedName = element.find('option:selected').html()
    expect(selectedName).toBe 'All reps on Cain'

  it 'should use a default name attribute if one is not provided.', () ->
    myName = element.find 'select'
      .get 0
      .getAttribute 'name'
    expect(myName).toBe 'sampleName'

    myDefaultName = element2.find 'select'
      .get 0
      .getAttribute 'name'
    expect(myDefaultName).toBe 'selectedToon'
