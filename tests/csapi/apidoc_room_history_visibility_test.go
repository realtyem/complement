package csapi_tests

import (
	"net/http"
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

func TestRoomHistoryVisibility(t *testing.T) {
	deployment := Deploy(t, b.BlueprintOneToOneRoom)
	defer deployment.Destroy(t)

	t.Run("Parallel", func(t *testing.T) {
		// Fetches an event after join, and succeeds.
		// sytest: /event/ on joined room works
		t.Run("FetchEvent", func(t *testing.T) {
			//formerly TestFetchEvent
			// with users @alice:hs1 and @bob:hs1
			t.Parallel()

			alice := deployment.RegisterUser(t, "hs1", "t01alice", "secret", false)
			bob := deployment.RegisterUser(t, "hs1", "t01bob", "secret", false)

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
		})

		// Tries to fetch an event before join, and fails.
		// history_visibility: joined
		// sytest: /event/ does not allow access to events before the user joined
		t.Run("FetchHistoricalJoinedEventDenied", func(t *testing.T) {
			//formerly TestFetchHistoricalJoinedEventDenied
			// with users @alice:hs1 and @bob:hs1
			t.Parallel()

			alice := deployment.RegisterUser(t, "hs1", "t02alice", "secret", false)
			bob := deployment.RegisterUser(t, "hs1", "t02bob", "secret", false)

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
		})

		// Tries to fetch an event before join, and succeeds.
		// history_visibility: shared
		t.Run("FetchHistoricalSharedEvent", func(t *testing.T) {
			//formerly TestFetchHistoricalSharedEvent
			// with users @alice:hs1 and @bob:hs1
			t.Parallel()

			alice := deployment.RegisterUser(t, "hs1", "t03alice", "secret", false)
			bob := deployment.RegisterUser(t, "hs1", "t03bob", "secret", false)

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
		})

		// Tries to fetch an event between being invited and joined, and succeeds.
		// history_visibility: invited
		t.Run("FetchHistoricalInvitedEventFromBetweenInvite", func(t *testing.T) {
			//formerly TestFetchHistoricalInvitedEventFromBetweenInvite
			// with users @alice:hs1 and @bob:hs1
			t.Parallel()

			alice := deployment.RegisterUser(t, "hs1", "t04alice", "secret", false)
			bob := deployment.RegisterUser(t, "hs1", "t04bob", "secret", false)

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
		})

		// Tries to fetch an event before being invited, and fails.
		// history_visibility: invited
		t.Run("FetchHistoricalInvitedEventFromBeforeInvite", func(t *testing.T) {
			//formerly TestFetchHistoricalInvitedEventFromBeforeInvite
			// with users @alice:hs1 and @bob:hs1
			t.Parallel()

			alice := deployment.RegisterUser(t, "hs1", "t05alice", "secret", false)
			bob := deployment.RegisterUser(t, "hs1", "t05bob", "secret", false)

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
		})

		// Tries to fetch an event without having joined, and fails.
		// history_visibility: shared
		// sytest: /event/ on non world readable room does not work
		t.Run("FetchEventNonWorldReadable", func(t *testing.T) {
			//formerly TestFetchEventNonWorldReadable
			// with users @alice:hs1 and @bob:hs1
			t.Parallel()

			alice := deployment.RegisterUser(t, "hs1", "t06alice", "secret", false)
			bob := deployment.RegisterUser(t, "hs1", "t06bob", "secret", false)

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
		})

		// Tries to fetch an event without having joined, and succeeds.
		// history_visibility: world_readable
		t.Run("FetchEventWorldReadable", func(t *testing.T) {
			// formerly TestFetchEventWorldReadable
			// with users @alice:hs1 and @bob:hs1
			t.Parallel()
			alice := deployment.RegisterUser(t, "hs1", "t07alice", "secret", false)
			bob := deployment.RegisterUser(t, "hs1", "t07bob", "secret", false)

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
		})
	})
}
