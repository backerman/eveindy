/*
Copyright © 2014–6 Brad Ackerman.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package server_test

import (
	"encoding/base64"
	"testing"

	"github.com/backerman/eveindy/pkg/server"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRandomness(t *testing.T) {
	Convey("Verify output of randomness generator is correctly formatted", t, func() {
		// 32 bytes of entropy is 42.6 bytes of Base64 output, which will be
		// padded to 44 bytes.
		Convey("Requesting 32 bytes of entropy", func() {
			out32b, err := server.GetRandomness(32)
			Convey("The output is of the correct length", func() {
				So(err, ShouldBeNil)
				So(len(out32b), ShouldEqual, 44)
			})
			Convey("It is valid Base-64", func() {
				decodedbytes, err := base64.StdEncoding.DecodeString(out32b)
				So(err, ShouldBeNil)
				So(len(decodedbytes), ShouldEqual, 32)
			})
		})

		// 9 bytes of entropy is 12 bytes of Base64 output.
		Convey("Requesting 9 bytes of entropy", func() {
			out9b, err := server.GetRandomness(9)
			Convey("The output is of the correct length", func() {
				So(err, ShouldBeNil)
				So(len(out9b), ShouldEqual, 12)
			})
			Convey("It is valid Base-64", func() {
				decodedbytes, err := base64.StdEncoding.DecodeString(out9b)
				So(err, ShouldBeNil)
				So(len(decodedbytes), ShouldEqual, 9)
			})
		})
	})
}
