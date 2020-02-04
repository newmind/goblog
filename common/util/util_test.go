package util

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestResolveIp(t *testing.T) {

	Convey("Given a Call request", t, func() {

		Convey("When", func() {
			ipAddress, err := ResolveIpFromHostsFile()

			Convey("Then", func() {
				So(err, ShouldBeNil)
				So(ipAddress, ShouldNotBeNil)
				So(string(ipAddress), ShouldContainSubstring, ".")
			})
		})
	})

}
