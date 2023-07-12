package csapi_tests

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/matrix-org/complement/internal/b"
	"github.com/matrix-org/complement/internal/client"
	"github.com/matrix-org/complement/internal/match"
	"github.com/matrix-org/complement/internal/must"
)

// TODO most of this can be refactored into data-driven tests

func fetchEvent(t *testing.T, c *client.CSAPI, roomId, eventId string) *http.Response {
	return c.DoFunc(t, "GET", []string{"_matrix", "client", "v3", "rooms", roomId, "event", eventId})
}

func createRoomWithVisibility(t *testing.T, c *client.CSAPI, visibility string) string {
	return c.CreateRoom(t, map[string]interface{}{
		"initial_state": []map[string]interface{}{
			{
				"content": map[string]interface{}{
					"history_visibility": visibility,
				},
				"type":      "m.room.history_visibility",
				"state_key": "",
			},
		},
		"preset": "public_chat",
	})
}

// Fetches an event after join, and succeeds.
// sytest: /event/ on joined room works
func TestFetchEvent(t *testing.T) {
	deployment := Deploy(t, b.BlueprintOneToOneRoom)
	defer deployment.Destroy(t)

	alice := deployment.Client(t, "hs1", "@alice:hs1")
	bob := deployment.Client(t, "hs1", "@bob:hs1")

	roomID := createRoomWithVisibility(t, alice, "shared")

	bob.JoinRoom(t, roomID, nil)

	bob.MustSyncUntil(t, client.SyncReq{}, client.SyncJoinedTo(bob.UserID, roomID))

	eventID := alice.SendEventSynced(t, roomID, b.Event{
		Type: "m.room.message",
		Content: map[string]interface{}{
			"msgtype": "m.text",
			"body":    "Hello world",
		},
	})

	res := fetchEvent(t, bob, roomID, eventID)

	must.MatchResponse(t, res, match.HTTPResponse{
		StatusCode: 200,
		JSON: []match.JSON{
			// No harm in checking if the event data is also as expected
			match.JSONKeyEqual("content", map[string]interface{}{
				"msgtype": "m.text",
				"body":    "Hello world",
			}),
			match.JSONKeyEqual("type", "m.room.message"),

			// the spec technically doesn't list these following keys, but we're still checking them because sytest did.
			// see: https://github.com/matrix-org/matrix-doc/issues/3540
			match.JSONKeyEqual("room_id", roomID),
			match.JSONKeyEqual("sender", alice.UserID),
			match.JSONKeyEqual("event_id", eventID),
			match.JSONKeyTypeEqual("origin_server_ts", gjson.Number),
		},
	})
}

// Tries to fetch an event before join, and fails.
// history_visibility: joined
// sytest: /event/ does not allow access to events before the user joined
func TestFetchHistoricalJoinedEventDenied(t *testing.T) {
	deployment := Deploy(t, b.BlueprintOneToOneRoom)
	defer deployment.Destroy(t)

	alice := deployment.Client(t, "hs1", "@alice:hs1")
	bob := deployment.Client(t, "hs1", "@bob:hs1")

	roomID := createRoomWithVisibility(t, alice, "joined")

	eventID := alice.SendEventSynced(t, roomID, b.Event{
		Type: "m.room.message",
		Content: map[string]interface{}{
			"msgtype": "m.text",
			"body":    "Hello world",
		},
	})

	bob.JoinRoom(t, roomID, nil)
	bob.MustSyncUntil(t, client.SyncReq{}, client.SyncJoinedTo(bob.UserID, roomID))

	res := fetchEvent(t, bob, roomID, eventID)

	must.MatchResponse(t, res, match.HTTPResponse{
		StatusCode: 404,
	})
}

// Tries to fetch an event before join, and succeeds.
// history_visibility: shared
func TestFetchHistoricalSharedEvent(t *testing.T) {
	deployment := Deploy(t, b.BlueprintOneToOneRoom)
	defer deployment.Destroy(t)

	alice := deployment.Client(t, "hs1", "@alice:hs1")
	bob := deployment.Client(t, "hs1", "@bob:hs1")

	roomID := createRoomWithVisibility(t, alice, "shared")

	eventID := alice.SendEventSynced(t, roomID, b.Event{
		Type: "m.room.message",
		Content: map[string]interface{}{
			"msgtype": "m.text",
			"body":    "Hello world",
		},
	})

	bob.JoinRoom(t, roomID, nil)
	bob.MustSyncUntil(t, client.SyncReq{}, client.SyncJoinedTo(bob.UserID, roomID))

	res := fetchEvent(t, bob, roomID, eventID)

	must.MatchResponse(t, res, match.HTTPResponse{
		StatusCode: 200,
		JSON: []match.JSON{
			// No harm in checking if the event data is also as expected
			match.JSONKeyEqual("content", map[string]interface{}{
				"msgtype": "m.text",
				"body":    "Hello world",
			}),
			match.JSONKeyEqual("type", "m.room.message"),

			// the spec technically doesn't list these following keys, but we're still checking them because sytest did.
			// see: https://github.com/matrix-org/matrix-doc/issues/3540
			match.JSONKeyEqual("room_id", roomID),
			match.JSONKeyEqual("sender", alice.UserID),
			match.JSONKeyEqual("event_id", eventID),
			match.JSONKeyTypeEqual("origin_server_ts", gjson.Number),
		},
	})
}

// Tries to fetch an event between being invited and joined, and succeeds.
// history_visibility: invited
func TestFetchHistoricalInvitedEventFromBetweenInvite(t *testing.T) {
	deployment := Deploy(t, b.BlueprintOneToOneRoom)
	defer deployment.Destroy(t)

	alice := deployment.Client(t, "hs1", "@alice:hs1")
	bob := deployment.Client(t, "hs1", "@bob:hs1")

	roomID := createRoomWithVisibility(t, alice, "invited")

	alice.InviteRoom(t, roomID, bob.UserID)
	bob.MustSyncUntil(t, client.SyncReq{}, client.SyncInvitedTo(bob.UserID, roomID))

	eventID := alice.SendEventSynced(t, roomID, b.Event{
		Type: "m.room.message",
		Content: map[string]interface{}{
			"msgtype": "m.text",
			"body":    "Hello world",
		},
	})

	bob.JoinRoom(t, roomID, nil)
	bob.MustSyncUntil(t, client.SyncReq{}, client.SyncJoinedTo(bob.UserID, roomID))

	res := fetchEvent(t, bob, roomID, eventID)

	must.MatchResponse(t, res, match.HTTPResponse{
		StatusCode: 200,
		JSON: []match.JSON{
			// No harm in checking if the event data is also as expected
			match.JSONKeyEqual("content", map[string]interface{}{
				"msgtype": "m.text",
				"body":    "Hello world",
			}),
			match.JSONKeyEqual("type", "m.room.message"),

			// the spec technically doesn't list these following keys, but we're still checking them because sytest did.
			// see: https://github.com/matrix-org/matrix-doc/issues/3540
			match.JSONKeyEqual("room_id", roomID),
			match.JSONKeyEqual("sender", alice.UserID),
			match.JSONKeyEqual("event_id", eventID),
			match.JSONKeyTypeEqual("origin_server_ts", gjson.Number),
		},
	})
}

// Tries to fetch an event before being invited, and fails.
// history_visibility: invited
func TestFetchHistoricalInvitedEventFromBeforeInvite(t *testing.T) {
	deployment := Deploy(t, b.BlueprintOneToOneRoom)
	defer deployment.Destroy(t)

	alice := deployment.Client(t, "hs1", "@alice:hs1")
	bob := deployment.Client(t, "hs1", "@bob:hs1")

	roomID := createRoomWithVisibility(t, alice, "invited")

	eventID := alice.SendEventSynced(t, roomID, b.Event{
		Type: "m.room.message",
		Content: map[string]interface{}{
			"msgtype": "m.text",
			"body":    "Hello world",
		},
	})

	alice.InviteRoom(t, roomID, bob.UserID)
	bob.MustSyncUntil(t, client.SyncReq{}, client.SyncInvitedTo(bob.UserID, roomID))

	bob.JoinRoom(t, roomID, nil)
	bob.MustSyncUntil(t, client.SyncReq{}, client.SyncJoinedTo(bob.UserID, roomID))

	res := fetchEvent(t, bob, roomID, eventID)

	must.MatchResponse(t, res, match.HTTPResponse{
		StatusCode: 404,
	})
}

// Tries to fetch an event without having joined, and fails.
// history_visibility: shared
// sytest: /event/ on non world readable room does not work
func TestFetchEventNonWorldReadable(t *testing.T) {
	deployment := Deploy(t, b.BlueprintOneToOneRoom)
	defer deployment.Destroy(t)

	alice := deployment.Client(t, "hs1", "@alice:hs1")
	bob := deployment.Client(t, "hs1", "@bob:hs1")

	roomID := createRoomWithVisibility(t, alice, "shared")

	eventID := alice.SendEventSynced(t, roomID, b.Event{
		Type: "m.room.message",
		Content: map[string]interface{}{
			"msgtype": "m.text",
			"body":    "Hello world",
		},
	})

	res := fetchEvent(t, bob, roomID, eventID)

	must.MatchResponse(t, res, match.HTTPResponse{
		StatusCode: 404,
	})
}

// Tries to fetch an event without having joined, and succeeds.
// history_visibility: world_readable
func TestFetchEventWorldReadable(t *testing.T) {
	deployment := Deploy(t, b.BlueprintOneToOneRoom)
	defer deployment.Destroy(t)

	alice := deployment.Client(t, "hs1", "@alice:hs1")
	bob := deployment.Client(t, "hs1", "@bob:hs1")

	roomID := createRoomWithVisibility(t, alice, "world_readable")

	eventID := alice.SendEventSynced(t, roomID, b.Event{
		Type: "m.room.message",
		Content: map[string]interface{}{
			"msgtype": "m.text",
			"body":    "Hello world",
		},
	})

	res := fetchEvent(t, bob, roomID, eventID)

	must.MatchResponse(t, res, match.HTTPResponse{
		StatusCode: 200,
		JSON: []match.JSON{
			// No harm in checking if the event data is also as expected
			match.JSONKeyEqual("content", map[string]interface{}{
				"msgtype": "m.text",
				"body":    "Hello world",
			}),
			match.JSONKeyEqual("type", "m.room.message"),

			// the spec technically doesn't list these following keys, but we're still checking them because sytest did.
			// see: https://github.com/matrix-org/matrix-doc/issues/3540
			match.JSONKeyEqual("room_id", roomID),
			match.JSONKeyEqual("sender", alice.UserID),
			match.JSONKeyEqual("event_id", eventID),
			match.JSONKeyTypeEqual("origin_server_ts", gjson.Number),
		},
	})
}

// Tries to fetch an event without having joined, and succeeds.
// history_visibility: world_readable
// NOTE: uses older /events api which is deprecated in order to simulate sytest.

// create two new users(hermetics)
// create world_readable room and join alice
// alice sends a message so the room has something beyond 'state' to sync.
// have alice sync until her message shows up(to get the stream_token)
// have bob call /events to get a sync token and clear out anything up to this point
// set presence on alice with a status message
// have alice sync again(this doesn't help the test, it's so I can collect debugging on the Synapse side)
// call /events on bob with the stream token to get anything new, which should get alice's status message.
func TestFetchEventWorldReadableUsingDeprecatedEventsEndpoint(t *testing.T) {
	deployment := Deploy(t, b.BlueprintOneToOneRoom)
	defer deployment.Destroy(t)

	alice := deployment.NewUser(t, "t01alice", "hs1")
	bob := deployment.NewUser(t, "t01bob", "hs1")

	roomID := createRoomWithVisibility(t, alice, "world_readable")

	_, BobsSyncToken := bob.MustSync(t, client.SyncReq{TimeoutMillis: "0"})
	t.Logf("JASON: bob's sync token %v", BobsSyncToken)

	alice.SendEventSynced(t, roomID, b.Event{
		Type: "m.room.message",
		Content: map[string]interface{}{
			"msgtype": "m.text",
			"body":    "Hello world",
		},
	})

	// This starts the actual test. First, get bob a sync token for later.
	// Apparently, room_id as a query on this endpoint is undocumented and not in the
	// spec, but works. This *appears* to duplicate how sytest uses this endpoint.
	query := url.Values{
		"timeout": []string{"500"},
		"from": []string{BobsSyncToken},
		"room_id": []string{roomID},
	}
	res := fetchNewEvents(t, bob, query)

	// pull the new sync token out from res
	res_json := client.ParseJSON(t, res)
	BobsSyncToken = client.GetJSONFieldStr(t, res_json, "end")
	t.Logf("JASON: bob's response %s", res_json)

	// grab alice's sync token so we can make sure she sees her own presence
	_, AlicesSyncToken := alice.MustSync(t, client.SyncReq{TimeoutMillis: "0"})

	// alice sets a presence status message
	statusMsg := "Update for room members"
	alice.MustDoFunc(t, "PUT", []string{"_matrix", "client", "v3", "presence", alice.UserID, "status"},
		client.WithJSONBody(t, map[string]interface{}{
			"status_msg": statusMsg,
			"presence":   "online",
		}),
	)

	// and then syncs until it shows up. Ultimately, this has no bearing on the test
	// itself, but allows hitting the UserPresenceSource to check for updates.
	alice.MustSyncUntil(t, client.SyncReq{Since: AlicesSyncToken},
			client.SyncPresenceHas(alice.UserID, b.Ptr("online"), func(ev gjson.Result) bool {
				return ev.Get("content.status_msg").Str == statusMsg
			}),
		)

	// reset and get new results, this should pick up alice's status message
	query["from"] = []string{BobsSyncToken}
	res = fetchNewEvents(t, bob, query)
	// pull the new sync token out from res
	res_json = client.ParseJSON(t, res)
	BobsSyncToken = client.GetJSONFieldStr(t, res_json, "end")

	t.Logf("JASON: bob's response %s", res_json)
	t.Logf("JASON: bob's new sync token %s", BobsSyncToken)

	// So this doesn't work right, my JSON matching skills aren't up to par.
	// Something something read on closed response body something
	must.MatchResponse(t, res, match.HTTPResponse{
		StatusCode: 200,
		JSON: []match.JSON{
			match.JSONKeyEqual("chunk", map[string]interface{}{
				"presence": "online",
				"user_id": alice.UserID,
				"status_message": statusMsg,
			}),
			match.JSONKeyEqual("type", "m.presence"),
		},
	})

}

func fetchNewEvents(t *testing.T, c *client.CSAPI, query url.Values) *http.Response {
	return c.MustDoFunc(t, "GET", []string{"_matrix", "client", "v3", "events"}, client.WithQueries(query))
}

