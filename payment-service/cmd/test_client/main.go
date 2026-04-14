// cmd/test_client/main.go
// Run this to manually test your Payment Service without a frontend
// Usage: go run cmd/test_client/main.go

package main

import (
	"context"
	"log"
	"time"

	pb "github.com/sumitDon47/payment-system/payment-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to the gRPC server
	conn, err := grpc.Dial(
		"localhost:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewPaymentServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ── Test: Get balance ────────────────────────────────────────────────────
	log.Println("Testing GetBalance...")
	balResp, err := client.GetBalance(ctx, &pb.GetBalanceRequest{
		UserID: "replace-with-real-user-uuid",
	})
	if err != nil {
		log.Printf("GetBalance error: %v", err)
	} else {
		log.Printf("Balance: %.2f %s", balResp.Balance, balResp.Currency)
	}

	// ── Test: Send payment ───────────────────────────────────────────────────
	log.Println("Testing SendPayment...")
	payResp, err := client.SendPayment(ctx, &pb.SendPaymentRequest{
		SenderID:   "replace-with-sender-uuid",
		ReceiverID: "replace-with-receiver-uuid",
		Amount:     500.00,
		Currency:   "NPR",
		Note:       "Test payment from gRPC client",
	})
	if err != nil {
		log.Printf("SendPayment error: %v", err)
	} else {
		log.Printf("Payment sent! TxID: %s | New balance: %.2f", payResp.TransactionID, payResp.SenderBalance)
	}

	// ── Test: Get transaction ────────────────────────────────────────────────
	if payResp != nil {
		log.Println("Testing GetTransaction...")
		txResp, err := client.GetTransaction(ctx, &pb.GetTransactionRequest{
			TransactionID: payResp.TransactionID,
		})
		if err != nil {
			log.Printf("GetTransaction error: %v", err)
		} else {
			log.Printf("Transaction: %s | %.2f %s | %s", txResp.TransactionID, txResp.Amount, txResp.Currency, txResp.Status)
		}
	}
}
