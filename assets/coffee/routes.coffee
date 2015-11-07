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

angular.module 'eveindy'
  .config ['$routeProvider', ($routeProvider) ->
    $routeProvider.when '/',
      templateUrl: 'view/reprocess.html'
      controller: 'ReprocessCtrl as ctrl'
    .when '/blueprints',
      templateUrl: 'view/blueprints.html'
      controller: 'BlueprintsCtrl as ctrl'
    .when '/settings',
      templateUrl: 'view/settings.html'
      controller: 'SettingsCtrl as ctrl'
    .otherwise
      redirectTo: '/'
    ]
