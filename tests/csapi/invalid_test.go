package csapi_tests

import (
	"strings"
	"testing"

	"github.com/matrix-org/complement/internal/b"
	"github.com/matrix-org/complement/internal/client"
	"github.com/matrix-org/complement/internal/match"
	"github.com/matrix-org/complement/internal/must"
	"github.com/matrix-org/complement/runtime"
)

func TestInvalid(t *testing.T) {
	deployment := Deploy(t, b.BlueprintAlice)
	defer deployment.Destroy(t)
	alice := deployment.RegisterUser(t, "hs1", "testInvalidAlice", "AliceSuperPassword", false)
	roomID := alice.CreateRoom(t, map[string]interface{}{
		"room_opts": map[string]interface{}{
			"room_version": "6",
		},
	})

	// these functions are declared in tests/csapi/invalid_test.go
	// if a room is needed, version 6 room or higher
	t.Run("TestJson", func(t *testing.T) {
		// sytest: Invalid JSON integers
		// sytest: Invalid JSON floats
		t.Run("Invalid numerical values", func(t *testing.T) { t.Parallel(); testInvalidJSONNumericalValues(t, alice, roomID) })
		// sytest: Invalid JSON special values
		t.Run("Invalid JSON special values", func(t *testing.T) { t.Parallel(); testInvalidJSONSpecialValues(t, alice, roomID) })
	})
	// these functions are declared in tests/csapi/invalid_test.go
	// if moved, make sure to bring the getFilters() helper function
	// sytest: Check creating invalid filters returns 4xx
	t.Run("TestFilter", func(t *testing.T) {
		runtime.SkipIf(t, runtime.Dendrite) // FIXME: https://github.com/matrix-org/dendrite/issues/2067
		testFilter(t, alice)
	})
	// these functions are declared in tests/csapi/invalid_test.go
	// if a room is needed, version 6 room or higher
	// sytest: Event size limits
	t.Run("TestEvent", func(t *testing.T) {
		t.Run("Large Event", func(t *testing.T) { t.Parallel(); testLargeEvent(t, alice, roomID) })
		t.Run("Large State Event", func(t *testing.T) { t.Parallel(); testLargeStateEvent(t, alice, roomID) })
	})
}

func testInvalidJSONNumericalValues(t *testing.T, userOne *client.CSAPI, roomID string) {
	testCases := [][]byte{
		[]byte(`{"body": 9007199254740992}`),
		[]byte(`{"body": -9007199254740992}`),
		[]byte(`{"body": 1.1}`),
	}

	for _, testCase := range testCases {
		res := userOne.DoFunc(t, "POST", []string{"_matrix", "client", "v3", "rooms", roomID, "send", "complement.dummy"}, client.WithJSONBody(t, testCase))

		must.MatchResponse(t, res, match.HTTPResponse{
			StatusCode: 400,
			JSON: []match.JSON{
				match.JSONKeyEqual("errcode", "M_BAD_JSON"),
			},
		})
	}

}
func testInvalidJSONSpecialValues(t *testing.T, userOne *client.CSAPI, roomID string) {
	testCases := [][]byte{
		[]byte(`{"body": Infinity}`),
		[]byte(`{"body": -Infinity}`),
		[]byte(`{"body": NaN}`),
	}

	for _, testCase := range testCases {
		res := userOne.DoFunc(t, "POST", []string{"_matrix", "client", "v3", "rooms", roomID, "send", "complement.dummy"}, client.WithJSONBody(t, testCase))

		must.MatchResponse(t, res, match.HTTPResponse{
			StatusCode: 400,
		})
	}
}

func testFilter(t *testing.T, userOne *client.CSAPI) {
	filters := getFilters()

	for _, filter := range filters {
		res := userOne.DoFunc(t, "POST", []string{"_matrix", "client", "v3", "user", userOne.UserID, "filter"}, client.WithJSONBody(t, filter))

		if res.StatusCode >= 500 || res.StatusCode < 400 {
			t.Errorf("Expected 4XX status code, got %d for testing filter %s", res.StatusCode, filter)
		}
	}
}

func testLargeEvent(t *testing.T, userOne *client.CSAPI, roomID string) {
	// needs a version 6 room and one user
	event := map[string]interface{}{
		"msgtype": "m.text",
		"body":    strings.Repeat("and they dont stop coming ", 2700), // 2700 * 26 == 70200
	}
	res := userOne.DoFunc(t, "PUT", []string{"_matrix", "client", "v3", "rooms", roomID, "send", "m.room.message", "1"}, client.WithJSONBody(t, event))
	must.MatchResponse(t, res, match.HTTPResponse{
		StatusCode: 413,
	})
}

func testLargeStateEvent(t *testing.T, userOne *client.CSAPI, roomID string) {
	// needs a version 6 room and one user
	stateEvent := map[string]interface{}{
		"body": strings.Repeat("Dormammu, I've Come To Bargain.\n", 2200), // 2200 * 32 == 70400
	}
	res := userOne.DoFunc(t, "PUT", []string{"_matrix", "client", "v3", "rooms", roomID, "state", "marvel.universe.fate"}, client.WithJSONBody(t, stateEvent))
	must.MatchResponse(t, res, match.HTTPResponse{
		StatusCode: 413,
	})
}

// small helper function to not bloat the main one
// todo: this should be more exhaustive and up-to-date
// todo: this should be easier to construct
func getFilters() []map[string]interface{} {
	const NAO = "not_an_object"
	const NAL = "not_a_list"

	return []map[string]interface{}{
		{
			"presence": NAO,
		},

		{
			"room": map[string]interface{}{
				"timeline": NAO,
			},
		},
		{
			"room": map[string]interface{}{
				"state": NAO,
			},
		},
		{
			"room": map[string]interface{}{
				"ephemeral": NAO,
			},
		},
		{
			"room": map[string]interface{}{
				"account_data": NAO,
			},
		},

		{
			"room": map[string]interface{}{
				"timeline": map[string]interface{}{
					"rooms": NAL,
				},
			},
		},
		{
			"room": map[string]interface{}{
				"timeline": map[string]interface{}{
					"not_rooms": NAL,
				},
			},
		},
		{
			"room": map[string]interface{}{
				"timeline": map[string]interface{}{
					"senders": NAL,
				},
			},
		},
		{
			"room": map[string]interface{}{
				"timeline": map[string]interface{}{
					"not_senders": NAL,
				},
			},
		},
		{
			"room": map[string]interface{}{
				"timeline": map[string]interface{}{
					"types": NAL,
				},
			},
		},
		{
			"room": map[string]interface{}{
				"timeline": map[string]interface{}{
					"not_types": NAL,
				},
			},
		},

		{
			"room": map[string]interface{}{
				"timeline": map[string]interface{}{
					"types": []int{1},
				},
			},
		},
		{
			"room": map[string]interface{}{
				"timeline": map[string]interface{}{
					"rooms": []string{"not_a_room_id"},
				},
			},
		},
		{
			"room": map[string]interface{}{
				"timeline": map[string]interface{}{
					"senders": []string{"not_a_sender_id"},
				},
			},
		},
	}
}
