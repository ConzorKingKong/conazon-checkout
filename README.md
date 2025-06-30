# Conazon Checkout

This is the checkout endpoint for the Conazon project.

## Quickstart

To test locally, setup a `.env` file in the root directory with the following variables:

```
JWTSECRET - JWT secret. Must match the secret used in the auth service REQUIRED
DATABASEURL - Url to postgres database. REQUIRED
RABBITMQURL - Url that RabbitMQ instance is running. Defaults to a rabbitmq container inside of docker-compose
PORT - Port to run server on. Defaults to 8083
```

Datbase url should be formatted like this if using `docker-compose up` - `'host=postgres port=5432 user=postgres dbname=conazon sslmode=disable'`

Then run:

`docker-compose up`

## Endpoints (later will have swagger)

- /

GET - Catch all 404

- /checkout

POST - Checks user out/kicks off rabbitmq queue

{userId, productArray}

- /checkout/{id}

GET - Returns cart id entry

- /checkout/user/{id}

GET - get all active items in users cart
