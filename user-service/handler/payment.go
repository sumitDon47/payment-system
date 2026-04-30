package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	pb "github.com/sumitDon47/payment-system/payment-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TransferRequest struct {
	ReceiverID string  `json:"receiver_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Note       string  `json:"note"`
}

func Transfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// 1. Get Sender ID from JWT Token (Set by AuthMiddleware in Header)
	senderID := r.Header.Get("X-User-ID")
	if senderID == "" {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// 2. Parse the request body
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.ReceiverID == "" || req.Amount <= 0 {
		http.Error(w, `{"error": "Invalid receiver or amount"}`, http.StatusBadRequest)
		return
	}
	if req.Currency == "" {
		req.Currency = "NPR" // default to NPR
	}

	// 3. Connect to the Payment gRPC Service
	paymentServiceURL := os.Getenv("PAYMENT_SERVICE_URL")
	if paymentServiceURL == "" {
		paymentServiceURL = "localhost:9090"
	}

	conn, err := grpc.NewClient(paymentServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to payment service: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := pb.NewPaymentServiceClient(conn)

	// 4. Create the gRPC request
	grpcReq := &pb.SendPaymentRequest{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Note:       req.Note,
	}

	// 5. Call the Payment Service
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.SendPayment(ctx, grpcReq)
	if err != nil {
		log.Printf("gRPC Call failed: %v", err)
		http.Error(w, `{"error": "Payment failed"}`, http.StatusInternalServerError)
		return
	}

	// 6. Invalidate cache for both sender and receiver to ensure fresh balances
	if Cache != nil {
		Cache.InvalidateMultiple(r.Context(), senderID, req.ReceiverID)
		log.Printf("💾 Cache invalidated for sender=%s and receiver=%s after transfer", senderID, req.ReceiverID)
	}

	// 7. Return the response to the mobile app
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
