package main

import (
    "fmt"
    "log"
    "net/http"
    "net/http/httputil"
    "os"
    "path"

    "github.com/alecthomas/kong"
    goshopify "github.com/bold-commerce/go-shopify/v3"
)

var version string = "v0.0.0-unset"

type config struct {
    ClientID     string `help:"shopify app API key"`
    ClientSecret string `help:"shopify app API secret"`
    Scope        string `help:"shopify app scope" default:"read_products,write_products,read_orders,write_products"`
    BaseURL      string `help:"shopify app OAuth redirect base URL (no path)"`
}

type server struct {
    app goshopify.App
    mux *http.ServeMux
}

func newServer(cfg config) *server {
    oauthCallbackPath := "/oauth_callback"
    installShopifyAppPath := "/install_shopify_app"

    s := &server{
        app: goshopify.App{
            ApiKey:      cfg.ClientID,
            ApiSecret:   cfg.ClientSecret,
            RedirectUrl: path.Join(cfg.BaseURL, oauthCallbackPath),
            Scope:       cfg.Scope,
        },
        mux: http.NewServeMux(),
    }
    s.mux.HandleFunc(installShopifyAppPath, s.installShopifyApp)
    s.mux.HandleFunc(oauthCallbackPath, s.oauthCallback)
    s.mux.HandleFunc("/version", s.version)
    s.mux.HandleFunc("/", s.logAll)

    return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    s.mux.ServeHTTP(w, r)
}

// installShopifyApp redirects to an oauth-authorize url for the app.
func (s *server) installShopifyApp(w http.ResponseWriter, r *http.Request) {
    shopName := r.URL.Query().Get("shop")
    state := "nonce"
    authUrl := s.app.AuthorizeUrl(shopName, state)
    http.Redirect(w, r, authUrl, http.StatusFound)
}

// oauthCallback fetches a permanent access token in the callback
func (s *server) oauthCallback(w http.ResponseWriter, r *http.Request) {
    if ok, _ := s.app.VerifyAuthorizationURL(r.URL); !ok {
        http.Error(w, "Invalid Signature", http.StatusUnauthorized)
        return
    }

    query := r.URL.Query()
    shopName := query.Get("shop")
    code := query.Get("code")
    token, err := s.app.GetAccessToken(shopName, code)
    if err != nil {
        log.Println("cannot get access token", err)
        return
    }

    fmt.Println("token", token)
    // Do something with the token, like store it in a DB.
}

func (s *server) version(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "version", version)
}

func (s *server) logAll(w http.ResponseWriter, r *http.Request) {
    fmt.Println("logAll:")
    b, _ := httputil.DumpRequest(r, true)
    fmt.Fprintf(os.Stdout, "%s\n", b)
    http.NotFound(w, r)
}

func main() {
    opts := []kong.Option{
        kong.Description("Basic Shopify Oauth demo server"),
        kong.DefaultEnvars("shopify"),
    }
    cfg := config{}
    _ = kong.Parse(&cfg, opts...)
    server := newServer(cfg)
    fmt.Println("starting server on http://localhost:8080")
    if err := http.ListenAndServe(":8080", server); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
