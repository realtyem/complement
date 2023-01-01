package b

// BlueprintFederationThreeHomeserversTwoUsersEach is a set of three two-user homeservers federating with each other.
// Users: @alice:hs1    @bob:hs1
//        @charlie:hs2  @elsie:hs2
//        @george:hs3   @helen:hs3
var BlueprintFederationThreeHomeserversTwoUsersEach = MustValidate(Blueprint{
	Name: "federation_three_homeservers",
	Homeservers: []Homeserver{
		{
			Name: "hs1",
			Users: []User{
				{
					Localpart:   "@alice",
					DisplayName: "Alice",
				},
				{
					Localpart:   "@bob",
					DisplayName: "Bob",
				},
			},
		},
		{
			Name: "hs2",
			Users: []User{
				{
					Localpart:   "@charlie",
					DisplayName: "Charlie",
				},
				{
					Localpart:   "@elsie",
					DisplayName: "Elsie",
				},
			},
		},
		{
			Name: "hs3",
			Users: []User{
				{
					Localpart:   "@george",
					DisplayName: "George",
				},
				{
					Localpart:   "@helen",
					DisplayName: "Helen",
				},
			},
		},
	},
})
