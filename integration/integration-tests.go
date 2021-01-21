package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/nmasse-itix/keycloak-client"
	"github.com/spf13/pflag"
)

const (
	tstRealm    = "__internal"
	importRealm = "__test_import"
)

func main() {
	var conf = getKeycloakConfig()
	fmt.Printf("Connecting to KC %s...\n", conf.AddrAPI)
	var client, err = keycloak.NewClient(*conf)
	if err != nil {
		log.Fatalf("could not create keycloak client: %v", err)
	}

	// Get access token
	accessToken, err := client.GetToken("master", "admin", "admin")
	if err != nil {
		log.Fatalf("could not get access token: %v", err)
	}

	// Delete realms
	client.DeleteRealm(accessToken, tstRealm)
	client.DeleteRealm(accessToken, importRealm)

	// Check existing realms
	var initialRealms []keycloak.RealmRepresentation
	{
		var err error
		initialRealms, err = client.GetRealms(accessToken)
		if err != nil {
			log.Fatalf("could not get realms: %v", err)
		}
		for _, r := range initialRealms {
			if *r.Realm == tstRealm {
				log.Fatalf("test realm should not exists yet")
			}
		}
	}

	// Create test realm.
	{
		var realm = tstRealm
		var err error
		_, err = client.CreateRealm(accessToken, keycloak.RealmRepresentation{
			Realm: &realm,
		})
		if err != nil {
			log.Fatalf("could not create keycloak client: %v", err)
		}
		fmt.Println("Test realm created.")
	}

	// Check getRealm.
	{
		var realmR, err = client.GetRealm(accessToken, tstRealm)
		if err != nil {
			log.Fatalf("could not get test realm: %v", err)
		}
		if *realmR.Realm != tstRealm {
			log.Fatalf("test realm has wrong name")
		}
		if realmR.DisplayName != nil {
			log.Fatalf("test realm should not have a field displayName")
		}
		fmt.Println("Test realm exists.")
	}

	// Update Realm
	{
		var displayName = "updated realm"
		var err = client.UpdateRealm(accessToken, tstRealm, keycloak.RealmRepresentation{
			DisplayName: &displayName,
		})
		if err != nil {
			log.Fatalf("could not update test realm: %v", err)
		}
		// Check update
		{
			var realmR, err = client.GetRealm(accessToken, tstRealm)
			if err != nil {
				log.Fatalf("could not get test realm: %v", err)
			}
			if *realmR.DisplayName != displayName {
				log.Fatalf("test realm update failed")
			}
		}
		fmt.Println("Test realm updated.")
	}

	// Count users.
	{
		var nbrUser, err = client.CountUsers(accessToken, tstRealm)
		if err != nil {
			log.Fatalf("could not count users: %v", err)
		}
		if nbrUser != 0 {
			log.Fatalf("there should be 0 users")
		}
	}

	// Create test users.
	{
		for _, u := range tstUsers {
			var username = strings.ToLower(u.firstname + "." + u.lastname)
			var email = username + "@cloudtrust.ch"
			var err error
			_, err = client.CreateUser(accessToken, tstRealm, keycloak.UserRepresentation{
				Username:  &username,
				FirstName: &u.firstname,
				LastName:  &u.lastname,
				Email:     &email,
			})
			if err != nil {
				log.Fatalf("could not create test users: %v", err)
			}

		}
		// Check that all users where created.
		{
			var nbrUser, err = client.CountUsers(accessToken, tstRealm)
			if err != nil {
				log.Fatalf("could not count users: %v", err)
			}
			if nbrUser != 50 {
				log.Fatalf("there should be 50 users")
			}
		}
		fmt.Println("Test users created.")
	}

	// Get users
	{
		{
			// No parameters.
			var users, err = client.GetUsers(accessToken, tstRealm)
			if err != nil {
				log.Fatalf("could not get users: %v", err)
			}
			if len(users) != 50 {
				log.Fatalf("there should be 50 users")
			}

			user, err := client.GetUser(accessToken, tstRealm, *(users[0].ID))
			if err != nil {
				log.Fatalf("could not get user")
			}

			if !(*(user.Username) != "") {
				log.Fatalf("Username should not be empty")
			}

			fmt.Println("Test user retrieved.")
		}
		{
			// email.
			var users, err = client.GetUsers(accessToken, tstRealm, "email", "john.doe@cloudtrust.ch")
			if err != nil {
				log.Fatalf("could not get users: %v", err)
			}
			if len(users) != 1 {
				log.Fatalf("there should be 1 user matched by email")
			}
		}
		{
			// firstname.
			var users, err = client.GetUsers(accessToken, tstRealm, "firstName", "John")
			if err != nil {
				log.Fatalf("could not get users: %v", err)
			}
			// Match John and Johnny
			if len(users) != 2 {
				log.Fatalf("there should be 2 user matched by firstname")
			}
		}
		{
			// lastname.
			var users, err = client.GetUsers(accessToken, tstRealm, "lastName", "Wells")
			if err != nil {
				log.Fatalf("could not get users: %v", err)
			}
			if len(users) != 3 {
				log.Fatalf("there should be 3 users matched by lastname")
			}
		}
		{
			// username.
			var users, err = client.GetUsers(accessToken, tstRealm, "username", "lucia.nelson")
			if err != nil {
				log.Fatalf("could not get users: %v", err)
			}
			if len(users) != 1 {
				log.Fatalf("there should be 1 user matched by username")
			}
		}
		{
			// first.
			var users, err = client.GetUsers(accessToken, tstRealm, "max", "7")
			if err != nil {
				log.Fatalf("could not get users: %v", err)
			}
			if len(users) != 7 {
				log.Fatalf("there should be 7 users matched by max")
			}
		}
		{
			// search.
			var users, err = client.GetUsers(accessToken, tstRealm, "search", "le")
			if err != nil {
				log.Fatalf("could not get users: %v", err)
			}
			if len(users) != 7 {
				log.Fatalf("there should be 7 users matched by search")
			}
		}

		fmt.Println("Test users retrieved.")
	}

	// Update user.
	{
		// Get user ID.
		var userID string
		{
			var users, err = client.GetUsers(accessToken, tstRealm, "search", "Maria")
			if err != nil {
				log.Fatalf("could not get Maria: %v", err)
			}
			if len(users) != 1 {
				log.Fatalf("there should be 1 users matched by search Maria")
			}
			if users[0].ID == nil {
				log.Fatalf("user ID should not be nil")
			}
			userID = *users[0].ID
		}
		// Update user.
		var username = "Maria"
		var updatedLastname = "updated"
		{

			var err = client.UpdateUser(accessToken, tstRealm, userID, keycloak.UserRepresentation{
				FirstName: &username,
				LastName:  &updatedLastname,
			})
			if err != nil {
				log.Fatalf("could not update user: %v", err)
			}
		}
		// Check that user was updated.
		{
			var users, err = client.GetUsers(accessToken, tstRealm, "search", "Maria")
			if err != nil {
				log.Fatalf("could not get Maria: %v", err)
			}
			if len(users) != 1 {
				log.Fatalf("there should be 1 users matched by search Maria")
			}
			if users[0].LastName == nil || *users[0].LastName != updatedLastname {
				log.Fatalf("user was not updated")
			}
		}
		fmt.Println("User updated.")
		// Check credentials
		{
			var creds, err = client.GetCredentials(accessToken, tstRealm, userID)
			if err != nil {
				log.Fatalf("could not get credentials: %v", err)
			}
			if len(creds) != 0 {
				log.Fatalf("Maria should not have credentials")
			}
		}
	}

	// Delete user.
	{
		// Get user ID.
		var userID string
		{
			var users, err = client.GetUsers(accessToken, tstRealm, "search", "Toni")
			if err != nil {
				log.Fatalf("could not get Toni: %v", err)
			}
			if len(users) != 1 {
				log.Fatalf("there should be 1 users matched by search Toni")
			}
			if users[0].ID == nil {
				log.Fatalf("user ID should not be nil")
			}
			userID = *users[0].ID
		}
		// Delete user.
		{
			var err = client.DeleteUser(accessToken, tstRealm, userID)
			if err != nil {
				log.Fatalf("could not delete user: %v", err)
			}
		}
		// Check that user was deleted.
		{
			var nbrUser, err = client.CountUsers(accessToken, tstRealm)
			if err != nil {
				log.Fatalf("could not count users: %v", err)
			}
			if nbrUser != 49 {
				log.Fatalf("there should be 49 users")
			}
		}
		fmt.Println("User deleted.")
	}

	// Create component
	{
		var c, err = client.CreateComponent(accessToken, tstRealm, keycloak.ComponentRepresentation{
			Name:         &ldapProviderName,
			ProviderType: &ldapProviderType,
			ProviderID:   &ldapProviderName,
			Config:       &ldapProviderConfig,
		})
		if err != nil {
			log.Fatalf("could not create ldap component: %v", err)
		}
		u, err := url.Parse(c)
		if err != nil {
			log.Fatalf("cannot : %v", err)
		}
		slugs := strings.Split(u.Path, "/")
		parentID := slugs[len(slugs)-1]

		for _, m := range ldapMapperAttrs {
			var err error
			config := keycloak.MultivaluedHashMap{
				"ldap.attribute":              {m.ldapAttribute},
				"is.mandatory.in.ldap":        {m.mandatory},
				"always.read.value.from.ldap": {m.alwaysRead},
				"read.only":                   {m.readonly},
				"user.model.attribute":        {m.modelAttribute},
			}
			_, err = client.CreateComponent(accessToken, tstRealm, keycloak.ComponentRepresentation{
				Name:         &m.name,
				ProviderType: &ldapMapperType,
				ProviderID:   &m.providerID,
				Config:       &config,
				ParentID:     &parentID,
			})
			if err != nil {
				log.Fatalf("could not create ldap mapper component: %v", err)
			}
		}
		fmt.Println("Components created.")
	}

	// Delete test realm.
	{
		var err = client.DeleteRealm(accessToken, tstRealm)
		if err != nil {
			log.Fatalf("could not delete test realm: %v", err)
		}
		// Check that the realm was deleted.
		{
			var realms, err = client.GetRealms(accessToken)
			if err != nil {
				log.Fatalf("could not get realms: %v", err)
			}
			for _, r := range realms {
				if *r.Realm == tstRealm {
					log.Fatalf("test realm should be deleted")
				}
			}
		}
		fmt.Println("Test realm deleted.")
	}

	// Realm Import
	{
		var err error
		var realm keycloak.RealmRepresentation
		err = json.Unmarshal([]byte(realmContent), &realm)
		if err != nil {
			log.Fatalf("cannot read realm JSON: %s", err)
		}

		_, err = client.CreateRealm(accessToken, realm)
		if err != nil {
			log.Fatalf("could not create keycloak realm: %v", err)
		}

		components, err := client.GetComponents(accessToken, importRealm)
		count := make(map[string]int)
		for _, component := range components {
			if component.ProviderType != nil {
				count[*component.ProviderType]++
			}
		}
		if count["org.keycloak.storage.UserStorageProvider"] != 1 {
			log.Fatalf("wrong number of ldap components: %d", count["org.keycloak.storage.UserStorageProvider"])
		}
		if count["org.keycloak.storage.ldap.mappers.LDAPStorageMapper"] != 6 {
			log.Fatalf("wrong number of ldap mapper components: %d", count["org.keycloak.storage.ldap.mappers.LDAPStorageMapper"])
		}

		fmt.Println("Test realm imported.")
	}

	// Delete imported realm.
	{
		var err = client.DeleteRealm(accessToken, importRealm)
		if err != nil {
			log.Fatalf("could not delete imported realm: %v", err)
		}
		fmt.Println("Test realm deleted.")
	}

}

func getKeycloakConfig() *keycloak.Config {
	var apiAddr = pflag.String("urlKc", "http://localhost:8080/auth", "keycloak address")
	var tokenAddr = pflag.String("url", "http://localhost:8080/auth/realms/master", "token address")
	pflag.Parse()

	return &keycloak.Config{
		AddrTokenProvider: *tokenAddr,
		AddrAPI:           *apiAddr,
		Timeout:           10 * time.Second,
	}
}

var tstUsers = []struct {
	firstname string
	lastname  string
}{
	{"John", "Doe"},
	{"Johnny", "Briggs"},
	{"Karen", "Sutton"},
	{"Cesar", "Mathis"},
	{"Ryan", "Kennedy"},
	{"Kent", "Phillips"},
	{"Loretta", "Curtis"},
	{"Derrick", "Cox"},
	{"Greg", "Wilkins"},
	{"Andy", "Reynolds"},
	{"Toni", "Meyer"},
	{"Joyce", "Sullivan"},
	{"Johanna", "Wells"},
	{"Judith", "Barnett"},
	{"Joanne", "Ward"},
	{"Bethany", "Johnson"},
	{"Maria", "Murphy"},
	{"Mattie", "Quinn"},
	{"Erick", "Robbins"},
	{"Beulah", "Greer"},
	{"Patty", "Wong"},
	{"Gayle", "Garrett"},
	{"Stewart", "Floyd"},
	{"Wilbur", "Schneider"},
	{"Diana", "Logan"},
	{"Eduardo", "Mitchell"},
	{"Lela", "Wells"},
	{"Homer", "Miles"},
	{"Audrey", "Park"},
	{"Rebecca", "Fuller"},
	{"Jeremiah", "Andrews"},
	{"Cedric", "Reyes"},
	{"Lee", "Griffin"},
	{"Ebony", "Knight"},
	{"Gilbert", "Franklin"},
	{"Jessie", "Norman"},
	{"Cary", "Wells"},
	{"Arlene", "James"},
	{"Jerry", "Chavez"},
	{"Marco", "Weber"},
	{"Celia", "Guerrero"},
	{"Faye", "Massey"},
	{"Jorge", "Mccarthy"},
	{"Jennifer", "Colon"},
	{"Angel", "Jordan"},
	{"Bennie", "Hubbard"},
	{"Terrance", "Norris"},
	{"May", "Sharp"},
	{"Glenda", "Hogan"},
	{"Lucia", "Nelson"},
}

var ldapProviderConfig keycloak.MultivaluedHashMap = keycloak.MultivaluedHashMap{
	"pagination":                           []string{"true"},
	"fullSyncPeriod":                       []string{"-1"},
	"usersDn":                              []string{"ou=users,dc=keycloak,dc=org"},
	"connectionPooling":                    []string{"true"},
	"cachePolicy":                          []string{"DEFAULT"},
	"useKerberosForPasswordAuthentication": []string{"false"},
	"importEnabled":                        []string{"true"},
	"enabled":                              []string{"true"},
	"bindCredential":                       []string{"keycloak"},
	"bindDn":                               []string{"cn=admin,dc=keycloak,dc=org"},
	"changedSyncPeriod":                    []string{"-1"},
	"usernameLDAPAttribute":                []string{"uid"},
	"vendor":                               []string{"other"},
	"uuidLDAPAttribute":                    []string{"entryUUID"},
	"connectionUrl":                        []string{"ldap://openldap.dns.podman:389/"},
	"allowKerberosAuthentication":          []string{"false"},
	"syncRegistrations":                    []string{"false"},
	"authType":                             []string{"simple"},
	"debug":                                []string{"false"},
	"searchScope":                          []string{"1"},
	"useTruststoreSpi":                     []string{"ldapsOnly"},
	"priority":                             []string{"0"},
	"trustEmail":                           []string{"true"},
	"userObjectClasses":                    []string{"inetOrgPerson, organizationalPerson"},
	"rdnLDAPAttribute":                     []string{"uid"},
	"editMode":                             []string{"READ_ONLY"},
	"validatePasswordPolicy":               []string{"false"},
	"batchSizeForSync":                     []string{"1000"},
}
var ldapProviderName string = "ldap"
var ldapProviderType string = "org.keycloak.storage.UserStorageProvider"

var ldapMapperType string = "org.keycloak.storage.ldap.mappers.LDAPStorageMapper"
var ldapMapperAttrs = []struct {
	name           string
	providerID     string
	ldapAttribute  string
	mandatory      string
	alwaysRead     string
	readonly       string
	modelAttribute string
}{
	{"modify date", "user-attribute-ldap-mapper", "modifyTimestamp", "false", "true", "true", "modifyTimestamp"},
	{"username", "user-attribute-ldap-mapper", "uid", "true", "true", "false", "username"},
	{"first name", "user-attribute-ldap-mapper", "cn", "true", "true", "true", "firstName"},
	{"email", "user-attribute-ldap-mapper", "mail", "false", "false", "true", "email"},
	{"last name", "user-attribute-ldap-mapper", "sn", "true", "true", "true", "lastName"},
	{"creation date", "user-attribute-ldap-mapper", "createTimestamp", "false", "true", "true", "createTimestamp"},
}

var realmContent string = `
{
	"id": "__test_import",
	"realm": "__test_import",
	"displayName": "__test_import",
	"notBefore": 0,
	"revokeRefreshToken": false,
	"refreshTokenMaxReuse": 0,
	"accessTokenLifespan": 300,
	"accessTokenLifespanForImplicitFlow": 900,
	"ssoSessionIdleTimeout": 1800,
	"ssoSessionMaxLifespan": 36000,
	"offlineSessionIdleTimeout": 2592000,
	"accessCodeLifespan": 60,
	"accessCodeLifespanUserAction": 300,
	"accessCodeLifespanLogin": 1800,
	"actionTokenGeneratedByAdminLifespan": 43200,
	"actionTokenGeneratedByUserLifespan": 300,
	"enabled": true,
	"sslRequired": "external",
	"registrationAllowed": false,
	"registrationEmailAsUsername": false,
	"rememberMe": false,
	"verifyEmail": false,
	"loginWithEmailAllowed": true,
	"duplicateEmailsAllowed": false,
	"resetPasswordAllowed": false,
	"editUsernameAllowed": false,
	"bruteForceProtected": false,
	"permanentLockout": false,
	"maxFailureWaitSeconds": 900,
	"minimumQuickLoginWaitSeconds": 60,
	"waitIncrementSeconds": 60,
	"quickLoginCheckMilliSeconds": 1000,
	"maxDeltaTimeSeconds": 43200,
	"failureFactor": 30,
	"users": [
	],
	"roles": {
	  "realm": [],
	  "client": {}
	},
	"defaultRoles": [],
	"requiredCredentials": [ "password" ],
	"scopeMappings": [],
	"clientScopeMappings": {},
	"clients": [
	  {
		"clientId": "app_000000",
		"name": "app_000000",
		"enabled": true,
		"publicClient": false,
		"redirectUris": [
		  "http://dummy/url"
		],
		"fullScopeAllowed": false,
		"standardFlowEnabled": true,
		"directAccessGrantsEnabled": true,
		"serviceAccountsEnabled": true,
		"clientAuthenticatorType": "client-secret",
		"secret": "26dd19e8-cccf-4783-8e07-95c001209e88"
	  },
	  {
		"clientId": "app_000001",
		"name": "app_000001",
		"enabled": true,
		"publicClient": false,
		"redirectUris": [
		  "http://dummy/url"
		],
		"fullScopeAllowed": false,
		"standardFlowEnabled": true,
		"directAccessGrantsEnabled": true,
		"serviceAccountsEnabled": true,
		"clientAuthenticatorType": "client-secret",
		"secret": "c0e1c9ba-94b5-4345-95cd-e75bd12840ac"
	  },
	  {
		"clientId": "app_000002",
		"name": "app_000002",
		"enabled": true,
		"publicClient": false,
		"redirectUris": [
		  "http://dummy/url"
		],
		"fullScopeAllowed": false,
		"standardFlowEnabled": true,
		"directAccessGrantsEnabled": true,
		"serviceAccountsEnabled": true,
		"clientAuthenticatorType": "client-secret",
		"secret": "cc0d4bc0-5b68-47f1-a4ce-a263a1bc87fa"
	  },
	  {
		"clientId": "app_000003",
		"name": "app_000003",
		"enabled": true,
		"publicClient": false,
		"redirectUris": [
		  "http://dummy/url"
		],
		"fullScopeAllowed": false,
		"standardFlowEnabled": true,
		"directAccessGrantsEnabled": true,
		"serviceAccountsEnabled": true,
		"clientAuthenticatorType": "client-secret",
		"secret": "fd9e46e2-1c07-4894-b33e-f82e0767d306"
	  },
	  {
		"clientId": "app_000004",
		"name": "app_000004",
		"enabled": true,
		"publicClient": false,
		"redirectUris": [
		  "http://dummy/url"
		],
		"fullScopeAllowed": false,
		"standardFlowEnabled": true,
		"directAccessGrantsEnabled": true,
		"serviceAccountsEnabled": true,
		"clientAuthenticatorType": "client-secret",
		"secret": "8258e375-11ea-4623-8666-79e5208b226d"
	  },
	  {
		"clientId": "app_000005",
		"name": "app_000005",
		"enabled": true,
		"publicClient": false,
		"redirectUris": [
		  "http://dummy/url"
		],
		"fullScopeAllowed": false,
		"standardFlowEnabled": true,
		"directAccessGrantsEnabled": true,
		"serviceAccountsEnabled": true,
		"clientAuthenticatorType": "client-secret",
		"secret": "6d32184c-de0f-4c59-a3dc-518b24a63a36"
	  },
	  {
		"clientId": "app_000006",
		"name": "app_000006",
		"enabled": true,
		"publicClient": false,
		"redirectUris": [
		  "http://dummy/url"
		],
		"fullScopeAllowed": false,
		"standardFlowEnabled": true,
		"directAccessGrantsEnabled": true,
		"serviceAccountsEnabled": true,
		"clientAuthenticatorType": "client-secret",
		"secret": "934630e9-513f-4561-bf0d-b43b15ef8d11"
	  },
	  {
		"clientId": "app_000007",
		"name": "app_000007",
		"enabled": true,
		"publicClient": false,
		"redirectUris": [
		  "http://dummy/url"
		],
		"fullScopeAllowed": false,
		"standardFlowEnabled": true,
		"directAccessGrantsEnabled": true,
		"serviceAccountsEnabled": true,
		"clientAuthenticatorType": "client-secret",
		"secret": "9ed0b1f7-d49a-4f19-b448-bc5777e0ba91"
	  },
	  {
		"clientId": "app_000008",
		"name": "app_000008",
		"enabled": true,
		"publicClient": false,
		"redirectUris": [
		  "http://dummy/url"
		],
		"fullScopeAllowed": false,
		"standardFlowEnabled": true,
		"directAccessGrantsEnabled": true,
		"serviceAccountsEnabled": true,
		"clientAuthenticatorType": "client-secret",
		"secret": "464b8054-620c-4226-a726-e60ab267b11b"
	  },
	  {
		"clientId": "app_000009",
		"name": "app_000009",
		"enabled": true,
		"publicClient": false,
		"redirectUris": [
		  "http://dummy/url"
		],
		"fullScopeAllowed": false,
		"standardFlowEnabled": true,
		"directAccessGrantsEnabled": true,
		"serviceAccountsEnabled": true,
		"clientAuthenticatorType": "client-secret",
		"secret": "78313217-1ff8-461a-b944-80bb76c01731"
	  }
	],
	"components": {
	  "org.keycloak.storage.UserStorageProvider": [
		{
		  "id": "84981250-eda0-4758-991e-ea6d79794d44",
		  "name": "ldap",
		  "providerId": "ldap",
		  "subComponents": {
			"org.keycloak.storage.ldap.mappers.LDAPStorageMapper": [
			  {
				"id": "cea411ae-009d-44a7-b964-27f23cb8c753",
				"name": "modify date",
				"providerId": "user-attribute-ldap-mapper",
				"subComponents": {},
				"config": {
				  "ldap.attribute": [
					"modifyTimestamp"
				  ],
				  "is.mandatory.in.ldap": [
					"false"
				  ],
				  "always.read.value.from.ldap": [
					"true"
				  ],
				  "read.only": [
					"true"
				  ],
				  "user.model.attribute": [
					"modifyTimestamp"
				  ]
				}
			  },
			  {
				"id": "ef8a7263-7db3-4131-9ba9-1e92211691b3",
				"name": "username",
				"providerId": "user-attribute-ldap-mapper",
				"subComponents": {},
				"config": {
				  "ldap.attribute": [
					"uid"
				  ],
				  "is.mandatory.in.ldap": [
					"true"
				  ],
				  "read.only": [
					"true"
				  ],
				  "always.read.value.from.ldap": [
					"false"
				  ],
				  "user.model.attribute": [
					"username"
				  ]
				}
			  },
			  {
				"id": "f00c2b5a-be41-4990-bdd0-165dbd47d6c2",
				"name": "first name",
				"providerId": "user-attribute-ldap-mapper",
				"subComponents": {},
				"config": {
				  "ldap.attribute": [
					"cn"
				  ],
				  "is.mandatory.in.ldap": [
					"true"
				  ],
				  "read.only": [
					"true"
				  ],
				  "always.read.value.from.ldap": [
					"true"
				  ],
				  "user.model.attribute": [
					"firstName"
				  ]
				}
			  },
			  {
				"id": "18349fe8-c8ff-4a1a-a588-8b0a1f13b880",
				"name": "email",
				"providerId": "user-attribute-ldap-mapper",
				"subComponents": {},
				"config": {
				  "ldap.attribute": [
					"mail"
				  ],
				  "is.mandatory.in.ldap": [
					"false"
				  ],
				  "always.read.value.from.ldap": [
					"false"
				  ],
				  "read.only": [
					"true"
				  ],
				  "user.model.attribute": [
					"email"
				  ]
				}
			  },
			  {
				"id": "0f59a566-2306-4df4-ad54-2d0c33dc2267",
				"name": "last name",
				"providerId": "user-attribute-ldap-mapper",
				"subComponents": {},
				"config": {
				  "ldap.attribute": [
					"sn"
				  ],
				  "is.mandatory.in.ldap": [
					"true"
				  ],
				  "always.read.value.from.ldap": [
					"true"
				  ],
				  "read.only": [
					"true"
				  ],
				  "user.model.attribute": [
					"lastName"
				  ]
				}
			  },
			  {
				"id": "656aa6c4-fc6c-4d2e-a61d-246872ad45f3",
				"name": "creation date",
				"providerId": "user-attribute-ldap-mapper",
				"subComponents": {},
				"config": {
				  "ldap.attribute": [
					"createTimestamp"
				  ],
				  "is.mandatory.in.ldap": [
					"false"
				  ],
				  "always.read.value.from.ldap": [
					"true"
				  ],
				  "read.only": [
					"true"
				  ],
				  "user.model.attribute": [
					"createTimestamp"
				  ]
				}
			  }
			]
		  },
		  "config": {
			"pagination": [
			  "true"
			],
			"fullSyncPeriod": [
			  "-1"
			],
			"usersDn": [
			  "ou=users,dc=keycloak,dc=org"
			],
			"connectionPooling": [
			  "true"
			],
			"cachePolicy": [
			  "DEFAULT"
			],
			"useKerberosForPasswordAuthentication": [
			  "false"
			],
			"importEnabled": [
			  "true"
			],
			"enabled": [
			  "true"
			],
			"bindCredential": [
			  "keycloak"
			],
			"bindDn": [
			  "cn=admin,dc=keycloak,dc=org"
			],
			"changedSyncPeriod": [
			  "-1"
			],
			"usernameLDAPAttribute": [
			  "uid"
			],
			"lastSync": [
			  "1611161804"
			],
			"vendor": [
			  "other"
			],
			"uuidLDAPAttribute": [
			  "entryUUID"
			],
			"connectionUrl": [
			  "ldap://openldap.dns.podman:389/"
			],
			"allowKerberosAuthentication": [
			  "false"
			],
			"syncRegistrations": [
			  "false"
			],
			"authType": [
			  "simple"
			],
			"debug": [
			  "false"
			],
			"searchScope": [
			  "1"
			],
			"useTruststoreSpi": [
			  "ldapsOnly"
			],
			"priority": [
			  "0"
			],
			"trustEmail": [
			  "true"
			],
			"userObjectClasses": [
			  "inetOrgPerson, organizationalPerson"
			],
			"rdnLDAPAttribute": [
			  "uid"
			],
			"editMode": [
			  "READ_ONLY"
			],
			"validatePasswordPolicy": [
			  "false"
			],
			"batchSizeForSync": [
			  "1000"
			]
		  }
		}
	  ]
	}
}`
