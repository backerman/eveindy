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

# N.B.: This file is where our modules are defined, and since the
# CoffeeScript files are included in lexicographical order, nothing
# should be named as to come before app.coffee.

angular.module 'eveindy', [
    'ui.bootstrap'
    'ngRoute'
    'datatables'
    'datatables.bootstrap'
    'angularPromiseButtons'
  ]
