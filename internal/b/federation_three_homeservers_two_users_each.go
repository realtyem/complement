package b

// BlueprintThreeHomeserversTwoUsersEach is a set of 3 homeservers with 2 users each.
var BlueprintThreeHomeserversTwoUsersEach = MustValidate(Blueprint{
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
