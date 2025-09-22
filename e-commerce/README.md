
# E-Commerce Go API

This project is a simple e-commerce backend written in Go, using Gin, MySQL, JWT authentication, and Stripe for payments.

## Features
- User registration and login (JWT-based)
- Add/remove products to/from cart
- Checkout with Stripe PaymentIntent
- Stripe webhook for payment confirmation
- Order and inventory management

---


## Project Structure

```
e-commerce/
├── cmd/
│   └── main.go                # Entry point, initializes DB and server
├── internal/
│   ├── controller/            # HTTP handlers (controllers)
│   │   └── controller.go
│   ├── service/               # Business logic (services)
│   │   ├── auth.go
│   │   └── cart.go
│   ├── db/                    # Database connection and queries
│   │   ├── db.go
│   │   └── product.go
│   ├── routes/                # Route registration
│   │   └── route.go
│   └── middleware/            # JWT middleware
│       └── jwt.go
├── model/
│   └── model.go               # Data models (User, Product, Cart, etc.)
├── utils/
│   └── utils.go               # Utility functions (hashing, JWT, etc.)
├── sample_data.sql            # Sample SQL for products and cart
├── README.md                  # Project documentation
├── go.mod / go.sum            # Go modules
└── config.env                 # Environment variables (if used)
```


1. **Clone the repository**
2. **Configure environment variables**
   - Copy `config.env` or `.env.example` to `.env` and set:
     - `STRIPE_SECRET_KEY`
     - `STRIPE_WEBHOOK_SECRET` (from Stripe CLI)
     - `HMAC_SECRET` (for JWT)

3. **Start MySQL** (Docker example):
   ```bash
   docker run --name mysql-ecommerce -e MYSQL_ROOT_PASSWORD=adminpass -e MYSQL_DATABASE=e-commerce -p 3306:3306 -d mysql:8
   ```

4. **Create the orders table** (if not already):
   ```sql
   CREATE TABLE orders (
     id BIGINT AUTO_INCREMENT PRIMARY KEY,
     user_id BIGINT NOT NULL,
     amount BIGINT NOT NULL,
     currency VARCHAR(10) NOT NULL,
     status VARCHAR(20) NOT NULL,
     stripe_intent VARCHAR(64) NOT NULL,
     created_at DATETIME NOT NULL
   );
   ```

5. **Install Go dependencies**
   ```bash
   go mod tidy
   ```

6. **Run the server**
   ```bash
   go run cmd/main.go
   ```

---

## API Flow

### 1. Register & Login
```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser", "password":"testpass"}'

curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser", "password":"testpass"}'
```
- Save the `token` from the login response.

### 2. Add to Cart (optional)
```bash
curl -X POST http://localhost:8080/cart/add \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{"product_id": 1, "quantity": 2}'
```

### 3. Checkout (creates PaymentIntent & order)
```bash
curl -X POST http://localhost:8080/checkout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{"amount": 1000, "currency": "usd"}'
```
- Response:
  ```json
  {
    "client_secret": "pi_XXX_secret_YYY",
    "payment_intent_id": "pi_XXX"
  }
  ```
- The order is created in the DB with status `pending` and the PaymentIntent ID.

### 4. Start Stripe CLI Webhook Listener
```bash
stripe listen --forward-to localhost:8080/webhook
```
- Copy the webhook secret (`whsec_...`) to your `.env` as `STRIPE_WEBHOOK_SECRET`.

### 5. Simulate Payment (for testing)
> **Important:** The command below creates a new PaymentIntent unrelated to your checkout/order. To test the real flow, complete the payment for the PaymentIntent returned by `/checkout` using the Stripe Dashboard or your client integration.
```bash
stripe trigger payment_intent.succeeded
```
- The webhook will update the order status to `paid` in the DB **only if the PaymentIntent ID matches an order**.
- If you see a 404 in your webhook logs, it means the PaymentIntent is not linked to any order (see troubleshooting below).

#### Troubleshooting
- If you get a 404 from `/webhook` for `payment_intent.succeeded`, make sure you are triggering the event for the PaymentIntent ID returned by `/checkout`.
- You can manually complete the payment in the Stripe Dashboard for the correct PaymentIntent, or use your client integration to pay.

### 6. Check Order Status
- Query your `orders` table to verify the order status is now `paid`.

---

## Local Stripe Payment Flow: Step-by-Step Example

1. **Start the server**
   ```bash
   go run cmd/main.go
   ```

2. **(Optional) Setup DB and create a user**
   - Register and login to get a JWT:
     ```bash
     curl -X POST http://localhost:8080/signup \
       -H "Content-Type: application/json" \
       -d '{"username":"testuser", "password":"testpass"}'
     curl -X POST http://localhost:8080/login \
       -H "Content-Type: application/json" \
       -d '{"username":"testuser", "password":"testpass"}'
     ```
   - Save the JWT token from the login response.

3. **Add to cart**
   ```bash
   curl -X POST http://localhost:8080/cart/add \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <JWT_TOKEN>" \
     -d '{"product_id": 1, "quantity": 2}'
   ```

4. **Start checkout**
   ```bash
   curl -X POST http://localhost:8080/checkout \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <JWT_TOKEN>" \
     -d '{}'
   ```
   - The response will include:
     ```json
     {"amount":199998,"client_secret":"pi_..._secret_...","currency":"usd","payment_intent_id":"pi_..."}
     ```
   - Save the `payment_intent_id` (e.g., `pi_3SA3QbJt5o8XzfsV0mRHqLHT`).

5. **Start Stripe CLI webhook listener**
   ```bash
   stripe listen --forward-to localhost:8080/webhook
   ```
   - Copy the webhook secret (`whsec_...`) to your `.env` as `STRIPE_WEBHOOK_SECRET` if not already set.

6. **Confirm the PaymentIntent using Stripe CLI**
   ```bash
   stripe payment_intents confirm <payment_intent_id> --payment-method pm_card_visa
   ```
   - Example:
     ```bash
     stripe payment_intents confirm pi_3SA3QbJt5o8XzfsV0mRHqLHT --payment-method pm_card_visa
     ```
   - This will mark the PaymentIntent as paid and Stripe will send a webhook to your local server.

7. **Check server logs and DB**
   - You should see logs like:
     ```
     [Webhook] Received event type: payment_intent.succeeded
     [Webhook] Successful payment: pi_3SA3QbJt5o8XzfsV0mRHqLHT
     [Webhook] Order <id> marked as paid and inventory updated.
     ```
   - Check your `orders` table:
     ```sql
     SELECT * FROM orders WHERE stripe_intent = '<payment_intent_id>';
     ```
   - The order status should now be `paid`.

---

### Troubleshooting
- If you get a 404 from `/webhook` for `payment_intent.succeeded`, make sure you are confirming the PaymentIntent ID returned by `/checkout`.
- If you do not see the webhook, check your Stripe Dashboard → Developers → Webhooks → Click your endpoint → See if the event was sent and if there are any delivery errors.
- Always use a new `/checkout` call for each payment attempt.

---

## Notes
- Always use a new `/checkout` call for each payment attempt.
- The webhook handler is robust to Stripe API version mismatches (for local testing).
- For production, remove `IgnoreAPIVersionMismatch: true` and keep your Stripe Go SDK up to date.

---

## Next Steps
- Add input validation, error handling, and unit tests.
- Move models and DB logic to `internal/model` for better separation.
- Add database migrations (e.g., with golang-migrate).

---
