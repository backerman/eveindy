# Copyright © 2014–6 Brad Ackerman.
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

angular.module 'eveindy'
  .directive 'charPicker', ['$timeout', ($timeout) ->
    templateUrl: 'view/directives/char-picker.html'
    restrict: 'E'
    transclude: true
    scope:
      characters: '='
      name: '@'
      selectedCharacter: '=charSelected'
      selected: '&changed'
    link: ($scope, $element, $attrs) ->
      # Set a default input field name.
      $scope.name ?= "selectedToon"
      # Ensure that the selected function is only executed after the model
      # has been updated.
      $scope.selectionChanged = (params) ->
        $timeout(() ->
          $scope.selected(params))
]
