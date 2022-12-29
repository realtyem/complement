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
	t.Run("KnockingInMSC3787Room", func(t *testing.T) {
		//t.Parallel()
		//TestKnockingInMSC3787Room
		// See TestKnocking
		doTestKnocking(t, msc3787RoomVersion, msc3787JoinRule)
	})
	t.Run("KnockRoomsInPublicRoomsDirectoryInMSC3787Room", func(t *testing.T) {
		//t.Parallel()
		// See TestKnockRoomsInPublicRoomsDirectory
		// TestKnockRoomsInPublicRoomsDirectoryInMSC3787Room
		doTestKnockRoomsInPublicRoomsDirectory(t, msc3787RoomVersion, msc3787JoinRule)
	})
	t.Run("CannotSendKnockViaSendKnockInMSC3787Room", func(t *testing.T) {
		//t.Parallel()
		// See TestCannotSendKnockViaSendKnock
		// TestCannotSendKnockViaSendKnockInMSC3787Room
		testValidationForSendMembershipEndpoint(t, "/_matrix/federation/v1/send_knock", "knock",
			map[string]interface{}{
				"preset":       "public_chat",
				"room_version": msc3787RoomVersion,
			},
		)
	})
	t.Run("RestrictedRoomsLocalJoinInMSC3787Room", func(t *testing.T) {
		t.Parallel()
		// See TestRestrictedRoomsLocalJoin
		// TestRestrictedRoomsLocalJoinInMSC3787Room
		deployment := Deploy(t, b.BlueprintOneToOneRoom)
		defer deployment.Destroy(t)

		// Setup the user, allowed room, and restricted room.
		alice, allowed_room, room := setupRestrictedRoom(t, deployment, msc3787RoomVersion, msc3787JoinRule)

		// Create a second user on the same homeserver.
		bob := deployment.Client(t, "hs1", "@bob:hs1")

		// Execute the checks.
		checkRestrictedRoom(t, alice, bob, allowed_room, room, msc3787JoinRule)
	})
	t.Run("RestrictedRoomsRemoteJoinInMSC3787Room", func(t *testing.T) {
		t.Parallel()
		// See TestRestrictedRoomsRemoteJoin
		// TestRestrictedRoomsRemoteJoinInMSC3787Room
		deployment := Deploy(t, b.BlueprintFederationOneToOneRoom)
		defer deployment.Destroy(t)

		// Setup the user, allowed room, and restricted room.
		alice, allowed_room, room := setupRestrictedRoom(t, deployment, msc3787RoomVersion, msc3787JoinRule)

		// Create a second user on a different homeserver.
		bob := deployment.Client(t, "hs2", "@bob:hs2")

		// Execute the checks.
		checkRestrictedRoom(t, alice, bob, allowed_room, room, msc3787JoinRule)
	})
	t.Run("RestrictedRoomsRemoteJoinLocalUserInMSC3787Room", func(t *testing.T) {
		t.Parallel()
		// See TestRestrictedRoomsRemoteJoinLocalUser
		// TestRestrictedRoomsRemoteJoinLocalUserInMSC3787Room
		doTestRestrictedRoomsRemoteJoinLocalUser(t, msc3787RoomVersion, msc3787JoinRule)
	})
	t.Run("RestrictedRoomsRemoteJoinFailOverInMSC3787Room", func(t *testing.T) {
		t.Parallel()
		// See TestRestrictedRoomsRemoteJoinFailOver
		// TestRestrictedRoomsRemoteJoinFailOverInMSC3787Room
		doTestRestrictedRoomsRemoteJoinFailOver(t, msc3787RoomVersion, msc3787JoinRule)
	})
}
