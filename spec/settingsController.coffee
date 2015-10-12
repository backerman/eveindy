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

  beforeEach () ->
    inject (Server, $rootScope) ->
      scope = $rootScope.$new()
      spyOn Server, 'deleteApiKey'
        .and.callFake (keyid) ->
          then: (callback) ->
            callback
              status: 'OK'

      spyOn Server, 'addApiKey'
        .and.callFake (key) ->
          then: (callback) ->
            callback
              data:
                status: 'OK'
                characters: []
            # then is being chained, so we need to ensure it also returns
            # a fake promise.
            then: (cb) ->
              cb()

      spyOn Server, 'getLoginStatus'
        .and.callFake () ->
          then: (callback) ->
            # fixture.load only loads a given file once - so make sure each test
            # has its own API keys to isolate individual tests.
            sessionInfo = JSON.parse JSON.stringify fixture.load('session.json')
            response =
              data: sessionInfo
            callback response
            then: (cb) ->
              cb()

      spyOn Server, 'logout'
        .and.returnValue
          then: (callback) ->
            callback
              status: 'OK'

      spyOn Server, 'getSkills'
        .and.returnValue
          then: (callback) ->
            callback []

      serverService = Server

  beforeEach () ->
    inject ($controller, Session) ->
      sessionService = Session
      ctrl = $controller 'SettingsCtrl',
        $scope: scope
        Session: sessionService

  it 'should get a user\'s API keys', () ->
    expect((k.id for k in ctrl.apikeys).sort()).toEqual [123456, 234567, 345678]

  it 'should remove a deleted key from its local model', () ->
    ctrl.deleteKey 234567
    expect(k.id for k in ctrl.apikeys).not.toContain 234567
    expect(ctrl.apikeys.length).toEqual 2

  it 'should add keys to the user\'s account', () ->
    expect(ctrl.apikeys.length).toEqual 3
    ctrl.newkey =
      id: 666
      userid: 101
      vcode: "abcdefg"
      label: "hijklmnop"
    ctrl.addKey()
    expect(ctrl.apikeys.length).toEqual 4
    expect(k.id for k in ctrl.apikeys).toContain 666
    expect(ctrl.newkey).toEqual {}

  it 'should correctly handle login', () ->
    sessionService._getSessionStatus()
    expect(ctrl.authenticated).toBeTruthy()
    expect(ctrl.apikeys.length).toEqual 3

  it 'should correctly handle logout', () ->
    sessionService.logout()
    expect(ctrl.authenticated).toBeFalsy()
    expect(ctrl.apikeys.length).toEqual 0
