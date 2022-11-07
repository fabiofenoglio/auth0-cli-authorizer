# auth0-cli-authorizer
implement the Device Authorization Flow on Auth0 to obtain and manage tokens from your CLI application

## Setup

```bash
go get github.com/fabiofenoglio/auth0-cli-authorizer
```

### Basic usage

With the default configuration, a call to Authorize() 
will automatically attempt to open a browser window
and prefill with the user code.

```go
package main

import (
	"context"
	"fmt"
	
	authorizer "github.com/fabiofenoglio/auth0-cli-authorizer"
)

func main() {

	auth, _ := authorizer.New(
		"https://<your-domain>.auth0.com", 
		"yourClientID", 
		"https://<your-audience>",
	)

	authorization, _ := auth.Authorize(context.TODO())

	fmt.Println("welcome " + authorization.User.Name + " !")
	// use authorization.Tokens.AccessToken from now on
}
```

### Refresh token example

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"

	authorizer "github.com/fabiofenoglio/auth0-cli-authorizer"
)

func main() {
	// you can customize many options with authorizer.With[...]
	auth, _ := authorizer.New(
		"https://<your-domain>.auth0.com",
		"yourClientID",
		"https://<your-audience>",
		authorizer.WithRequireOfflineAccess(true), // also enabled by default
	)

	authorization, err := auth.Authorize(context.TODO())
	if err != nil {
		panic(err)
	}
	
	// ...
	// later on, if you store the refresh token and you want
	// to fetch a new access token:
	refreshed, _ := auth.Refresh(context.TODO(), authorization.Tokens.RefreshToken)

	pretty, _ := json.MarshalIndent(refreshed, "", "  ")
	fmt.Println(string(pretty))
}
```

### Caching on local filesystem

If you enable a store option (like `WithAppDataStore`) the token will
automatically be saved and restored from a local cache, optionally
refreshing it as needed.

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	authorizer "github.com/fabiofenoglio/auth0-cli-authorizer"
)

func main() {
	// you can customize many options with authorizer.With[...]
	auth, _ := authorizer.New(
		"https://<your-domain>.auth0.com",
		"yourClientID",
		"https://<your-audience>",
		// will require a refresh_token
		authorizer.WithRequireOfflineAccess(true),
		
		// will cache the token on filesystem
		// and refresh it automatically 
		// if it's expired or expires in less than 5 minutes
		authorizer.WithAppDataStore(5*time.Minute), 
	)

	authorization, err := auth.Authorize(context.TODO())
	if err != nil {
		panic(err)
	}

	// use authorization.Tokens.AccessToken from now on
	// if you run this program again after the first time
	// you will not be prompted for authorization again
	
	pretty, _ := json.MarshalIndent(authorization, "", "  ")
	fmt.Println(string(pretty))
}
```

### Complete example

```go
package main

import (
	"context"
	"encoding/json"
	"time"

	authorizer "github.com/fabiofenoglio/auth0-cli-authorizer"
	"github.com/sirupsen/logrus"
)

func main() {
	// you can pass a custom logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:   false,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		PadLevelText:    true,
	})

	// you can customize many options with authorizer.With[...]
	auth, _ := authorizer.New(
		"https://<your-domain>.auth0.com",
		"yourClientID",
		"https://<your-audience>",
		authorizer.WithLogger(logger),
		
		// will require a refresh_token
		authorizer.WithRequireOfflineAccess(true),
		
		// will automatically open the browser if needed,
		// prefilled with the device code
		authorizer.WithAutoOpenBrowser(true),
		authorizer.WithPrefillDeviceCode(true),

		// will cache the token on filesystem
		// and refresh it automatically 
		// if it's expired or expires in less than 5 minutes
		authorizer.WithAppDataStore(5*time.Minute), 
	)

	authorization, err := auth.Authorize(context.TODO())
	if err != nil {
		panic(err)
	}
	// use authorization.Tokens.AccessToken from now on

	pretty, _ := json.MarshalIndent(authorization, "", "  ")
	logger.Info(string(pretty))
	
	// ...
	// if you store the refresh token and you want
	// to fetch a new access token:
	refreshed, _ := auth.Refresh(context.TODO(), authorization.Tokens.RefreshToken)

	pretty, _ = json.MarshalIndent(refreshed, "", "  ")
	logger.Info(string(pretty))
}
```
