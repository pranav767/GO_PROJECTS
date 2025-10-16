package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func url(path string) string { return "http://localhost" + path }

func waitFor(url string) {
	client := &http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 30; i++ {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode < 500 {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func TestFlow(t *testing.T) {
	// Ensure services are up
	waitFor(url(":80/products/health"))
	waitFor(url(":80/users/health"))

	// Signup
	signup := map[string]string{"Username": "e2euser", "Password": "pass"}
	b, _ := json.Marshal(signup)
	resp, err := http.Post(url(":80/users/signup"), "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	// Login
	resp, err = http.Post(url(":80/users/login"), "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	var lr map[string]string
	json.NewDecoder(resp.Body).Decode(&lr)
	resp.Body.Close()
	token, ok := lr["token"]
	if !ok {
		t.Fatalf("no token returned")
	}

	// Add to cart
	client := &http.Client{}
	add := map[string]interface{}{"product_id": 1, "quantity": 1}
	badd, _ := json.Marshal(add)
	req, _ := http.NewRequest("POST", url(":80/cart/add"), bytes.NewReader(badd))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("add to cart failed: %v", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	// Create order
	req, _ = http.NewRequest("POST", url(":80/orders"), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("create order failed: %v", err)
	}
	var or map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&or)
	resp.Body.Close()
	if _, ok := or["order_id"]; !ok {
		t.Fatalf("order not created")
	}

	// Create payment intent
	oid := int(or["order_id"].(float64))
	ci := map[string]int{"order_id": oid}
	bci, _ := json.Marshal(ci)
	resp, err = http.Post(url(":80/payments/create_intent"), "application/json", bytes.NewReader(bci))
	if err != nil {
		t.Fatalf("create intent failed: %v", err)
	}
	var pi map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&pi)
	resp.Body.Close()
	pid, ok := pi["payment_intent_id"].(string)
	if !ok || pid == "" {
		t.Fatalf("no payment intent id")
	}

	// Simulate webhook
	ev := map[string]interface{}{"type": "payment_intent.succeeded", "data": map[string]string{"id": pid}}
	bev, _ := json.Marshal(ev)
	resp, err = http.Post(url(":80/payments/webhook"), "application/json", bytes.NewReader(bev))
	if err != nil {
		t.Fatalf("webhook failed: %v", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	// Success
	if os.Getenv("CI") == "true" {
		t.Log("CI run")
	}
}
