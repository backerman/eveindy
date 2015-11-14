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

describe 'Controller: SettingsCtrl', () ->
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
      spyOn Server, 'deleteApiKey'
        .and.callFake (keyid) ->
          deferred = $q.defer()
          deferred.resolve
            data:
              status: 'OK'
          deferred.promise

      spyOn Server, 'addApiKey'
        .and.callFake (key) ->
          deferred = $q.defer()
          if throwError
            deferred.reject 'ERROR! RUN!!!'
          else
            deferred.resolve
              data:
                status: 'OK'
                characters: []
          deferred.promise

      spyOn Server, 'getLoginStatus'
        .and.callFake () ->
          deferred = $q.defer()
          # fixture.load only loads a given file once - so make sure each test
          # has its own API keys to isolate individual tests.
          sessionInfo = JSON.parse JSON.stringify fixture.load('session.json')
          deferred.resolve
            data: sessionInfo
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

      serverService = Server

  beforeEach () ->
    inject ($controller, Session) ->
      sessionService = Session
      ctrl = $controller 'SettingsCtrl',
        $scope: scope
        Session: sessionService

  it 'should get a user\'s API keys', () ->
    # $apply() needs to be called to execute asynchronous methods.
    scope.$apply()
    expect((k.id for k in ctrl.apikeys).sort()).toEqual [123456, 234567, 345678]

  it 'should remove a deleted key from its local model', inject () ->
    ctrl.deleteKey 234567
    scope.$apply()
    expect(k.id for k in ctrl.apikeys).not.toContain 234567
    expect(ctrl.apikeys.length).toEqual 2

  it 'should add keys to the user\'s account', () ->
    scope.$apply()
    expect(ctrl.apikeys.length).toEqual 3
    ctrl.newkey =
      id: 666
      userid: 101
      vcode: "abcdefg"
      label: "hijklmnop"
    ctrl.addKey()
    scope.$apply()
    expect(ctrl.apikeys.length).toEqual 4
    expect(k.id for k in ctrl.apikeys).toContain 666
    expect(ctrl.newkey).toEqual {}

  it 'should handle failed API key adds', () ->
    scope.$apply()
    throwError = true
    ctrl.newkey =
      id: 666
      userid: 101
      vcode: "abcdefg"
      label: "hijklmnop"
    ctrl.addKey()
    console.log 'Not yet applied.'
    scope.$apply()
    console.log 'Applied.'
    expect(ctrl.alerts.length).toEqual 1
    ctrl.closeAlert(0)
    expect(ctrl.alerts.length).toEqual 0

  it 'should correctly handle login', () ->
    sessionService._getSessionStatus()
    scope.$apply()
    expect(ctrl.authenticated).toBeTruthy()
    expect(ctrl.apikeys.length).toEqual 3

  it 'should correctly handle logout', () ->
    sessionService.logout()
    scope.$apply()
    expect(ctrl.authenticated).toBeFalsy()
    expect(ctrl.apikeys.length).toEqual 0
