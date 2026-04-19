//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	pb "github.com/sumitDon47/payment-system/payment-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type registerResponse struct {
	Token string `json:"token"`
	User  struct {
		ID string `json:"id"`
	} `json:"user"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type walletResponse struct {
	Data map[string]float64 `json:"data"`
}

func TestE2EPaymentFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	userServiceURL := envOrDefault("USER_SERVICE_URL", "http://localhost:8080")
	paymentServiceAddr := envOrDefault("PAYMENT_SERVICE_ADDR", "localhost:9090")

	waitForUserService(t, userServiceURL)

	testSuffix := fmt.Sprintf("%d", time.Now().UnixNano())
	senderEmail := "sender+" + testSuffix + "@test.com"
	receiverEmail := "receiver+" + testSuffix + "@test.com"
	password := "SecurePass123!"

	sender := registerUser(t, userServiceURL, "Sender", senderEmail, password)
	receiver := registerUser(t, userServiceURL, "Receiver", receiverEmail, password)

	if sender.Token == "" {
		t.Fatal("expected register response token for sender")
	}
	if receiver.Token == "" {
		t.Fatal("expected register response token for receiver")
	}

	loginToken := loginUser(t, userServiceURL, senderEmail, password)
	if loginToken == "" {
		t.Fatal("expected non-empty login token")
	}

	dbConn := openDB(t)
	defer dbConn.Close()

	setBalance(t, dbConn, sender.User.ID, 1000)
	setBalance(t, dbConn, receiver.User.ID, 0)

	grpcConn, err := grpc.Dial(paymentServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect payment service: %v", err)
	}
	defer grpcConn.Close()

	client := pb.NewPaymentServiceClient(grpcConn)

	sendResp, err := client.SendPayment(ctx, &pb.SendPaymentRequest{
		SenderID:   sender.User.ID,
		ReceiverID: receiver.User.ID,
		Amount:     250,
		Currency:   "NPR",
		Note:       "integration-e2e",
	})
	if err != nil {
		t.Fatalf("SendPayment failed: %v", err)
	}
	if sendResp.TransactionID == "" {
		t.Fatal("expected transaction_id in SendPayment response")
	}
	if sendResp.Status != "completed" {
		t.Fatalf("expected completed status, got %q", sendResp.Status)
	}

	txResp, err := client.GetTransaction(ctx, &pb.GetTransactionRequest{TransactionID: sendResp.TransactionID})
	if err != nil {
		t.Fatalf("GetTransaction failed: %v", err)
	}
	if txResp.SenderID != sender.User.ID || txResp.ReceiverID != receiver.User.ID {
		t.Fatalf("unexpected transaction participants: sender=%s receiver=%s", txResp.SenderID, txResp.ReceiverID)
	}
	if txResp.Amount != 250 {
		t.Fatalf("expected transaction amount 250, got %.2f", txResp.Amount)
	}

	senderBal, err := client.GetBalance(ctx, &pb.GetBalanceRequest{UserID: sender.User.ID})
	if err != nil {
		t.Fatalf("GetBalance sender failed: %v", err)
	}
	receiverBal, err := client.GetBalance(ctx, &pb.GetBalanceRequest{UserID: receiver.User.ID})
	if err != nil {
		t.Fatalf("GetBalance receiver failed: %v", err)
	}
	if senderBal.Balance != 750 {
		t.Fatalf("expected sender balance 750, got %.2f", senderBal.Balance)
	}
	if receiverBal.Balance != 250 {
		t.Fatalf("expected receiver balance 250, got %.2f", receiverBal.Balance)
	}

	walletBal := getWalletBalance(t, userServiceURL, loginToken)
	if walletBal != 750 {
		t.Fatalf("expected wallet balance 750, got %.2f", walletBal)
	}
}

func TestE2EInsufficientFunds(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userServiceURL := envOrDefault("USER_SERVICE_URL", "http://localhost:8080")
	paymentServiceAddr := envOrDefault("PAYMENT_SERVICE_ADDR", "localhost:9090")

	waitForUserService(t, userServiceURL)

	testSuffix := fmt.Sprintf("%d", time.Now().UnixNano())
	senderEmail := "poor-sender+" + testSuffix + "@test.com"
	receiverEmail := "poor-receiver+" + testSuffix + "@test.com"
	password := "SecurePass123!"

	sender := registerUser(t, userServiceURL, "Poor Sender", senderEmail, password)
	receiver := registerUser(t, userServiceURL, "Poor Receiver", receiverEmail, password)

	dbConn := openDB(t)
	defer dbConn.Close()

	setBalance(t, dbConn, sender.User.ID, 10)
	setBalance(t, dbConn, receiver.User.ID, 0)

	grpcConn, err := grpc.Dial(paymentServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect payment service: %v", err)
	}
	defer grpcConn.Close()

	client := pb.NewPaymentServiceClient(grpcConn)

	_, err = client.SendPayment(ctx, &pb.SendPaymentRequest{
		SenderID:   sender.User.ID,
		ReceiverID: receiver.User.ID,
		Amount:     100,
		Currency:   "NPR",
		Note:       "integration-insufficient-funds",
	})
	if err == nil {
		t.Fatal("expected insufficient funds error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "insufficient funds") {
		t.Fatalf("expected insufficient funds error, got: %v", err)
	}
}

func waitForUserService(t *testing.T, userServiceURL string) {
	t.Helper()

	client := &http.Client{Timeout: 3 * time.Second}
	deadline := time.Now().Add(45 * time.Second)

	for time.Now().Before(deadline) {
		req, err := http.NewRequest(http.MethodGet, userServiceURL+"/health", nil)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return
				}
			}
		}
		time.Sleep(2 * time.Second)
	}

	t.Fatalf("user service did not become healthy at %s/health", userServiceURL)
}

func registerUser(t *testing.T, baseURL, name, email, password string) registerResponse {
	t.Helper()

	payload := map[string]string{
		"name":     name,
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(payload)

	resp, respBody := doJSONRequest(t, http.MethodPost, baseURL+"/register", body, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var out registerResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		t.Fatalf("failed to parse register response: %v", err)
	}
	if out.User.ID == "" {
		t.Fatalf("register response missing user id: %s", string(respBody))
	}
	return out
}

func loginUser(t *testing.T, baseURL, email, password string) string {
	t.Helper()

	payload := map[string]string{"email": email, "password": password}
	body, _ := json.Marshal(payload)

	resp, respBody := doJSONRequest(t, http.MethodPost, baseURL+"/login", body, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var out loginResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}
	return out.Token
}

func getWalletBalance(t *testing.T, baseURL, token string) float64 {
	t.Helper()

	resp, respBody := doJSONRequest(t, http.MethodGet, baseURL+"/wallet", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("wallet request failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var out walletResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		t.Fatalf("failed to parse wallet response: %v", err)
	}
	balance, ok := out.Data["balance"]
	if !ok {
		t.Fatalf("wallet response missing balance field: %s", string(respBody))
	}
	return balance
}

func doJSONRequest(t *testing.T, method, url string, body []byte, token string) (*http.Response, []byte) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("failed to create request %s %s: %v", method, url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed %s %s: %v", method, url, err)
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	// Return a shallow copy with a no-op body to avoid accidental reuse issues.
	respCopy := *resp
	respCopy.Body = io.NopCloser(bytes.NewReader(nil))
	return &respCopy, respBody
}

func openDB(t *testing.T) *sql.DB {
	t.Helper()

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		envOrDefault("DB_HOST", "localhost"),
		envOrDefault("DB_PORT", "5432"),
		envOrDefault("DB_USER", "postgres"),
		envOrDefault("DB_PASSWORD", "yourpassword"),
		envOrDefault("DB_NAME", "payment_db"),
	)

	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed opening db: %v", err)
	}
	if err := dbConn.Ping(); err != nil {
		t.Fatalf("failed pinging db: %v", err)
	}
	return dbConn
}

func setBalance(t *testing.T, dbConn *sql.DB, userID string, amount float64) {
	t.Helper()

	_, err := dbConn.Exec(`UPDATE users SET balance = $1, updated_at = NOW() WHERE id = $2`, amount, userID)
	if err != nil {
		t.Fatalf("failed to set user balance: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
