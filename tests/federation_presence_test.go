package tests

import (
	"testing"

	"github.com/tidwall/gjson"

	"github.com/matrix-org/complement/internal/b"
	"github.com/matrix-org/complement/internal/client"
)

func TestRemotePresence(t *testing.T) {
	deployment := Deploy(t, b.BlueprintFederationOneToOneRoom)
	defer deployment.Destroy(t)

	alicePassword := "alice_secret_password"
	bobPassword := "bob_lame_password"

	// sytest: Presence changes are also reported to remote room members
	t.Run("Presence changes are also reported to remote room members", func(t *testing.T) {
		//alice := deployment.Client(t, "hs1", "@alice:hs1")
		//bob := deployment.Client(t, "hs2", "@bob:hs2")
		alice := deployment.RegisterUser(t, "hs1", "testRemotePresence01Alice", alicePassword, false)
		bob := deployment.RegisterUser(t, "hs2", "testRemotePresence01Bob", bobPassword, false)

		roomID := alice.CreateRoom(t, map[string]interface{}{
			"preset":     "public_chat",		})

		bob.JoinRoom(t, roomID, []string{"hs1"})

		_, bobSinceToken := bob.MustSync(t, client.SyncReq{TimeoutMillis: "0"})

		statusMsg := "Update for room members"
		alice.MustDoFunc(t, "PUT", []string{"_matrix", "client", "v3", "presence", alice.UserID, "status"},
			client.WithJSONBody(t, map[string]interface{}{
				"status_msg": statusMsg,
				"presence":   "online",
			}),
		)

		bob.MustSyncUntil(t, client.SyncReq{Since: bobSinceToken},
			client.SyncPresenceHas(alice.UserID, b.Ptr("online"), func(ev gjson.Result) bool {
				return ev.Get("content.status_msg").Str == statusMsg
			}),
		)
	})
	// sytest: Presence changes to UNAVAILABLE are reported to remote room members
	t.Run("Presence changes to UNAVAILABLE are reported to remote room members", func(t *testing.T) {
		//alice := deployment.Client(t, "hs1", "@alice:hs1")
		//bob := deployment.Client(t, "hs2", "@bob:hs2")
		alice := deployment.RegisterUser(t, "hs1", "testRemotePresence02Alice", alicePassword, false)
		bob := deployment.RegisterUser(t, "hs2", "testRemotePresence02Bob", bobPassword, false)

		roomID := alice.CreateRoom(t, map[string]interface{}{
			"preset":     "public_chat",		})

		bob.JoinRoom(t, roomID, []string{"hs1"})

		_, bobSinceToken := bob.MustSync(t, client.SyncReq{TimeoutMillis: "0"})

		alice.MustDoFunc(t, "PUT", []string{"_matrix", "client", "v3", "presence", alice.UserID, "status"},
			client.WithJSONBody(t, map[string]interface{}{
				"presence": "unavailable",
			}),
		)

		bob.MustSyncUntil(t, client.SyncReq{Since: bobSinceToken},
			client.SyncPresenceHas(alice.UserID, b.Ptr("unavailable")),
		)
	})
}
