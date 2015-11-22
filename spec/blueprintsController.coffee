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

describe 'Controller: BlueprintsCtrl', () ->
  beforeEach () -> module 'eveindy'
  ctrl = undefined
  serverService = undefined
  sessionService = undefined
  apiKeys = undefined
  scope = undefined
  throwError = false

  beforeEach () ->
    inject (Server, $rootScope, $q) ->
      scope = $rootScope.$new()
      throwError = false

      spyOn Server, 'getLoginStatus'
        .and.callFake () ->
          deferred = $q.defer()
          # fixture.load only loads a given file once - so make sure each test
          # has its own API keys to isolate individual tests.
          sessionInfo = JSON.parse JSON.stringify fixture.load('session.json')
          deferred.resolve
            data: sessionInfo
          deferred.promise

      spyOn Server, 'getBlueprints'
        .and.callFake (charID) ->
          deferred = $q.defer()
          blueprints = JSON.parse JSON.stringify fixture.load 'blueprints.json'
          deferred.resolve
            data: blueprints
          deferred.promise

      spyOn Server, 'logout'
        .and.callFake () ->
          deferred = $q.defer()
          deferred.resolve
            data:
              status: 'OK'
          deferred.promise

      spyOn Server, 'getSkills'
        .and.callFake () ->
          deferred = $q.defer()
          deferred.resolve []
          deferred.promise

      # FIXME need real test data here
      spyOn Server, 'getUnusedSalvage'
        .and.callFake () ->
          deferred = $q.defer()
          deferred.resolve
            data:
              items: []
              stations: {}
              itemInfo: {}
          deferred.promise

      serverService = Server

  beforeEach () ->
    inject ($controller, Session) ->
      sessionService = Session
      ctrl = $controller 'BlueprintsCtrl',
        $scope: scope
        Session: sessionService

  it 'should get the current characters', () ->
    # $apply() needs to be called to execute asynchronous methods.
    scope.$apply()
    expect((c.id for c in ctrl.characters).sort()).toEqual [123456, 94319654]

  it 'should default to having a character selected', () ->
    scope.$apply()
    expect(ctrl.selectedToon.id).toEqual 94319654

  it 'should get the character\'s blueprints', () ->
    scope.$apply()
    expect(ctrl.blueprints.length).toEqual 4
