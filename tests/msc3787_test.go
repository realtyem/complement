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

	"github.com/matrix-org/complement/internal/b"
)

var (
	msc3787RoomVersion = "org.matrix.msc3787"
	msc3787JoinRule    = "knock_restricted"
)

func TestMSC3787(t *testing.T) {
	deployment := Deploy(t, b.BlueprintFederationThreeHomeserversTwoUsersEach)
	defer deployment.Destroy(t)

	t.Run("knocking", func(t *testing.T) {
		// See TestKnocking
		// needs users: alice:hs1, bob:hs1, charlie:hs2
		doTestKnocking(t, msc3787RoomVersion, msc3787JoinRule, deployment)
	})
	t.Run("knockRoomsInPublicRoomsDirectory", func(t *testing.T) {
		// See TestKnockRoomsInPublicRoomsDirectory
		// needs users: alice:hs1
		doTestKnockRoomsInPublicRoomsDirectory(t, msc3787RoomVersion, msc3787JoinRule, deployment)
	})
	t.Run("cannotSendKnockViaSendKnock", func(t *testing.T) {
		// See TestCannotSendKnockViaSendKnock
		// needs user: alice:hs1
		testValidationForSendMembershipEndpoint(t, "/_matrix/federation/v1/send_knock", "knock",
			map[string]interface{}{
				"preset":       "public_chat",
				"room_version": msc3787RoomVersion,
			},
			deployment,
		)
	})
	t.Run("restrictedRoomsLocalJoin", func(t *testing.T) {
		// See TestRestrictedRoomsLocalJoin
		// need users: alice:hs1, bob:hs1
		// Setup the user, allowed room, and restricted room.
		alice, allowed_room, room := setupRestrictedRoom(t, deployment, msc3787RoomVersion, msc3787JoinRule)

		// Create a second user on the same homeserver.
		bob := deployment.Client(t, "hs1", "@bob:hs1")

		// Execute the checks.
		checkRestrictedRoom(t, alice, bob, allowed_room, room, msc3787JoinRule)
	})
	t.Run("restrictedRoomsRemoteJoin", func(t *testing.T) {
		// See TestRestrictedRoomsRemoteJoin
		// need users: alice:hs1, charlie:hs2
		// Setup the user, allowed room, and restricted room.
		alice, allowed_room, room := setupRestrictedRoom(t, deployment, msc3787RoomVersion, msc3787JoinRule)

		// Create a second user on a different homeserver.
		charlie := deployment.Client(t, "hs2", "@charlie:hs2")

		// Execute the checks.
		checkRestrictedRoom(t, alice, charlie, allowed_room, room, msc3787JoinRule)
	})
	t.Run("restrictedRoomsRemoteJoinLocalUser", func(t *testing.T) {
		// See TestRestrictedRoomsRemoteJoinLocalUser
		// needs users: alice:hs1, bob:hs1, charlie:hs2
		doTestRestrictedRoomsRemoteJoinLocalUser(t, msc3787RoomVersion, msc3787JoinRule, deployment)
	})
	t.Run("restrictedRoomsRemoteJoinFailOver", func(t *testing.T) {
	// See TestRestrictedRoomsRemoteJoinFailOver
		doTestRestrictedRoomsRemoteJoinFailOver(t, msc3787RoomVersion, msc3787JoinRule, deployment)
	})
}
