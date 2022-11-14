# Shopify OAuth demo

This project is a most basic shopify OAuth demo with a docker image on
[julia/shopify](https://hub.docker.com/repository/docker/julia/shopify).

**Deploy** this app to your own domain, e.g. yourdomain.com. 

Setup a **Custom Shopify App**  in the [shopify patner](https://www.shopify.com/partners) settings.
For Custom Apps, there has to be a 1:1 mapping between custom app and a given store:

1. Click `Create App`
2. Create app from scratch, enter app name, e.g. `yourapp-target-store`.

Wire the correct **OAuth URLs** for this demo server:

1. Navigate to `App Setup` on left side panel
2. Under `App URL` enter `<your-base-url>/auth?app=<yourapp-name>`.  
   E.g. `https://yourdomain.com/auth?app=yourapp-target-store`.
3. Under `Allowed redirection URL(s)` enter `<your-base-url>/redirect?app=<yourapp-name>`  
   E.g. `https://yourdomain.com/redirect?app=yourapp-target-store`.


Copy `ClientID` and `ClientSecret` and register with server:

1. Navigate to `Overview` on left side panel.
2. Copy `ClientID` and `ClientSecret` and create and HTTP POST request to
 `<your-base-url>/new` with a JSON payload encoding the app's `name`, `clientID`, `clientSecret` and desired `scope`, e.g.

```
curl https://yourdomain.com/new -d '{
  "name": "yourapp-target-store",
  "clientID": "34ad8fcafecafecafecafecafecafe679",
  "clientSecret": "78b29ecafecafecafecafecafecafe56bf0",
  "scope": "read_orders,write_orders,read_products,write_products"
}'
```

Last create a **Single Merchant Install Link**

1. Navigate to `Distribution` on left side panel
2. Click `Choose Distribution`
3. Click `Choose single-merchant install link`
4. Enter shop URL (https://<TARGET-SHOP>.myshopify.com)
5. Copy Install URL

When a person with admin access to `TARGET-SHOP` clicks the Install URL
this shopify server should print your permanent access token in the
logs or other errors / problems.

All going well the clicked installation link will eventually respond with

	"yourapp-target-store" connected successfully.
