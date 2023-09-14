package pagebuilder

import (
	"path"
	"testing"
)

func TestSlugReg(t *testing.T) {
	cases := []string{
		"/apple",
		"/Apple",
		"/fruit/Apple",
		"/fruit/Apple/Red",
		"/fruit/bAnAnA/yElLoW",
		"/vegetable/Carrot/Orange",
		"/animal/Dog/Brown",
		"/animal/Cat/Gray",
		"/vehicle/Car/Red",
		"/City/New-York",
		"/Country/United-States/New-York",
		"/fruit/apple",
		"/fruit/apple/red",
		"/fruit/banana/yellow",
		"/vegetable/carrot/orange",
		"/animal/dog/brown",
		"/animal/cat/gray",
		"/vehicle/car/red",
		"/city/new-york",
		"/country/united-states/new-york",
		"/user/john_doe",
		"/product/12345",
		"/page/home",
		"/file/file-name",
		"/images/image_01",
		"/folder/sub_folder",
		"/route/route-1",
		"/data/data123",
		"/user/user12345/profile",
	}

	casesNotMatch := []string{
		"apple",
		"*apple",
		"$fruit/apple",
		"/fruit/apple&banana",
		"#vegetable/carrot",
		"/animal/dog:",
		"/animal/dog#cat",
		"%20images/image01",
		`/user\john`,
		"/product?item=123",
		`/my$folder`,
		`/usr#bin`,
		`/home/user&docs`,
		`/var/www/my folder`,
		`/home/user/documents/with space`,
		`/home/user*docs`,
		`/usr?bin`,
		`/var/www/my\tfolder`,
		`/home/user/documents\nwithnewline`,
	}

	for _, c := range cases {
		if !directoryRe.MatchString(path.Clean(c)) {
			t.Errorf("directoryRe.MatchString(%q) = false, want true", c)
		}
	}

	for _, c := range casesNotMatch {
		if directoryRe.MatchString(path.Clean(c)) {
			t.Errorf("directoryRe.MatchString(%q) = true, want false", c)
		}
	}
}
