package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "net/url"
    "os"
    "sync"

    "github.com/alecthomas/kong"
    goshopify "github.com/bold-commerce/go-shopify/v3"
)

var version string = "v0.0.0-unset"

type config struct {
    Address string  `help:"" default:":8080"`
    BaseURL url.URL `help:"base URL used in redirect"`
}

type server struct {
    locker      *sync.RWMutex
    apps        map[string]goshopify.App
    mux         *http.ServeMux
    redirectURL *url.URL
}

type App struct {
    Name         string
    ClientID     string
    ClientSecret string
    Scope        string
}

func newServer(cfg config) *server {
    redirectPath := "/redirect"
    s := &server{
        locker:      &sync.RWMutex{},
        apps:        map[string]goshopify.App{},
        mux:         http.NewServeMux(),
        redirectURL: cfg.BaseURL.JoinPath(redirectPath),
    }
    s.mux.HandleFunc("/auth", s.handleAuth)
    s.mux.HandleFunc(redirectPath, s.handleRedirect)
    s.mux.HandleFunc("/new", s.handleNew)
    s.mux.HandleFunc("/version", s.version)
    s.mux.HandleFunc("/", s.handleNoMatch)

    return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    s.mux.ServeHTTP(w, r)
}

func (s *server) setApp(appName string, app goshopify.App) {
    s.locker.Lock()
    defer s.locker.Unlock()
    s.apps[appName] = app
}

func (s *server) app(appName string) (goshopify.App, bool) {
    s.locker.RLock()
    defer s.locker.RUnlock()
    app, ok := s.apps[appName]
    return app, ok
}

func (s *server) makeRedirectURL(appName string) string {
    u := *s.redirectURL
    q := u.Query()
    q.Set("app", appName)
    u.RawQuery = q.Encode()
    return u.String()
}

// handleAuth redirects to an oauth-authorize url for the app.
func (s *server) handleAuth(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(os.Stdout, "auth: url: %s method: %s\n", r.URL.String(), r.Method)
    if r.Method != http.MethodGet {
        http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
        return
    }
    appName := r.URL.Query().Get("app")
    app, ok := s.app(appName)
    if !ok {
        msg := fmt.Sprintf("app %q not found. create new app at '/new'", appName)
        http.Error(w, msg, http.StatusBadRequest)
        return
    }
    shopName := r.URL.Query().Get("shop")
    state := "nonce"
    authUrl := app.AuthorizeUrl(shopName, state)
    fmt.Println()
    http.Redirect(w, r, authUrl, http.StatusFound)
}

// handleRedirect fetches a permanent access token in the callback
func (s *server) handleRedirect(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(os.Stdout, "redirect: url: %s method: %s\n", r.URL.String(), r.Method)
    if r.Method != http.MethodGet {
        http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
        return
    }
    appName := r.URL.Query().Get("app")
    app, ok := s.app(appName)
    if !ok {
        msg := fmt.Sprintf("app %q not found. create new app at '/new'", appName)
        http.Error(w, msg, http.StatusBadRequest)
        return
    }
    if ok, _ := app.VerifyAuthorizationURL(r.URL); !ok {
        http.Error(w, "Invalid Signature", http.StatusUnauthorized)
        return
    }

    query := r.URL.Query()
    shopName := query.Get("shop")
    code := query.Get("code")
    token, err := app.GetAccessToken(shopName, code)
    if err != nil {
        log.Println("cannot get access token", err)
        return
    }

    fmt.Println("token", token)
    fmt.Fprintf(w, "%q connected successfully.\n", appName)
    // Do something with the token, like store it in a DB.
}

// handleRedirect fetches a permanent access token in the callback
func (s *server) handleNew(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(os.Stdout, "new: url: %s method: %s\n", r.URL.String(), r.Method)
    if r.Method != http.MethodPost {
        http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
        return
    }
    app := App{}
    err := json.NewDecoder(r.Body).Decode(&app)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if app.Name == "" || app.ClientID == "" || app.ClientSecret == "" || app.Scope == "" {
        msg := `"name", "clientID", "clientSecret" and "scope" cannot be empty`
        http.Error(w, msg, http.StatusBadRequest)
        return
    }
    shopifyApp := goshopify.App{
        ApiKey:      app.ClientID,
        ApiSecret:   app.ClientSecret,
        Scope:       app.Scope,
        RedirectUrl: s.makeRedirectURL(app.Name),
    }
    s.setApp(app.Name, shopifyApp)
    fmt.Println("successfully added", app.Name)
    fmt.Fprintln(w, "successfully added", app.Name)
}

func (s *server) version(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "version", version)
}

func (s *server) handleNoMatch(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(os.Stdout, "no-math: url: %s method: %s\n", r.URL.String(), r.Method)
    http.NotFound(w, r)
}

func main() {
    opts := []kong.Option{
        kong.Description("Basic Shopify Oauth demo server"),
        kong.DefaultEnvars("shopify"),
    }
    cfg := config{}
    _ = kong.Parse(&cfg, opts...)
    fmt.Printf("%#v\n", cfg)
    server := newServer(cfg)
    fmt.Println("starting server on http://localhost:8080")
    if err := http.ListenAndServe(cfg.Address, server); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
