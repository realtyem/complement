//go:build msc3787
// +build msc3787

// This file contains tests for a join rule which mixes concepts of restricted joins
// and knocking. This is currently experimental and defined by MSC3787, found here:
// https://github.com/matrix-org/matrix-spec-proposals/pull/3787
//
// Generally, this is a combination of knocking_test and restricted_rooms_test.

package tests

import (
	"testing"
	"time"

	"github.com/matrix-org/gomatrixserverlib"
	"github.com/tidwall/gjson"

	"github.com/matrix-org/complement/internal/b"
	"github.com/matrix-org/complement/internal/client"
	"github.com/matrix-org/complement/internal/federation"
	"github.com/matrix-org/complement/internal/match"
	"github.com/matrix-org/complement/internal/must"
	"github.com/matrix-org/complement/runtime"
)

var (
	msc3787RoomVersion = "org.matrix.msc3787"
	msc3787JoinRule    = "knock_restricted"
)

func TestMSC3787(t *testing.T) {
	//deployment := Deploy(t, b.BlueprintFederationTwoLocalOneRemote)
	// fed_2L1R has:
	//  alice:hs1
	//  bob:hs1
	//  charlie:hs2
	// fed_3hs has:
	//  alice:hs1   bob:hs1
	//  charlie:hs2 elise:hs2
	//  george:hs3  helen:hs3
	// server has:
	//  david
	deployment := Deploy(t, b.BlueprintThreeHomeserversTwoUsersEach)
	defer deployment.Destroy(t)

	t.Run("parallel", func(t *testing.T) {
		// uses BlueprintFederationTwoLocalOneRemote
		// with users: alice:hs1, bob:hs1, charlie:hs2

		t.Run("CannotSendKnockViaSendKnockInMSC3787Room", func(t *testing.T) {
			ParallelIfNotGithub(t)
			// See TestCannotSendKnockViaSendKnock(in federation_room_join_test.go)
			// TestCannotSendKnockViaSendKnockInMSC3787Room
			// with users: alice and fake user charlie
			testValidationForSendMembershipEndpoint(t, "/_matrix/federation/v1/send_knock", "knock",
				map[string]interface{}{
					"preset":       "public_chat",
					"room_version": msc3787RoomVersion,
				},
				deployment,
			)
		})
		t.Run("KnockingInMSC3787Room", func(t *testing.T) {
			ParallelIfNotGithub(t)
			// TestKnockingInMSC3787Room
			// See TestKnocking(in knocking_test.go)
			// with users: alice, bob, charlie and fake user david
			//doTestKnocking(t, msc3787RoomVersion, msc3787JoinRule)
			roomVersion := msc3787RoomVersion
			joinRule := msc3787JoinRule
			// Create a client for one local user
			alice := deployment.Client(t, "hs1", "@alice:hs1")

			// Create a client for another local user
			bob := deployment.Client(t, "hs1", "@bob:hs1")

			// Create a client for a remote user
			charlie := deployment.Client(t, "hs2", "@charlie:hs2")

			// Create a server to observe
			inviteWaiter := NewWaiter()
			srv := federation.NewServer(t, deployment,
				federation.HandleKeyRequests(),
				federation.HandleInviteRequests(func(ev *gomatrixserverlib.Event) {
					inviteWaiter.Finish()
				}),
				federation.HandleTransactionRequests(nil, nil),
			)
			cancel := srv.Listen()
			defer cancel()
			srv.UnexpectedRequestsAreErrors = false
			david := srv.UserID("david")

			// Create a room for alice and bob to test knocking with
			roomIDOne := alice.CreateRoom(t, struct {
				Preset      string `json:"preset"`
				RoomVersion string `json:"room_version"`
			}{
				"private_chat", // Set to private in order to get an invite-only room
				roomVersion,
			})
			alice.InviteRoom(t, roomIDOne, david)
			inviteWaiter.Wait(t, 5*time.Second)
			serverRoomOne := srv.MustJoinRoom(t, deployment, "hs1", roomIDOne, david)

			// Test knocking between two users on the same homeserver
			knockingBetweenTwoUsersTest(t, roomIDOne, alice, bob, serverRoomOne, false, joinRule)

			// Create a room for alice and charlie to test knocking with
			roomIDTwo := alice.CreateRoom(t, struct {
				Preset      string `json:"preset"`
				RoomVersion string `json:"room_version"`
			}{
				"private_chat", // Set to private in order to get an invite-only room
				roomVersion,
			})
			inviteWaiter = NewWaiter()
			alice.InviteRoom(t, roomIDTwo, david)
			inviteWaiter.Wait(t, 5*time.Second)
			serverRoomTwo := srv.MustJoinRoom(t, deployment, "hs1", roomIDTwo, david)

			// Test knocking between two users, each on a separate homeserver
			knockingBetweenTwoUsersTest(t, roomIDTwo, alice, charlie, serverRoomTwo, true, joinRule)


		})
		t.Run("KnockRoomsInPublicRoomsDirectoryInMSC3787Room", func(t *testing.T) {
			ParallelIfNotGithub(t)
			// See TestKnockRoomsInPublicRoomsDirectory(in knocking_test.go)
			// TestKnockRoomsInPublicRoomsDirectoryInMSC3787Room
			// with users: alice
			doTestKnockRoomsInPublicRoomsDirectory(t, msc3787RoomVersion, msc3787JoinRule, deployment)
		})
		t.Run("RestrictedRoomsRemoteJoinLocalUserInMSC3787Room", func(t *testing.T) {
			// TestRestrictedRoomsRemoteJoinLocalUserInMSC3787Room
			// See TestRestrictedRoomsRemoteJoinLocalUser(in restricted_rooms_test.go)
			// with users: alice, bob, charlie
			//doTestRestrictedRoomsRemoteJoinLocalUser(t, msc3787RoomVersion, msc3787JoinRule)

			runtime.SkipIf(t, runtime.Dendrite) // FIXME: https://github.com/matrix-org/dendrite/issues/2801
			ParallelIfNotGithub(t)
			roomVersion := msc3787RoomVersion
			joinRule := msc3787JoinRule

			// Charlie sets up the allowed room so it is on the other server.
			//
			// This is the room which membership checks are delegated to. In practice,
			// this will often be an MSC1772 space, but that is not required.
			charlie := deployment.Client(t, "hs2", "@charlie:hs2")
			allowed_room := charlie.CreateRoom(t, map[string]interface{}{
				"preset": "public_chat",
				"name":   "Space",
			})
			// The room is room version 8 which supports the restricted join_rule.
			room := charlie.CreateRoom(t, map[string]interface{}{
				"preset":       "public_chat",
				"name":         "Room",
				"room_version": roomVersion,
				"initial_state": []map[string]interface{}{
					{
						"type":      "m.room.join_rules",
						"state_key": "",
						"content": map[string]interface{}{
							"join_rule": joinRule,
							"allow": []map[string]interface{}{
								{
									"type":    "m.room_membership",
									"room_id": &allowed_room,
									"via":     []string{"hs2"},
								},
							},
						},
					},
				},
			})

			// Invite alice manually and accept it.
			alice := deployment.Client(t, "hs1", "@alice:hs1")
			charlie.InviteRoom(t, room, alice.UserID)
			alice.JoinRoom(t, room, []string{"hs2"})

			// Confirm that Alice cannot issue invites (due to the default power levels).
			bob := deployment.Client(t, "hs1", "@bob:hs1")
			body := map[string]interface{}{
				"user_id": bob.UserID,
			}
			res := alice.DoFunc(
				t,
				"POST",
				[]string{"_matrix", "client", "v3", "rooms", room, "invite"},
				client.WithJSONBody(t, body),
			)
			must.MatchResponse(t, res, match.HTTPResponse{
				StatusCode: 403,
			})

			// Bob cannot join the room.
			failJoinRoom(t, bob, room, "hs1")

			// Join the allowed room via hs2.
			bob.JoinRoom(t, allowed_room, []string{"hs2"})
			// Joining the room should work, although we're joining via hs1, it will end up
			// as a remote join through hs2.
			bob.JoinRoom(t, room, []string{"hs1"})

			// Ensure that the join comes down sync on hs2. Note that we want to ensure hs2
			// accepted the event.
			charlie.MustSyncUntil(t, client.SyncReq{}, client.SyncTimelineHas(
				room,
				func(ev gjson.Result) bool {
					if ev.Get("type").Str != "m.room.member" || ev.Get("state_key").Str != bob.UserID {
						return false
					}
					must.EqualStr(t, ev.Get("sender").Str, bob.UserID, "Bob should have joined by himself")
					must.EqualStr(t, ev.Get("content").Get("membership").Str, "join", "Bob failed to join the room")

					return true
				},
			))

			// Raise the power level so that users on hs1 can invite people and then leave
			// the room.
			state_key := ""
			charlie.SendEventSynced(t, room, b.Event{
				Type:     "m.room.power_levels",
				StateKey: &state_key,
				Content: map[string]interface{}{
					"invite": 0,
					"users": map[string]interface{}{
						charlie.UserID: 100,
					},
				},
			})
			charlie.LeaveRoom(t, room)

			// Ensure the events have synced to hs1.
			alice.MustSyncUntil(t, client.SyncReq{}, client.SyncTimelineHas(
				room,
				func(ev gjson.Result) bool {
					if ev.Get("type").Str != "m.room.member" || ev.Get("state_key").Str != charlie.UserID {
						return false
					}
					must.EqualStr(t, ev.Get("content").Get("membership").Str, "leave", "Charlie failed to leave the room")

					return true
				},
			))

			// Have bob leave and rejoin. This should still work even though hs2 isn't in
			// the room anymore!
			bob.LeaveRoom(t, room)
			bob.JoinRoom(t, room, []string{"hs1"})

		})
		t.Run("RestrictedRoomsLocalJoinInMSC3787Room", func(t *testing.T) {
			ParallelIfNotGithub(t)
			// See TestRestrictedRoomsLocalJoin
			// TestRestrictedRoomsLocalJoinInMSC3787Room

			// Setup the user, allowed room, and restricted room.
			alice, allowed_room, room := setupRestrictedRoom(t, deployment, msc3787RoomVersion, msc3787JoinRule)

			// Create a second user on the same homeserver.
			bob := deployment.Client(t, "hs1", "@bob:hs1")

			// Execute the checks.
			checkRestrictedRoom(t, alice, bob, allowed_room, room, msc3787JoinRule)
		})
		t.Run("RestrictedRoomsRemoteJoinInMSC3787Room", func(t *testing.T) {
			ParallelIfNotGithub(t)
			// See TestRestrictedRoomsRemoteJoin
			// TestRestrictedRoomsRemoteJoinInMSC3787Room

			// Setup the user, allowed room, and restricted room.
			alice, allowed_room, room := setupRestrictedRoom(t, deployment, msc3787RoomVersion, msc3787JoinRule)

			// Create a second user on a different homeserver.
			charlie := deployment.Client(t, "hs2", "@charlie:hs2")

			// Execute the checks.
			checkRestrictedRoom(t, alice, charlie, allowed_room, room, msc3787JoinRule)
		})
	})
	t.Run("RestrictedRoomsRemoteJoinFailOverInMSC3787Room", func(t *testing.T) {
		// See TestRestrictedRoomsRemoteJoinFailOver
		// TestRestrictedRoomsRemoteJoinFailOverInMSC3787Room
		// uses custom Blueprint with 3 homeservers
		// with users: alice:hs1, bob:hs2, charlie:hs3

		//  alice:hs1   bob:hs1
		//  charlie:hs2 elise:hs2
		//  george:hs3  helen:hs3

		runtime.SkipIf(t, runtime.Dendrite) // FIXME: https://github.com/matrix-org/dendrite/issues/2801
		doTestRestrictedRoomsRemoteJoinFailOver(t, msc3787RoomVersion, msc3787JoinRule, deployment)

	})
}
