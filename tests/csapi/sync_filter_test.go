package csapi_tests

import (
	"io/ioutil"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/matrix-org/complement/internal/client"
	"github.com/matrix-org/complement/internal/docker"
	"github.com/matrix-org/complement/internal/match"
	"github.com/matrix-org/complement/internal/must"
)

// abstracted tests below. These are now called from sync_test.go

// sytest: Can create filter
// sytest: Can download filter
func testSyncCreateAndDownloadFilter(t *testing.T, deployment *docker.Deployment) {
	alice := deployment.NewUser(t, "tSyncFilterAlice", "hs1")

	filterID := createFilter(t, alice, map[string]interface{}{
		"room": map[string]interface{}{
			"timeline": map[string]int{
				"limit": 10,
			},
		},
	})
	res := alice.MustDoFunc(t, "GET", []string{"_matrix", "client", "v3", "user", alice.UserID, "filter", filterID})
	must.MatchResponse(t, res, match.HTTPResponse{
		JSON: []match.JSON{
			match.JSONKeyPresent("room"),
			match.JSONKeyEqual("room.timeline.limit", float64(10)),
		},
	})
}

func createFilter(t *testing.T, c *client.CSAPI, filterContent map[string]interface{}) string {
	t.Helper()
	res := c.MustDoFunc(t, "POST", []string{"_matrix", "client", "v3", "user", c.UserID, "filter"}, client.WithJSONBody(t, filterContent))
	if res.StatusCode != 200 {
		t.Fatalf("MatchResponse got status %d want 200", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("unable to read response body: %v", err)
	}

	filterID := gjson.GetBytes(body, "filter_id").Str

	return filterID

}
