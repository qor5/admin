package publish_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_admin"
	"github.com/qor5/admin/v3/utils/testflow"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/stretchr/testify/assert"
)

func MustSplitIDVersion(expr string) []string {
	segs := strings.Split(expr, "_")
	if len(segs) < 2 {
		panic(fmt.Errorf("invalid expr %q", expr))
	}
	return segs[0:2]
}

func MustIDVersion(expr string) (string, string) {
	segs := MustSplitIDVersion(expr)
	return segs[0], segs[1]
}

func GetNextVersion(currentVersion string) (string, error) {
	parts := strings.Split(currentVersion, "_")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid version format")
	}

	id := parts[0]
	dateVersionPart := parts[1]
	dateVersion := strings.Split(dateVersionPart, "-")
	if len(dateVersion) != 4 {
		return "", fmt.Errorf("invalid date-version part format")
	}

	dateStr, versionStr := strings.Join(dateVersion[0:3], "-"), dateVersion[3]
	versionNumberStr := strings.TrimPrefix(versionStr, "v")
	versionNumber, err := strconv.Atoi(versionNumberStr)
	if err != nil {
		return "", fmt.Errorf("invalid version number")
	}

	currentDate := time.Now().Local().Format("2006-01-02")

	var nextVersion string
	if dateStr == currentDate {
		nextVersion = fmt.Sprintf("%s_%s-v%02d", id, currentDate, versionNumber+1)
	} else {
		nextVersion = fmt.Sprintf("%s_%s-v01", id, currentDate)
	}

	return nextVersion, nil
}

func ContainsVersionBar(body string) bool {
	return strings.Contains(body, "presets_OpenListingDialog") && strings.Contains(body, "-version-list-dialog")
}

var reListContent = regexp.MustCompile(`<tr[\s\S]+?<td[^>]*>[\s\S]+?<v-radio :model-value='([^']+)'\s*:true-value='([^']+)'[\s\S]+?</v-radio>\s*([^<]+)?\s*</div>[\s\S]+?</tr>`)

func EnsureVersionListDisplay(selected string, dislayModels []*examples_admin.WithPublishProduct) testflow.ValidatorFunc {
	return testflow.Combine(
		// Ensure list head display
		testflow.ContainsInOrderAtUpdatePortal(0,
			// Ensure tabs display
			"<v-tabs",
			"active_filter_tab", "f_all", "f_select_id", selected, "All Versions",
			"active_filter_tab", "f_named_versions", "f_select_id", selected, "Named Versions",
			"</v-tabs>",
			// Ensure columns display
			"<tr>", ">Version</th>", ">Status</th>", ">Start At</th>", ">End At</th>", ">Notes</th>", ">Option</th>", "</tr>",
		),
		// Ensure list content display
		testflow.WrapEvent(func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, e multipartestutils.TestEventResponse) {
			subs := reListContent.FindAllStringSubmatch(e.UpdatePortals[0].Body, -1)
			assert.Len(t, subs, len(dislayModels))
			for i, sub := range subs {
				// ensure only selected item be marked
				modelValue, _ := strconv.Unquote(sub[1])
				trueValue, _ := strconv.Unquote(sub[2])
				assert.Equal(t, dislayModels[i].PrimarySlug(), modelValue)
				assert.Equal(t, selected, trueValue)
				// ensure display version name , not version
				assert.Equal(t, dislayModels[i].Version.VersionName, sub[3])
			}
		}),
	)
}
