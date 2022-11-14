# Shopify OAuth demo

This project is a most basic shopify OAuaht demo with a docker image on
[julia/shopify](https://hub.docker.com/repository/docker/julia/shopify).

Deploy this app to your own domain, e.g. yourdomain.com. Setup a custom
app in the[shopify patner](https://www.shopify.com/partners) settings.
There has to be 1:1 mapping between custom app and a given store:

1. Click `Create App`
2. Create app from scratch, enter app name.
3. Navigate to `App Setup` on left side panel
4. Under `App URL` enter the URL where you deployed this go web server, e.g. `https://yourdomain.com/install_shopify_app`.
5. Under `Allowed redirection URL(s)` enter the URL from above with `api/auth` appended, e.g. `https://example.com/oauth_callback`

Then create "Single Merchant Install Link"

1. Navigate to `Distribution` on left side panel
2. Click `Choose Distribution`
3. Click `Choose single-merchant install link`
4. Enter shop URL (https://<TARGET-SHOP>.myshopify.com)
5. Copy Install URL

When a person with admin access to `TARGET-SHOP` clicks the Install URL
this shopify server should print your permanent access token in the
logs.
